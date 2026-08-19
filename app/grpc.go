package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"github.com/swayrider/swlib/grpc/interceptors"
	"github.com/swayrider/swlib/http/middlewares"
	log "github.com/swayrider/swlib/logger"
	"github.com/swayrider/swlib/ratelimit"
	"github.com/swayrider/swlib/security"
)

// GrpcInterceptor is a bitmask type for selecting gRPC server interceptors.
// Multiple interceptors can be combined using bitwise OR.
type GrpcInterceptor = uint16

// gRPC interceptor options for server-side middleware.
const (
	// NoInterceptor disables all interceptors
	NoInterceptor GrpcInterceptor = 0x0000
	// AuthInterceptor enables JWT authentication for incoming requests
	AuthInterceptor GrpcInterceptor = 0x0001
	// ClientInfoInterceptor extracts client information (IP, user agent) from requests
	ClientInfoInterceptor GrpcInterceptor = 0x0002
	// RateLimitInterceptor applies a coarse per-peer-IP rate limit ahead of
	// authentication. Requires RateLimiter to be set on GrpcConfig.
	RateLimitInterceptor GrpcInterceptor = 0x0004
)

// ServiceRegistrar is a function type for registering gRPC services with the server.
// It receives the gRPC ServiceRegistrar and the App instance.
type ServiceRegistrar func(grpc.ServiceRegistrar, App)

// ServiceHTTPHandler is a function type for registering HTTP gateway handlers.
// It's used to expose gRPC services via REST endpoints using grpc-gateway.
type ServiceHTTPHandler func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error

// ForwardResponseFn is a callback for modifying HTTP responses in the gateway.
type ForwardResponseFn func(context.Context, http.ResponseWriter, proto.Message) error

// HeaderMathcerFn is used to customize HTTP header to gRPC metadata mapping.
type HeaderMathcerFn = runtime.HeaderMatcherFunc

// GrpcServiceHooks combines the gRPC service registration and HTTP handler setup.
type GrpcServiceHooks struct {
	ServiceRegistrar   ServiceRegistrar
	ServiceHTTPHandler ServiceHTTPHandler
}

// GrpcConfig holds the configuration for the gRPC server and HTTP gateway.
type GrpcConfig struct {
	// Interceptors specifies which gRPC interceptors to enable
	Interceptors GrpcInterceptor
	// JWTPublicKeysFn provides public keys for JWT validation (required if AuthInterceptor enabled)
	JWTPublicKeysFn security.PublicKeysFn
	// ServiceRegistrars contains the gRPC service registration hooks
	ServiceRegistrars []GrpcServiceHooks
	// ForwardResponseFn is an optional callback for modifying HTTP responses
	ForwardResponseFn ForwardResponseFn
	// HeaderMatcherFn is an optional callback for customizing header mapping
	HeaderMatcherFn HeaderMathcerFn
	// RateLimiter backs RateLimitInterceptor when that bit is set in Interceptors.
	// The same limiter and bit also gate the HTTP gateway's rate-limit
	// middleware, so one opt-in covers both the raw gRPC port and REST.
	RateLimiter *ratelimit.Limiter
	// MaxRecvMsgSizeBytes caps incoming gRPC message size when > 0. Leave
	// unset (0) to keep gRPC's implicit ~4MB default.
	MaxRecvMsgSizeBytes int
	// AllowCredentials controls whether the HTTP gateway's CORS policy allows
	// credentialed cross-origin requests (cookies, or fetch credentials:
	// "include"). Defaults to false — only services that actually set
	// cookies (e.g. authservice's refresh-token cookie) should opt in.
	AllowCredentials bool
}

// NewGrpcConfig creates a new GrpcConfig with the specified interceptors,
// JWT public keys function, and service registrars.
//
// Example:
//
//	grpcConfig := app.NewGrpcConfig(
//	    app.AuthInterceptor | app.ClientInfoInterceptor,
//	    getPublicKeys,
//	    app.GrpcServiceHooks{
//	        ServiceRegistrar:   registerMyService,
//	        ServiceHTTPHandler: pb.RegisterMyServiceHandlerFromEndpoint,
//	    },
//	)
func NewGrpcConfig(
	interceptors GrpcInterceptor,
	jwtPublicKeysFn security.PublicKeysFn,
	serviceRegistrars ...GrpcServiceHooks,
) *GrpcConfig {
	return &GrpcConfig{
		Interceptors:      interceptors,
		JWTPublicKeysFn:   jwtPublicKeysFn,
		ServiceRegistrars: serviceRegistrars,
	}
}

func (cfg *GrpcConfig) SetForwardResponseFn(fn ForwardResponseFn) {
	cfg.ForwardResponseFn = fn
}

func (cfg *GrpcConfig) SetHeaderMatcherFn(fn HeaderMathcerFn) {
	cfg.HeaderMatcherFn = fn
}

func (cfg *GrpcConfig) SetRateLimiter(l *ratelimit.Limiter) {
	cfg.RateLimiter = l
}

func (cfg *GrpcConfig) SetMaxRecvMsgSize(n int) {
	cfg.MaxRecvMsgSizeBytes = n
}

func (cfg *GrpcConfig) SetAllowCredentials(v bool) {
	cfg.AllowCredentials = v
}

// validateCORSOrigins rejects a bare "*" origin paired with
// AllowCredentials: true. rs/cors would turn that combination into "reflect
// every origin with credentials", silently opening credentialed cross-origin
// access to any site. Scoped wildcard patterns like "https://*.example.com"
// are unaffected — this only guards against a bare "*" entry.
func validateCORSOrigins(origins []string) error {
	if slices.Contains(origins, "*") {
		return errors.New("CORS misconfiguration: AllowCredentials is enabled but origins contain \"*\"")
	}
	return nil
}

func (a *app) startGrpc() {
	lg := a.lg.Derive(log.WithFunction("startGrpc"))
	if a.grpcConfig == nil {
		return
	}

	httpPort := GetConfigField[int](a.cfg, KeyHttpPort)
	grpcPort := GetConfigField[int](a.cfg, KeyGrpcPort)

	// Hook up grpc
	interceptorList := a.grpcInterceptors(lg)
	var serverOpts []grpc.ServerOption
	if len(interceptorList) == 1 {
		serverOpts = append(serverOpts, grpc.UnaryInterceptor(interceptorList[0]))
	} else if len(interceptorList) > 0 {
		serverOpts = append(serverOpts, grpc.ChainUnaryInterceptor(interceptorList...))
	}
	if a.grpcConfig.MaxRecvMsgSizeBytes > 0 {
		serverOpts = append(serverOpts, grpc.MaxRecvMsgSize(a.grpcConfig.MaxRecvMsgSizeBytes))
	}
	a.grpcServer = grpc.NewServer(serverOpts...)

	for _, r := range a.grpcConfig.ServiceRegistrars {
		r.ServiceRegistrar(a.grpcServer, a)
	}

	// start grpc server
	lis, err := net.Listen(
		"tcp",
		fmt.Sprintf("[::]:%d", grpcPort))
	if err != nil {
		lg.Fatalf("failed to listen: %v", err)
	}
	go func() {
		if err := a.grpcServer.Serve(lis); err != nil {
			lg.Fatalf("gRPC server stopped with error: %v", err)
		}
	}()

	// HTTP startup
	var opts []runtime.ServeMuxOption
	if a.grpcConfig.ForwardResponseFn != nil {
		opts = append(opts, runtime.WithForwardResponseOption(a.grpcConfig.ForwardResponseFn))
	}
	if a.grpcConfig.HeaderMatcherFn != nil {
		opts = append(opts, runtime.WithIncomingHeaderMatcher(a.grpcConfig.HeaderMatcherFn))
	}
	mux := runtime.NewServeMux(opts...)

	gwOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	}

	for _, r := range a.grpcConfig.ServiceRegistrars {
		if err := r.ServiceHTTPHandler(
			context.Background(),
			mux,
			fmt.Sprintf("[::]:%d", grpcPort),
			gwOpts); err != nil {
			lg.Fatalf("failed to register HTTP handler: %v", err)
		}
	}

	corsOrigins := []string{
		"http://localhost:5173",
		"http://*.hevanto-it.com",
		"https://*.hevanto-it.com",
		"https://*.swayrider.com",
	}
	if a.grpcConfig.AllowCredentials {
		if err := validateCORSOrigins(corsOrigins); err != nil {
			lg.Fatalf("%v", err)
		}
	}

	handler := cors.New(cors.Options{
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{"OPTIONS", "GET", "POST"},
		AllowedOrigins:   corsOrigins,
		AllowCredentials: a.grpcConfig.AllowCredentials,
	}).Handler(mux)

	if (a.grpcConfig.Interceptors&RateLimitInterceptor) == RateLimitInterceptor && a.grpcConfig.RateLimiter != nil {
		handler = middlewares.RateLimit(handler, a.grpcConfig.RateLimiter, lg)
	}

	a.httpGateway = &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: handler,
	}
	go func() {
		if err := a.httpGateway.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			lg.Fatalf("HTTP server stopped with error: %v", err)
		}
	}()

	lg.Infof("HTTP server running on port: %d", httpPort)
	lg.Infof("gRPC server running on port: %d", grpcPort)
}

func (a *app) stopGrpcServer() {
	lg := a.lg.Derive(log.WithFunction("stopGrpcServer"))
	if a.grpcConfig == nil {
		return
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.httpGateway.Shutdown(shutdownCtx); err != nil {
		lg.Errorf("failed to shutdown http server: %v", err)
	} else {
		lg.Infoln("HTTP server stopped")
	}

	a.grpcServer.GracefulStop()
	lg.Infoln("GRPC server stopped")
}

func (a *app) grpcInterceptors(lg *log.Logger) []grpc.UnaryServerInterceptor {
	var lst []grpc.UnaryServerInterceptor

	// RateLimit runs first, ahead of auth/client-info parsing, so an
	// over-limit caller is rejected before the server does any further work
	// on the request.
	if (a.grpcConfig.Interceptors & RateLimitInterceptor) == RateLimitInterceptor {
		lst = append(lst, interceptors.RateLimitInterceptor(
			a.grpcConfig.RateLimiter, lg))
	}
	if (a.grpcConfig.Interceptors & AuthInterceptor) == AuthInterceptor {
		lst = append(lst, interceptors.AuthInterceptor(
			a.grpcConfig.JWTPublicKeysFn, lg))
	}
	if (a.grpcConfig.Interceptors & ClientInfoInterceptor) == ClientInfoInterceptor {
		lst = append(lst, interceptors.ClientInfoInterceptor(a.lg))
	}
	return lst
}
