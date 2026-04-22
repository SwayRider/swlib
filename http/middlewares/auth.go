// Package middlewares provides HTTP middleware functions for authentication,
// content-type handling, file serving security, and request metadata extraction.
//
// # Authentication Middleware
//
// The Auth and WebAuth middlewares validate JWT tokens and enforce endpoint
// authorization profiles. Auth is for API endpoints, WebAuth for web pages
// (with redirect support).
//
// # Usage
//
//	// API authentication
//	handler = middlewares.Auth(handler, getPublicKeys, logger)
//
//	// Web page authentication with login redirect
//	handler = middlewares.WebAuth(handler, getPublicKeys, logger, "https://auth.example.com")
//
//	// Add source info (host, secure flag) to context
//	handler = middlewares.SourceInfo(handler, logger)
//
//	// Set Content-Type based on file extension
//	handler = middlewares.MimeType(handler)
//
//	// Prevent directory listing
//	handler = middlewares.NeuterFS(handler)
package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/swayrider/swlib/http/cookies"
	log "github.com/swayrider/swlib/logger"
	"github.com/swayrider/swlib/security"
)

// Auth creates an HTTP middleware that validates JWT tokens for API endpoints.
// It extracts tokens from cookies (access_token) or the Authorization header (Bearer scheme).
//
// On successful authentication, JWT claims are added to the request context.
// On failure, it returns HTTP 401 Unauthorized.
//
// Parameters:
//   - next: The handler to call after successful authentication
//   - publicKeysFn: Function that returns public keys for JWT verification
//   - l: Logger instance for debug output
func Auth(
	next http.Handler,
	publicKeysFn security.PublicKeysFn,
	l *log.Logger,
) http.Handler {
	return auth(next, publicKeysFn, l, false, "")
}

func auth(
	next http.Handler,
	publicKeysFn security.PublicKeysFn,
	l *log.Logger,
	isWebPage bool,
	server string,
) http.Handler {
	lg := l.Derive(log.WithFunction("middlewares.Auth"))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenStr *string

		// First we try cookies
		cookie, err := r.Cookie(cookies.FullCookieName("access_token"))
		if err != nil {
			// Then we try the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				ts := strings.TrimPrefix(authHeader, "Bearer ")
				tokenStr = &ts
			}
		} else {
			bytes, err := cookies.DecodeValue(cookie)
			if err != nil {
				lg.Debugf("authorization error: %v", err)
				if isWebPage && r.Method == http.MethodGet {
					callback := r.URL.Query().Get("callback")
					if callback == "" {
						callback = r.URL.Path
					}
					loginUrl := fmt.Sprintf(
						"%s/login?callback=%s&r=1",
						server,
						url.QueryEscape(callback))
					http.Redirect(w, r, loginUrl, http.StatusSeeOther)
					return
				}
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			str := string(bytes)
			tokenStr = &str
		}

		profile := security.GetEndpointProfileForMethod(r.URL.Path, r.Method)
		claims, err := profile.Evaluate(tokenStr, publicKeysFn, lg)
		if err != nil {
			lg.Debugf("authorization error: %v", err)
			if isWebPage && r.Method == http.MethodGet {
				callback := r.URL.Query().Get("callback")
				if callback == "" {
					callback = r.URL.Path
				}
				if err == security.ErrUserNotVerified {
					verifyUrl := fmt.Sprintf(
						"%s/verify?callback=%s",
						server,
						url.QueryEscape(callback))
					http.Redirect(w, r, verifyUrl, http.StatusSeeOther)
					return
				}
				loginUrl := fmt.Sprintf(
					"%s/login?callback=%s&r=1",
					server,
					url.QueryEscape(callback))
				http.Redirect(w, r, loginUrl, http.StatusSeeOther)
				return
			}
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := r.Context()
		if claims != nil {
			key := security.ClaimsKey
			ctx = context.WithValue(ctx, key, claims)
			key = security.JwtKey
			ctx = context.WithValue(ctx, key, *tokenStr)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
