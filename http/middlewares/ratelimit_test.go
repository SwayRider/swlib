package middlewares

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	log "github.com/swayrider/swlib/logger"
	"github.com/swayrider/swlib/ratelimit"
)

func init() {
	log.SetOutput(io.Discard)
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func doRequest(h http.Handler, remoteAddr string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remoteAddr
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	limiter := ratelimit.New(100, 5, time.Minute)
	h := RateLimit(okHandler(), limiter, log.New())

	rec := doRequest(h, "203.0.113.1:1234", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRateLimit_BlocksOverLimit_Returns429(t *testing.T) {
	limiter := ratelimit.New(1, 1, time.Minute)
	h := RateLimit(okHandler(), limiter, log.New())

	if rec := doRequest(h, "203.0.113.1:1234", nil); rec.Code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", rec.Code)
	}
	rec := doRequest(h, "203.0.113.1:5678", nil)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected 429, got %d", rec.Code)
	}
}

func TestRateLimit_UsesRemoteAddrNotForgedXFF(t *testing.T) {
	limiter := ratelimit.New(1, 1, time.Minute)
	h := RateLimit(okHandler(), limiter, log.New())

	// Real peer is the same on both requests, but the caller claims a
	// different address via a spoofable header each time. The bucket must
	// be keyed off RemoteAddr, not the forged claim.
	if rec := doRequest(h, "203.0.113.1:1234", map[string]string{"X-Forwarded-For": "9.9.9.9"}); rec.Code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", rec.Code)
	}
	rec := doRequest(h, "203.0.113.1:5678", map[string]string{"X-Forwarded-For": "8.8.8.8"})
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected the second request from the same real peer to be blocked despite a different forged XFF, got %d", rec.Code)
	}
}

func TestRateLimit_ExemptsLoopback(t *testing.T) {
	limiter := ratelimit.New(1, 1, time.Minute)
	h := RateLimit(okHandler(), limiter, log.New())

	// Loopback peers (e.g. container-internal health checks) share one
	// address across many independent callers, so they must be exempt.
	for i := range 5 {
		rec := doRequest(h, "127.0.0.1:1234", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d from loopback: expected 200, got %d", i, rec.Code)
		}
	}
}
