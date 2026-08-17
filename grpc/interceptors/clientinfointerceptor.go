package interceptors

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	log "github.com/swayrider/swlib/logger"
	"github.com/swayrider/swlib/security"
)

// HTTP header names for extracting client information
const (
	// OrigIpHeader carries the client IP as resolved by the API gateway and
	// forwarded as gRPC metadata. It is deliberately NOT "x-forwarded-for":
	// that header is client-controllable on any direct (non-gatewayed)
	// connection, and this interceptor runs on services that have no trusted
	// reverse proxy of their own. Only the gateway is a legitimate producer
	// of this value.
	OrigIpHeader      = "x-orig-ip"
	OrigHost          = "x-forwarded-host"
	Authority         = ":authority"
	Secure            = "x-forwarded-proto"
	ForwaredUserAgent = "grpcgateway-user-agent"
	UserAgent         = "user-agent"
)

// firstIP returns the first non-empty, whitespace-trimmed entry of a
// comma-separated IP chain. Per the X-Forwarded-For convention the original
// client is the first entry; a gateway that resolved the chain forwards a
// single IP, so this is normally a no-op. It guarantees the context value is
// always a single IP, never a raw comma-joined header.
func firstIP(values []string) string {
	for _, v := range values {
		for _, part := range strings.Split(v, ",") {
			if ip := strings.TrimSpace(part); ip != "" {
				return ip
			}
		}
	}
	return ""
}

// ClientInfoInterceptor creates a gRPC unary server interceptor that extracts
// client information from request metadata and adds it to the context.
//
// Extracted information includes:
//   - Client IP address (from the gateway-forwarded x-orig-ip metadata)
//   - Host (from X-Forwarded-Host or :authority)
//   - Secure flag (from X-Forwarded-Proto)
//   - User agent (from grpcgateway-user-agent or user-agent)
//
// The extracted values can be retrieved using security.GetOrigIp(),
// security.GetHost(), security.GetSecure(), and security.GetUserAgent().
//
// Client-supplied X-Forwarded-For is deliberately NOT honored as the client
// IP: on a direct (non-gatewayed) connection it is entirely attacker-
// controlled. Only the gateway-forwarded x-orig-ip metadata is trusted.
//
// Example:
//
//	server := grpc.NewServer(
//	    grpc.UnaryInterceptor(interceptors.ClientInfoInterceptor(logger)),
//	)
func ClientInfoInterceptor(l *log.Logger) grpc.UnaryServerInterceptor {
	lg := l.Derive(log.WithFunction("ClientInfoInterceptor"))
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			lg.Debugln("No metadata found in grpc context")
			return nil, status.Errorf(codes.Unauthenticated, "no metadata found")
		}

		var ip, host, hostFallback, forwardedUserAgent, userAgent string
		var secure bool

		for key, value := range md {
			switch strings.ToLower(key) {
			case OrigIpHeader:
				ip = firstIP(value)
			case OrigHost:
				host = value[0]
			case Authority:
				hostFallback = value[0]
			case Secure:
				secure = strings.EqualFold(value[0], "https")
			case ForwaredUserAgent:
				forwardedUserAgent = value[0]
			case UserAgent:
				userAgent = value[0]
			}
		}
		if host == "" {
			host = hostFallback
		}
		if strings.Contains(host, ":") {
			host, _, _ = strings.Cut(host, ":")
		}
		if host != "" {
			lg.Debugf("OrigHost: %s (secure=%v)", host, secure)
			ctx = context.WithValue(ctx, security.HostKey, host)
			ctx = context.WithValue(ctx, security.SecureKey, secure)
		}

		if ip != "" {
			lg.Debugf("OrigIpHeader: %s", ip)
			ctx = context.WithValue(ctx, security.OrigIpKey, ip)
		}
		if forwardedUserAgent != "" {
			lg.Debugf("ForwardedUserAgent: %s", forwardedUserAgent)
			ctx = context.WithValue(ctx, security.UserAgentKey, forwardedUserAgent)
		} else if userAgent != "" {
			lg.Debugf("UserAgent: %s", userAgent)
			ctx = context.WithValue(ctx, security.UserAgentKey, userAgent)
		}

		return handler(ctx, req)
	}
}
