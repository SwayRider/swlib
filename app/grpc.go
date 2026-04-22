package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"github.com/swayrider/swlib/grpc/interceptors"
	log "github.com/swayrider/swlib/logger"
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

func (a *app) startGrpc() {
	lg := a.lg.Derive(log.WithFunction("startGrpc"))
	if a.grpcConfig == nil {
		return
	}

	httpPort := GetConfigField[int](a.cfg, KeyHttpPort)
	grpcPort := GetConfigField[int](a.cfg, KeyGrpcPort)

	// Hook up grpc
	interceptorList := a.grpcInterceptors(lg)
	if len(interceptorList) == 1 {
		a.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(interceptorList[0]))
	} else if len(interceptorList) > 0 {
		a.grpcServer = grpc.NewServer(grpc.ChainUnaryInterceptor(interceptorList...))
	} else {
		a.grpcServer = grpc.NewServer()
	}

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
		r.ServiceHTTPHandler(
			context.Background(),
			mux,
			fmt.Sprintf("[::]:%d", grpcPort),
			gwOpts)
	}

	handler := cors.New(cors.Options{
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{"OPTIONS", "GET", "POST"},
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://*.hevanto-it.com",
			"https://*.hevanto-it.com",
			"https://*.swayrider.com",
		},
		AllowCredentials: true,
	}).Handler(mux)

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

	if (a.grpcConfig.Interceptors & AuthInterceptor) == AuthInterceptor {
		lst = append(lst, interceptors.AuthInterceptor(
			a.grpcConfig.JWTPublicKeysFn, lg))
	}
	if (a.grpcConfig.Interceptors & ClientInfoInterceptor) == ClientInfoInterceptor {
		lst = append(lst, interceptors.ClientInfoInterceptor(a.lg))
	}
	return lst
}
