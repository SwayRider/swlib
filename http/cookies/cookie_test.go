package cookies

import (
	"net/http"
	"testing"
)

func TestParseSameSite(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    http.SameSite
		wantErr bool
	}{
		{"empty defaults to strict", "", http.SameSiteStrictMode, false},
		{"strict", "strict", http.SameSiteStrictMode, false},
		{"lax", "lax", http.SameSiteLaxMode, false},
		{"case insensitive strict", "Strict", http.SameSiteStrictMode, false},
		{"case insensitive lax", "LAX", http.SameSiteLaxMode, false},
		{"none rejected", "none", http.SameSiteStrictMode, true},
		{"unknown rejected", "garbage", http.SameSiteStrictMode, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSameSite(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseSameSite(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseSameSite(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestSetDefaultSameSite(t *testing.T) {
	original := defaultSameSite
	t.Cleanup(func() { defaultSameSite = original })

	SetDefaultSameSite(http.SameSiteStrictMode)

	if c := NewServerCookie("t", []byte("x")); c.SameSite != http.SameSiteStrictMode {
		t.Errorf("NewServerCookie default SameSite = %v, want %v", c.SameSite, http.SameSiteStrictMode)
	}
	if c := ClearCookie("t"); c.SameSite != http.SameSiteStrictMode {
		t.Errorf("ClearCookie default SameSite = %v, want %v", c.SameSite, http.SameSiteStrictMode)
	}

	// Explicit opts still override the configured default.
	opts := NewCookieOpts()
	opts.SetSameSite(http.SameSiteLaxMode)

	if c := NewServerCookie("t", []byte("x"), opts); c.SameSite != http.SameSiteLaxMode {
		t.Errorf("NewServerCookie explicit SameSite = %v, want %v", c.SameSite, http.SameSiteLaxMode)
	}
	if c := ClearCookie("t", opts); c.SameSite != http.SameSiteLaxMode {
		t.Errorf("ClearCookie explicit SameSite = %v, want %v", c.SameSite, http.SameSiteLaxMode)
	}
}
