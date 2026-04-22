package middlewares

import (
	"context"
	"net/http"
	"strings"

	log "github.com/swayrider/swlib/logger"
	"github.com/swayrider/swlib/security"
)

// SourceInfo is an HTTP middleware that extracts request source information
// and adds it to the request context for use by downstream handlers.
//
// Extracted information:
//   - Host: from X-Forwarded-Host header, or r.Host as fallback (port stripped)
//   - Secure: true if X-Forwarded-Proto is "https" or r.TLS is present
//
// These values can be retrieved using security.GetHost(ctx) and security.GetSecure(ctx),
// and are used by the cookies package to create properly configured cookies.
func SourceInfo(
	next http.Handler,
	l *log.Logger,
) http.Handler {
	return sourceInfo(next, l)
}

func sourceInfo(
	next http.Handler,
	l *log.Logger,
) http.Handler {
	lg := l.Derive(log.WithFunction("middlewares.SourceInfo"))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		host := getRequestHost(r)
		key := security.HostKey
		ctx = context.WithValue(ctx, key, host)

		secure := getRequestSecure(r)
		key = security.SecureKey
		ctx = context.WithValue(ctx, key, secure)

		lg.Debugf("sourceInfor: %s (secure=%v)", host, secure)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getRequestHost extracts the host from the request, preferring X-Forwarded-Host
// header (set by reverse proxies) over r.Host. Port numbers are stripped.
func getRequestHost(r *http.Request) string {
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	if strings.Contains(host, ":") {
		host, _, _ = strings.Cut(host, ":")
	}
	return host
}

// getRequestSecure determines if the request was made over HTTPS by checking
// X-Forwarded-Proto header (set by reverse proxies) or the presence of TLS.
func getRequestSecure(r *http.Request) bool {
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return strings.EqualFold(proto, "https")
	}
	if r.TLS != nil {
		return true
	}
	return false
}
