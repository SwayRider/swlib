package interceptors

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	log "github.com/swayrider/swlib/logger"
	"github.com/swayrider/swlib/security"
)

// invoke runs ClientInfoInterceptor over the given incoming metadata and
// captures the values it extracts into the context.
func invoke(t *testing.T, md metadata.MD) (ip, host, ua string, secure bool) {
	t.Helper()
	interceptor := ClientInfoInterceptor(log.New())

	var gotIP, gotHost, gotUA string
	var gotSecure bool
	handler := func(ctx context.Context, _ any) (any, error) {
		gotIP, _ = security.GetOrigIp(ctx)
		gotHost, _ = security.GetHost(ctx)
		gotSecure, _ = security.GetSecure(ctx)
		gotUA, _ = security.GetUserAgent(ctx)
		return nil, nil
	}

	ctx := metadata.NewIncomingContext(context.Background(), md)
	if _, err := interceptor(
		ctx,
		&struct{}{},
		&grpc.UnaryServerInfo{FullMethod: "/test/method"},
		handler,
	); err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}
	return gotIP, gotHost, gotUA, gotSecure
}

func TestClientInfoInterceptor_OrigIP(t *testing.T) {
	ip, _, _, _ := invoke(t, metadata.Pairs(OrigIpHeader, "1.2.3.4"))
	if ip != "1.2.3.4" {
		t.Errorf("ip = %q, want %q", ip, "1.2.3.4")
	}
}

func TestClientInfoInterceptor_OrigIP_NormalizesChain(t *testing.T) {
	// The gateway forwards a single resolved IP, but a chain must still be
	// reduced to its first entry so the context value is never ambiguous.
	for _, chain := range []string{
		"1.2.3.4, 10.0.0.1",
		"1.2.3.4,10.0.0.1",
		"  1.2.3.4 , 10.0.0.1",
	} {
		ip, _, _, _ := invoke(t, metadata.Pairs(OrigIpHeader, chain))
		if ip != "1.2.3.4" {
			t.Errorf("chain %q: ip = %q, want %q", chain, ip, "1.2.3.4")
		}
	}
}

func TestClientInfoInterceptor_IgnoresClientXForwardedFor(t *testing.T) {
	// Client-supplied X-Forwarded-For must NOT be honored as the client IP:
	// on a direct (non-gatewayed) connection it is attacker-controlled.
	ip, _, _, _ := invoke(t, metadata.Pairs("x-forwarded-for", "6.6.6.6"))
	if ip != "" {
		t.Errorf("ip = %q, want empty (client X-Forwarded-For ignored)", ip)
	}
}

func TestClientInfoInterceptor_OrigIP_Missing(t *testing.T) {
	ip, _, _, _ := invoke(t, metadata.Pairs("x-forwarded-host", "example.com"))
	if ip != "" {
		t.Errorf("ip = %q, want empty when x-orig-ip is absent", ip)
	}
}

func TestClientInfoInterceptor_HostAndSecure(t *testing.T) {
	_, host, _, secure := invoke(t, metadata.Pairs(
		OrigHost, "example.com:443",
		Secure, "https",
	))
	if host != "example.com" {
		t.Errorf("host = %q, want %q (port stripped)", host, "example.com")
	}
	if !secure {
		t.Error("secure = false, want true for x-forwarded-proto https")
	}
}

func TestClientInfoInterceptor_HostAndSecure_Insecure(t *testing.T) {
	_, host, _, secure := invoke(t, metadata.Pairs(
		OrigHost, "example.com",
		Secure, "http",
	))
	if host != "example.com" {
		t.Errorf("host = %q, want %q", host, "example.com")
	}
	if secure {
		t.Error("secure = true, want false for x-forwarded-proto http")
	}
}

func TestClientInfoInterceptor_UserAgent(t *testing.T) {
	_, _, ua, _ := invoke(t, metadata.Pairs(
		UserAgent, "curl/8.0",
	))
	if ua != "curl/8.0" {
		t.Errorf("ua = %q, want %q", ua, "curl/8.0")
	}
}

func TestClientInfoInterceptor_ForwardedUserAgentTakesPrecedence(t *testing.T) {
	_, _, ua, _ := invoke(t, metadata.Pairs(
		ForwaredUserAgent, "grpc-gateway-ua",
		UserAgent, "curl/8.0",
	))
	if ua != "grpc-gateway-ua" {
		t.Errorf("ua = %q, want grpc-gateway-user-agent to take precedence", ua)
	}
}
