package interceptors

import (
	"context"
	"io"
	"net"
	"strconv"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	log "github.com/swayrider/swlib/logger"
	"github.com/swayrider/swlib/ratelimit"
	"github.com/swayrider/swlib/security"
)

func init() {
	log.SetOutput(io.Discard)
}

type tcpAddr struct {
	ip   string
	port int
}

func (a tcpAddr) Network() string { return "tcp" }
func (a tcpAddr) String() string  { return net.JoinHostPort(a.ip, strconv.Itoa(a.port)) }

func ctxWithPeer(ip string, port int) context.Context {
	return peer.NewContext(context.Background(), &peer.Peer{Addr: tcpAddr{ip: ip, port: port}})
}

func noopHandler(_ context.Context, _ any) (any, error) {
	return "ok", nil
}

func TestRateLimitInterceptor_AllowsUnderLimit(t *testing.T) {
	limiter := ratelimit.New(100, 5, time.Minute)
	interceptor := RateLimitInterceptor(limiter, log.New())
	ctx := ctxWithPeer("203.0.113.1", 1234)

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}, noopHandler)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRateLimitInterceptor_BlocksOverLimit_ReturnsResourceExhausted(t *testing.T) {
	limiter := ratelimit.New(1, 1, time.Minute)
	interceptor := RateLimitInterceptor(limiter, log.New())
	ctx := ctxWithPeer("203.0.113.1", 1234)
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}

	if _, err := interceptor(ctx, nil, info, noopHandler); err != nil {
		t.Fatalf("first request: expected no error, got %v", err)
	}
	_, err := interceptor(ctx, nil, info, noopHandler)
	if err == nil {
		t.Fatal("second request: expected ResourceExhausted, got nil")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.ResourceExhausted {
		t.Errorf("code = %v, want %v", st.Code(), codes.ResourceExhausted)
	}
}

func TestRateLimitInterceptor_UsesPeerIPNotForgedXFF(t *testing.T) {
	limiter := ratelimit.New(1, 1, time.Minute)
	interceptor := RateLimitInterceptor(limiter, log.New())
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}

	// Real peer is 203.0.113.1, but the caller claims (via a spoofable
	// header) to be a completely different address. The bucket must be
	// keyed off the real peer, not the forged claim.
	ctx := ctxWithPeer("203.0.113.1", 1234)
	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("x-forwarded-for", "9.9.9.9"))

	if _, err := interceptor(ctx, nil, info, noopHandler); err != nil {
		t.Fatalf("first request: expected no error, got %v", err)
	}

	// Same real peer, different claimed XFF -- still the same bucket, so
	// this must be blocked.
	ctx2 := ctxWithPeer("203.0.113.1", 5678)
	ctx2 = metadata.NewIncomingContext(ctx2, metadata.Pairs("x-forwarded-for", "8.8.8.8"))
	_, err := interceptor(ctx2, nil, info, noopHandler)
	if err == nil {
		t.Fatal("expected the second request from the same real peer to be blocked despite a different forged XFF")
	}
}

func TestRateLimitInterceptor_StripsPortFromPeerAddr(t *testing.T) {
	limiter := ratelimit.New(1, 1, time.Minute)
	interceptor := RateLimitInterceptor(limiter, log.New())
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}

	if _, err := interceptor(ctxWithPeer("203.0.113.1", 1111), nil, info, noopHandler); err != nil {
		t.Fatalf("first request: expected no error, got %v", err)
	}
	// Same IP, different ephemeral port -- must share the same bucket, or
	// every new TCP connection would get its own bucket (no limiting at all).
	_, err := interceptor(ctxWithPeer("203.0.113.1", 2222), nil, info, noopHandler)
	if err == nil {
		t.Fatal("expected same-IP-different-port request to share the exhausted bucket")
	}
}

func TestRateLimitInterceptor_SkipsExemptEndpoint(t *testing.T) {
	security.SkipRateLimitEndpoint("/health.v1.HealthService/Ping")

	limiter := ratelimit.New(1, 1, time.Minute)
	interceptor := RateLimitInterceptor(limiter, log.New())
	info := &grpc.UnaryServerInfo{FullMethod: "/health.v1.HealthService/Ping"}
	ctx := ctxWithPeer("203.0.113.1", 1234)

	for i := range 5 {
		if _, err := interceptor(ctx, nil, info, noopHandler); err != nil {
			t.Fatalf("request %d to exempt endpoint: expected no error, got %v", i, err)
		}
	}
}

func TestRateLimitInterceptor_MissingPeerInfo_FailsOpen(t *testing.T) {
	limiter := ratelimit.New(1, 1, time.Minute)
	interceptor := RateLimitInterceptor(limiter, log.New())
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}

	_, err := interceptor(context.Background(), nil, info, noopHandler)
	if err != nil {
		t.Fatalf("expected fail-open (no error) when peer info is missing, got %v", err)
	}
}

func TestRateLimitInterceptor_ExemptsLoopbackPeer(t *testing.T) {
	limiter := ratelimit.New(1, 1, time.Minute)
	interceptor := RateLimitInterceptor(limiter, log.New())
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}
	ctx := ctxWithPeer("127.0.0.1", 1234)

	// Loopback peers (the in-process grpc-gateway proxy) share one address
	// across every real end user's request, so they must be exempt --
	// otherwise all REST-gateway traffic would collapse into one bucket.
	for i := range 5 {
		if _, err := interceptor(ctx, nil, info, noopHandler); err != nil {
			t.Fatalf("request %d from loopback peer: expected no error, got %v", i, err)
		}
	}
}
