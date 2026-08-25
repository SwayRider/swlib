package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/swayrider/swlib/http/cookies"
	log "github.com/swayrider/swlib/logger"
	"github.com/swayrider/swlib/security"
)

func noKeys() ([]string, error) { return nil, nil }

// garbageCookie returns an access_token cookie whose value is not valid
// base64 -- the exact condition that used to hard-401 every /web/* request
// regardless of whether the endpoint was public.
func garbageCookie() *http.Cookie {
	return &http.Cookie{
		Name:  cookies.FullCookieName("access_token"),
		Value: "not-valid-base64!!!",
	}
}

func TestAuth_UndecodableCookie_PublicEndpoint_Allowed(t *testing.T) {
	security.PublicEndpoint("/auth-test/public-undecodable")

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	handler := Auth(next, noKeys, log.New())

	req := httptest.NewRequest(http.MethodGet, "/auth-test/public-undecodable", nil)
	req.AddCookie(garbageCookie())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !called {
		t.Error("expected the wrapped handler to run for a public endpoint")
	}
}

func TestAuth_UndecodableCookie_ProtectedEndpoint_Rejected(t *testing.T) {
	// Deliberately not registered via security.PublicEndpoint, so it falls
	// back to the default (auth-required) profile.
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	handler := Auth(next, noKeys, log.New())

	req := httptest.NewRequest(http.MethodGet, "/auth-test/protected-undecodable", nil)
	req.AddCookie(garbageCookie())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if called {
		t.Error("the wrapped handler must not run for a protected endpoint without a valid token")
	}
}

func TestAuth_NoCookie_PublicEndpoint_Allowed(t *testing.T) {
	security.PublicEndpoint("/auth-test/public-no-cookie")

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	handler := Auth(next, noKeys, log.New())

	req := httptest.NewRequest(http.MethodGet, "/auth-test/public-no-cookie", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !called {
		t.Error("expected the wrapped handler to run for a public endpoint")
	}
}
