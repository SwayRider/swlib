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
	OrigIpHeader      = "x-forwarded-for"
	OrigHost          = "x-forwarded-host"
	Authority         = ":authority"
	Secure            = "x-forwarded-proto"
	ForwaredUserAgent = "grpcgateway-user-agent"
	UserAgent         = "user-agent"
)

// ClientInfoInterceptor creates a gRPC unary server interceptor that extracts
// client information from request metadata and adds it to the context.
//
// Extracted information includes:
//   - Client IP address (from X-Forwarded-For header)
//   - Host (from X-Forwarded-Host or :authority)
//   - Secure flag (from X-Forwarded-Proto)
//   - User agent (from grpcgateway-user-agent or user-agent)
//
// The extracted values can be retrieved using security.GetOrigIp(),
// security.GetHost(), security.GetSecure(), and security.GetUserAgent().
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
				ip = value[0]
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
