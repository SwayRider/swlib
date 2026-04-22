package middlewares

import (
	"net/http"

	log "github.com/swayrider/swlib/logger"
	"github.com/swayrider/swlib/security"
)

// WebAuth creates an HTTP middleware for web page authentication with redirect support.
// Unlike Auth, on authentication failure for GET requests, it redirects to the login page
// instead of returning 401 Unauthorized.
//
// The server parameter should be the base URL of the auth server (e.g., "https://auth.example.com").
// On auth failure, users are redirected to {server}/login?callback={currentPath}&r=1
// On unverified user, users are redirected to {server}/verify?callback={currentPath}
//
// Parameters:
//   - next: The handler to call after successful authentication
//   - publicKeysFn: Function that returns public keys for JWT verification
//   - l: Logger instance for debug output
//   - server: Base URL of the authentication server for redirects
func WebAuth(
	next http.Handler,
	publicKeysFn security.PublicKeysFn,
	l *log.Logger,
	server string,
) http.Handler {
	return auth(next, publicKeysFn, l, true, server)
}
