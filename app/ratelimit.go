package app

import (
	"time"

	log "github.com/swayrider/swlib/logger"
	"github.com/swayrider/swlib/ratelimit"
)

// Standard RATE_LIMIT_* config fields, shared by every service that enables
// RateLimitInterceptor. Kept in one place so the limiter's tuning knobs are
// consistent across services.
const (
	FldRateLimitRPS         = "rate-limit-rps"
	FldRateLimitBurst       = "rate-limit-burst"
	FldRateLimitIdleTTLSecs = "rate-limit-idle-ttl-secs"

	EnvRateLimitRPS         = "RATE_LIMIT_RPS"
	EnvRateLimitBurst       = "RATE_LIMIT_BURST"
	EnvRateLimitIdleTTLSecs = "RATE_LIMIT_IDLE_TTL_SECS"

	DefRateLimitRPS         = 50
	DefRateLimitBurst       = 100
	DefRateLimitIdleTTLSecs = 300
)

// RateLimitConfigFields returns the standard RATE_LIMIT_* config fields.
// Pass to WithConfigFields on any service that enables RateLimitInterceptor.
func RateLimitConfigFields() []ConfigField {
	return []ConfigField{
		NewIntConfigField(
			FldRateLimitRPS, EnvRateLimitRPS,
			"Sustained requests/sec allowed per source IP, on both the gRPC port and the HTTP gateway",
			DefRateLimitRPS),
		NewIntConfigField(
			FldRateLimitBurst, EnvRateLimitBurst,
			"Burst allowance per source IP", DefRateLimitBurst),
		NewIntConfigField(
			FldRateLimitIdleTTLSecs, EnvRateLimitIdleTTLSecs,
			"Seconds of inactivity before a source IP's rate-limit bucket is evicted",
			DefRateLimitIdleTTLSecs),
	}
}

// RateLimiterInitializer returns an initializer Callback that builds a
// ratelimit.Limiter from parsed RATE_LIMIT_* config and attaches it to
// grpcConfig via SetRateLimiter. Must run as an initializer (after config
// parse, before startGrpc reads GrpcConfig.RateLimiter to build the
// interceptor chain and HTTP middleware), i.e. pass to WithInitializers.
func RateLimiterInitializer(grpcConfig *GrpcConfig) Callback {
	return func(a App) error {
		rps := GetConfigField[int](a.Config(), FldRateLimitRPS)
		burst := GetConfigField[int](a.Config(), FldRateLimitBurst)
		idleTTLSecs := GetConfigField[int](a.Config(), FldRateLimitIdleTTLSecs)
		grpcConfig.SetRateLimiter(
			ratelimit.New(float64(rps), burst, time.Duration(idleTTLSecs)*time.Second))
		return nil
	}
}

// RateLimitEvictor returns a background routine that periodically prunes
// idle rate-limit buckets from grpcConfig.RateLimiter, bounding memory under
// sustained, high-cardinality traffic. Reads grpcConfig.RateLimiter at each
// tick rather than capturing it directly, since it isn't set until
// RateLimiterInitializer runs. Pass to WithBackgroundRoutines alongside
// RateLimiterInitializer.
func RateLimitEvictor(grpcConfig *GrpcConfig) BackgroundRoutine {
	return func(a App) {
		lg := a.Logger().Derive(log.WithFunction("RateLimitEvictor"))
		ctx := a.BackgroundContext()
		defer a.BackgroundWaitGroup().Done()

		ticker := time.NewTicker(5 * time.Minute)
		for {
			select {
			case <-ticker.C:
				if grpcConfig.RateLimiter != nil {
					grpcConfig.RateLimiter.Evict()
				}
			case <-ctx.Done():
				lg.Infoln("stopping rate limit evictor")
				ticker.Stop()
				return
			}
		}
	}
}
