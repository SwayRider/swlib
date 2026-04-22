// Package security provides authentication context utilities and endpoint
// authorization profiles for SwayRider services.
//
// # Context Values
//
// After authentication middleware runs, various values are stored in the
// request context and can be retrieved using the Get* functions:
//
//	claims, ok := security.GetClaims(ctx)     // JWT claims
//	token, ok := security.GetJwt(ctx)         // Raw JWT string
//	ip, ok := security.GetOrigIp(ctx)         // Client IP address
//	host, ok := security.GetHost(ctx)         // Request host
//
// # Endpoint Profiles
//
// Endpoint profiles define authorization requirements for API endpoints:
//
//	// Allow public access (no authentication required)
//	security.PublicEndpoint("/api/health")
//
//	// Require admin privileges
//	security.AdminEndpoint("/api/admin/users")
//
//	// Allow service-to-service calls with specific scopes
//	security.ServiceClientEndpoint("/api/internal/mail", []string{"mail:send"})
package security

import (
	"context"

	"github.com/swayrider/swlib/jwt"
)

// ContextKey is a string type used for context value keys to avoid collisions.
type ContextKey string

// Context keys for storing authentication and request metadata.
const (
	ClaimsKey    ContextKey = "claims"     // JWT claims (*jwt.Claims)
	JwtKey       ContextKey = "jwt"        // Raw JWT token string
	RefreshKey   ContextKey = "refresh"    // Refresh token string
	OrigIpKey    ContextKey = "originalIp" // Original client IP address
	UserAgentKey ContextKey = "userAgent"  // Client user agent string
	HostKey      ContextKey = "host"       // Request host (domain)
	SecureKey    ContextKey = "secure"     // Whether request was over HTTPS
)

// GetClaims retrieves the parsed JWT claims from the context.
// Returns nil and false if no claims are present (unauthenticated request).
func GetClaims(ctx context.Context) (c *jwt.Claims, ok bool) {
	iface := ctx.Value(ClaimsKey)
	c, ok = iface.(*jwt.Claims)
	return
}

// GetJwt retrieves the raw JWT token string from the context.
func GetJwt(ctx context.Context) (j string, ok bool) {
	iface := ctx.Value(JwtKey)
	j, ok = iface.(string)
	return
}

// GetRefreshToken retrieves the refresh token from the context (from cookies).
func GetRefreshToken(ctx context.Context) (r string, ok bool) {
	iface := ctx.Value(RefreshKey)
	r, ok = iface.(string)
	return
}

// GetOrigIp retrieves the original client IP address (from X-Forwarded-For).
func GetOrigIp(ctx context.Context) (ip string, ok bool) {
	iface := ctx.Value(OrigIpKey)
	ip, ok = iface.(string)
	return
}

// GetUserAgent retrieves the client's user agent string.
func GetUserAgent(ctx context.Context) (ua string, ok bool) {
	iface := ctx.Value(UserAgentKey)
	ua, ok = iface.(string)
	return
}

// GetHost retrieves the request host (domain without port).
func GetHost(ctx context.Context) (h string, ok bool) {
	iface := ctx.Value(HostKey)
	h, ok = iface.(string)
	return
}

// GetSecure returns whether the request was made over HTTPS.
func GetSecure(ctx context.Context) (s bool, ok bool) {
	iface := ctx.Value(SecureKey)
	s, ok = iface.(bool)
	return
}
