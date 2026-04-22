// Package interceptors provides gRPC server interceptors for authentication
// and client information extraction.
package interceptors

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"github.com/swayrider/swlib/http/cookies"
	log "github.com/swayrider/swlib/logger"
	"github.com/swayrider/swlib/security"
)

// AuthInterceptor creates a gRPC unary server interceptor that validates JWT tokens.
// It extracts tokens from the Authorization header (Bearer scheme) and validates
// them against the configured endpoint profiles.
//
// On successful authentication, JWT claims are added to the request context.
// The refresh token from cookies is also extracted and added to the context if present.
//
// Parameters:
//   - publicKeysFn: Function that returns the public keys for JWT verification
//   - l: Logger instance for debug output
//
// Example:
//
//	server := grpc.NewServer(
//	    grpc.UnaryInterceptor(interceptors.AuthInterceptor(getPublicKeys, logger)),
//	)
func AuthInterceptor(
	publicKeysFn security.PublicKeysFn,
	l *log.Logger,
) grpc.UnaryServerInterceptor {
	lg := l.Derive(log.WithFunction("AuthInterceptor"))

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			lg.Debugln("No metadata found in grpc context")
			return nil, status.Errorf(codes.Internal, "no metadata found")
		}

		refreshToken := ""
		if cookieList := md.Get("cookie"); len(cookieList) > 0 {
			cookieName := cookies.FullCookieName("refresh_token")
			if bytes, ok := cookies.GetCookie(cookieList[0], cookieName); ok {
				refreshToken = string(bytes)
			}
		}

		var tokenStr *string
		authHeader := md.Get("authorization")
		if len(authHeader) != 0 {
			ts := strings.TrimPrefix(authHeader[0], "Bearer ")
			tokenStr = &ts
		}

		profile := security.GetEndpointProfile(info.FullMethod)
		claims, err := profile.Evaluate(tokenStr, publicKeysFn, lg)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "%v", err)
		}
		if claims != nil {
			key := security.ClaimsKey
			ctx = context.WithValue(ctx, key, claims)
			key = security.JwtKey
			ctx = context.WithValue(ctx, key, *tokenStr)
		}

		if refreshToken != "" {
			key := security.RefreshKey
			ctx = context.WithValue(ctx, key, refreshToken)
		}

		return handler(ctx, req)
	}
}
