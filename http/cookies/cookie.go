// Package cookies provides HTTP cookie utilities for secure cookie management.
//
// It handles base64-encoded cookie values with configurable security settings
// including secure flags, domain restrictions, and TTL.
//
// # Cookie Naming
//
// All cookies are namespaced with a configurable prefix to avoid conflicts
// with other applications. Call SetNamespace at startup to configure the
// prefix; the default is "com.hevanto-it.swayrider". Use FullCookieName
// to get the full namespaced name.
//
// # Usage
//
//	// Configure namespace at startup (call once before serving requests)
//	cookies.SetNamespace("com.example.myapp")
//
//	// Create cookie options from request context
//	opts := cookies.NewCookieOptsFromContext(ctx)
//
//	// Create a server cookie with data
//	cookie := cookies.NewServerCookie("access_token", tokenData, opts)
//	http.SetCookie(w, cookie)
//
//	// Clear a cookie
//	http.SetCookie(w, cookies.ClearCookie("access_token", opts))
//
//	// Get cookie from raw cookie header string
//	if data, ok := cookies.GetCookie(cookieHeader, cookies.FullCookieName("access_token")); ok {
//	    // use data
//	}
package cookies

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/swayrider/swlib/security"
)

// Default TTL values for cookies.
var (
	// TTLDefault is the default cookie lifetime (2 hours).
	TTLDefault = 2 * time.Hour
	// TTLRemeberLogin is the extended lifetime for "remember me" sessions (30 days).
	TTLRemeberLogin = 30 * 24 * time.Hour
)

// cookieNamespace is the prefix applied to all cookie names.
var cookieNamespace = "com.hevanto-it.swayrider"

// defaultSameSite is the SameSite policy applied to cookies when the caller
// does not explicitly set one via CookieOpts. Configure it once at service
// startup with SetDefaultSameSite; the default is Lax.
var defaultSameSite = http.SameSiteLaxMode

// SetNamespace sets the namespace prefix used for all cookie names.
// Call this once at service startup before any requests are handled.
// An empty string disables namespacing (cookies use their bare names).
func SetNamespace(ns string) {
	cookieNamespace = ns
}

// SetDefaultSameSite sets the SameSite policy applied to cookies when the
// caller does not explicitly set one via CookieOpts. Call this once at
// service startup before any requests are handled.
func SetDefaultSameSite(s http.SameSite) {
	defaultSameSite = s
}

// ParseSameSite parses a SameSite policy string into an http.SameSite value.
// The input is case-insensitive and an empty string selects the recommended
// default (Strict). "none" is deliberately unsupported: it requires the
// Secure flag, which callers cannot guarantee for every request. On any
// unrecognized value it returns Strict alongside an error so callers can log
// and still land on a valid policy.
func ParseSameSite(s string) (http.SameSite, error) {
	switch strings.ToLower(s) {
	case "":
		return http.SameSiteStrictMode, nil
	case "strict":
		return http.SameSiteStrictMode, nil
	case "lax":
		return http.SameSiteLaxMode, nil
	default:
		return http.SameSiteStrictMode, fmt.Errorf(
			"unsupported SameSite value %q (expected \"strict\" or \"lax\")", s)
	}
}

// FullCookieName returns the full namespaced cookie name.
// If the namespace is empty the bare name is returned unchanged.
func FullCookieName(name string) string {
	if cookieNamespace == "" {
		return name
	}
	return fmt.Sprintf("%s.%s", cookieNamespace, name)
}

// CookieOpts configures cookie behavior including security and lifetime settings.
type CookieOpts struct {
	secure   bool          // Whether to set the Secure flag (HTTPS only)
	domain   string        // Domain restriction for the cookie
	ttl      time.Duration // Time-to-live for the cookie
	path     string        // Path scope; overrides default "/" when non-empty
	sameSite http.SameSite // SameSite policy; overrides the configured default when non-zero
}

// NewCookieOpts creates CookieOpts with default values (not secure, no domain, default TTL).
func NewCookieOpts() CookieOpts {
	return CookieOpts{
		secure: false,
		domain: "",
		ttl:    TTLDefault,
	}
}

// NewCookieOptsFromContext creates CookieOpts by extracting security and host
// information from the context. This is useful for creating cookies that match
// the request's security context (HTTPS vs HTTP, correct domain).
func NewCookieOptsFromContext(ctx context.Context) CookieOpts {
	opts := NewCookieOpts()

	key := security.SecureKey
	if secure, ok := ctx.Value(key).(bool); ok {
		opts.secure = secure
	}

	key = security.HostKey
	if host, ok := ctx.Value(key).(string); ok {
		opts.domain = host
	}

	return opts
}

// SetSecure sets whether the cookie should only be transmitted over HTTPS.
func (co *CookieOpts) SetSecure(secure bool) {
	co.secure = secure
}

// SetTTL sets the time-to-live for the cookie.
func (co *CookieOpts) SetTTL(ttl time.Duration) {
	co.ttl = ttl
}

// SetPath overrides the default path "/" for the cookie.
func (co *CookieOpts) SetPath(path string) {
	co.path = path
}

// SetSameSite sets the SameSite policy for the cookie.
func (co *CookieOpts) SetSameSite(s http.SameSite) {
	co.sameSite = s
}

// NewServerCookie creates a new HTTP cookie with the given name and data.
// The data is base64-encoded before being stored in the cookie value.
// Cookies are created with the HttpOnly flag and the configured SameSite
// policy (Lax unless SetDefaultSameSite is called).
//
// If opts is provided, the first CookieOpts is used to configure
// secure flag, TTL, domain, path, and SameSite policy.
func NewServerCookie(name string, data []byte, opts ...CookieOpts) *http.Cookie {
	value := base64.StdEncoding.EncodeToString(data)

	cookie := &http.Cookie{
		Name:     FullCookieName(name),
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		MaxAge:   int(TTLDefault.Seconds()),
		SameSite: defaultSameSite,
	}

	if len(opts) > 0 {
		cookie.Secure = opts[0].secure
		cookie.MaxAge = int(opts[0].ttl.Seconds())
		cookie.Domain = opts[0].domain
		if opts[0].path != "" {
			cookie.Path = opts[0].path
		}
		if opts[0].sameSite != 0 {
			cookie.SameSite = opts[0].sameSite
		}
	}

	return cookie
}

// ClearCookie creates a cookie that clears/deletes an existing cookie.
// It sets MaxAge to -1 which instructs the browser to delete the cookie.
// Pass the same opts used when setting the cookie so that path and domain match.
func ClearCookie(name string, opts ...CookieOpts) *http.Cookie {
	cookie := &http.Cookie{
		Name:     FullCookieName(name),
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		MaxAge:   -1,
		SameSite: defaultSameSite,
	}

	if len(opts) > 0 {
		cookie.Secure = opts[0].secure
		cookie.Domain = opts[0].domain
		if opts[0].path != "" {
			cookie.Path = opts[0].path
		}
		if opts[0].sameSite != 0 {
			cookie.SameSite = opts[0].sameSite
		}
	}

	return cookie
}

// DecodeValue decodes the base64-encoded value from a cookie.
func DecodeValue(cookie *http.Cookie) ([]byte, error) {
	return base64.StdEncoding.DecodeString(cookie.Value)
}

// GetCookie extracts and decodes a cookie value from a raw cookie header string.
// The fullKey should be the full namespaced cookie name (use FullCookieName).
// Returns the decoded bytes and true if found, or nil and false if not found or decode fails.
func GetCookie(cookies string, fullKey string) ([]byte, bool) {
	iter := strings.SplitSeq(cookies, ";")
	for c := range iter {
		c = strings.TrimSpace(c)
		if !strings.HasPrefix(c, fullKey+"=") {
			continue
		}
		encVal := strings.TrimPrefix(c, fullKey+"=")
		bytes, err := base64.StdEncoding.DecodeString(encVal)
		if err != nil {
			return nil, false
		}
		return bytes, true
	}
	return nil, false
}
