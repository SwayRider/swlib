package middlewares

import (
	"net"
	"net/http"

	log "github.com/swayrider/swlib/logger"
	"github.com/swayrider/swlib/ratelimit"
)

// RateLimit is an HTTP middleware that applies a coarse, per-client-IP rate
// limit to the gRPC-gateway's HTTP surface.
//
// The gRPC-side RateLimitInterceptor (swlib/grpc/interceptors) can't protect
// this surface on its own: it keys on the gRPC transport peer address, but
// grpc-gateway proxies REST calls to the gRPC server in-process over
// loopback, so every REST caller looks like the same loopback peer to that
// interceptor and is exempted from it entirely. This middleware runs earlier,
// directly in the HTTP server, where r.RemoteAddr is the real TCP peer
// address filled in by net/http itself -- not a client-suppliable header
// like X-Forwarded-For, so it can't be forged by the caller.
func RateLimit(next http.Handler, limiter *ratelimit.Limiter, l *log.Logger) http.Handler {
	lg := l.Derive(log.WithFunction("middlewares.RateLimit"))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}

		if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
			next.ServeHTTP(w, r)
			return
		}

		if !limiter.Allow(host) {
			lg.Warnf("rate limit exceeded for %s %s", host, r.URL.Path)
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
