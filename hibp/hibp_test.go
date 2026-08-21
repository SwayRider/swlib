package hibp

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	log "github.com/swayrider/swlib/logger"
)

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	m.Run()
}

// testClient builds a Client pointed at srv with the given settings. The test
// lives in the same package so it can override baseURL (white-box).
func testClient(srv *httptest.Server, enabled bool, minCount int) *Client {
	c := New(enabled, 3*time.Second, minCount, log.New())
	c.baseURL = srv.URL
	return c
}

// sha1Hex returns the uppercase SHA-1 hex digest of s.
func sha1Hex(s string) string {
	return fmt.Sprintf("%X", sha1.Sum([]byte(s)))
}

func TestIsBreached_OnlyPrefixLeavesTheServer(t *testing.T) {
	password := "Correct Horse Battery Staple 42!"
	hash := sha1Hex(password)

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = fmt.Fprintf(w, "%s:5\n", strings.ToUpper(hash[5:]))
	}))
	defer srv.Close()

	c := testClient(srv, true, 1)
	breached, count, err := c.IsBreached(context.Background(), password)
	if err != nil {
		t.Fatalf("IsBreached failed: %v", err)
	}

	wantPrefix := "/range/" + hash[:5]
	if gotPath != wantPrefix {
		t.Errorf("request path = %q, want %q (only the 5-char prefix may leave the server)", gotPath, wantPrefix)
	}
	if strings.Contains(gotPath, hash[5:]) {
		t.Errorf("request path leaks the full hash suffix: %q", gotPath)
	}
	if !breached {
		t.Error("expected breached = true")
	}
	if count != 5 {
		t.Errorf("count = %d, want 5", count)
	}
}

func TestIsBreached_HeadersPresent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing User-Agent header (the API rejects requests without one)")
		}
		if r.Header.Get("Add-Padding") != "true" {
			t.Errorf("Add-Padding = %q, want \"true\"", r.Header.Get("Add-Padding"))
		}
		_, _ = fmt.Fprintln(w, "ABCDEF01:3")
	}))
	defer srv.Close()

	c := testClient(srv, true, 1)
	if _, _, err := c.IsBreached(context.Background(), "some-password"); err != nil {
		t.Fatalf("IsBreached failed: %v", err)
	}
}

func TestIsBreached_SuffixMatchIsCaseInsensitive(t *testing.T) {
	password := "Lowercase Hash Comparison"
	hash := sha1Hex(password)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with the lowercase suffix even though the API returns
		// uppercase; matching must be case-insensitive.
		_, _ = fmt.Fprintf(w, "%s:7\n", strings.ToLower(hash[5:]))
	}))
	defer srv.Close()

	c := testClient(srv, true, 1)
	breached, count, err := c.IsBreached(context.Background(), password)
	if err != nil {
		t.Fatalf("IsBreached failed: %v", err)
	}
	if !breached || count != 7 {
		t.Errorf("breached = %v, count = %d, want true, 7", breached, count)
	}
}

func TestIsBreached_NoMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, "DEADBEEF:123")
	}))
	defer srv.Close()

	c := testClient(srv, true, 1)
	breached, count, err := c.IsBreached(context.Background(), "not-in-the-list")
	if err != nil {
		t.Fatalf("IsBreached failed: %v", err)
	}
	if breached {
		t.Error("expected breached = false")
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
}

func TestIsBreached_CountBelowMinCount(t *testing.T) {
	password := "Low Count Password"
	hash := sha1Hex(password)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "%s:2\n", hash[5:])
	}))
	defer srv.Close()

	c := testClient(srv, true, 5)
	breached, count, err := c.IsBreached(context.Background(), password)
	if err != nil {
		t.Fatalf("IsBreached failed: %v", err)
	}
	if breached {
		t.Error("expected breached = false when count is below minCount")
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestIsBreached_PaddingEntriesAreIgnored(t *testing.T) {
	password := "Padding Aware"
	hash := sha1Hex(password)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add-Padding appends dummy entries that may lack counts entirely.
		_, _ = fmt.Fprintln(w, "AAAAAAAAAA")
		_, _ = fmt.Fprintf(w, "%s:9\n", hash[5:])
		_, _ = fmt.Fprintln(w, "BBBBBBBBBB:0")
	}))
	defer srv.Close()

	c := testClient(srv, true, 1)
	breached, count, err := c.IsBreached(context.Background(), password)
	if err != nil {
		t.Fatalf("IsBreached failed: %v", err)
	}
	if !breached || count != 9 {
		t.Errorf("breached = %v, count = %d, want true, 9", breached, count)
	}
}

func TestIsBreached_Non200ReturnsError(t *testing.T) {
	for _, status := range []int{http.StatusForbidden, http.StatusTooManyRequests, http.StatusInternalServerError} {
		t.Run(fmt.Sprintf("status_%d", status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
			}))
			defer srv.Close()

			c := testClient(srv, true, 1)
			_, _, err := c.IsBreached(context.Background(), "some-password")
			if err == nil {
				t.Fatalf("expected error for status %d, got nil", status)
			}
		})
	}
}

func TestIsBreached_DisabledMakesNoNetworkCall(t *testing.T) {
	hit := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
	}))
	defer srv.Close()

	c := testClient(srv, false, 1)
	breached, count, err := c.IsBreached(context.Background(), "whatever")
	if err != nil {
		t.Fatalf("IsBreached failed: %v", err)
	}
	if breached || count != 0 {
		t.Errorf("breached = %v, count = %d, want false, 0", breached, count)
	}
	if hit {
		t.Error("disabled client must not make a network call")
	}
}

func TestIsBreached_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, "ABCDEF01:1")
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := testClient(srv, true, 1)
	if _, _, err := c.IsBreached(ctx, "some-password"); err == nil {
		t.Error("expected error for cancelled context")
	}
}
