package interceptors

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	log "github.com/swayrider/swlib/logger"
	"github.com/swayrider/swlib/ratelimit"
	"github.com/swayrider/swlib/security"
)

// RateLimitInterceptor creates a gRPC unary server interceptor that applies a
// coarse, per-caller-IP rate limit ahead of authentication -- a fallback
// safety net for services that must never be exposed directly (see each
// service's README security-boundary note), in case that boundary is ever
// bypassed or misconfigured.
//
// It deliberately does NOT trust security.GetOrigIp/X-Forwarded-For: those
// are set by ClientInfoInterceptor from caller-supplied metadata, which is
// exactly what's forgeable (or absent) on a direct, un-gatewayed connection.
// It uses the real transport peer address instead, via peer.FromContext, so
// it has no ordering dependency on ClientInfoInterceptor.
//
// Traffic proxied in-process by a grpc-gateway HTTP handler (the normal path
// for a service's own REST API) arrives here over a local loopback
// connection shared by every such request, regardless of which end user
// originally made the HTTP call -- so bucketing by peer IP would collapse
// all real REST users into one shared bucket. Loopback peers are exempted
// for this reason: that path is already fronted by the HTTP gateway's own
// protections, and this interceptor's job is specifically to catch direct,
// non-gatewayed network callers.
func RateLimitInterceptor(limiter *ratelimit.Limiter, l *log.Logger) grpc.UnaryServerInterceptor {
	lg := l.Derive(log.WithFunction("RateLimitInterceptor"))
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if security.GetEndpointProfileForMethod(info.FullMethod, "").SkipRateLimit {
			return handler(ctx, req)
		}

		p, ok := peer.FromContext(ctx)
		if !ok || p.Addr == nil {
			// Best-effort coarse control: a rare transport edge case where
			// peer info is unavailable must not itself take down traffic to
			// a security-critical service.
			lg.Warnf("no peer info available, failing open for %s", info.FullMethod)
			return handler(ctx, req)
		}

		host, _, err := net.SplitHostPort(p.Addr.String())
		if err != nil {
			// Some peer.Addr implementations (e.g. in-memory/bufconn) have
			// no port to strip.
			host = p.Addr.String()
		}

		if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
			return handler(ctx, req)
		}

		if !limiter.Allow(host) {
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}
		return handler(ctx, req)
	}
}
