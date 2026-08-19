package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMimeType_SetsContentTypeForKnownExtension(t *testing.T) {
	h := MimeType(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/styles/app.css", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/css") {
		t.Errorf("expected Content-Type to start with text/css, got %q", ct)
	}
}

func TestMimeType_NoDotInPathDoesNotPanic(t *testing.T) {
	h := MimeType(okHandler())

	for _, p := range []string{"/", "/api/users", "/health"} {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("path %q: expected 200, got %d", p, rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "" {
			t.Errorf("path %q: expected no Content-Type header, got %q", p, ct)
		}
	}
}

func TestMimeType_UnknownExtensionLeavesHeaderUnset(t *testing.T) {
	h := MimeType(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/file.unknownext123", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "" {
		t.Errorf("expected no Content-Type header, got %q", ct)
	}
}

func TestMimeType_TrailingDotDoesNotPanic(t *testing.T) {
	h := MimeType(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/file.", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "" {
		t.Errorf("expected no Content-Type header, got %q", ct)
	}
}
