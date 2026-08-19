package app

import (
	"time"

	"github.com/swayrider/grpcclients/authclient"
	"github.com/swayrider/swlib/jwtkeys"
	log "github.com/swayrider/swlib/logger"
)

// Standard JWT_KEYS_* config fields, shared by every service that verifies
// JWTs issued by authservice via a jwtkeys.Cache.
const (
	FldJWTKeysRefreshIntervalSecs = "jwt-keys-refresh-interval-secs"
	EnvJWTKeysRefreshIntervalSecs = "JWT_KEYS_REFRESH_INTERVAL_SECS"
	DefJWTKeysRefreshIntervalSecs = 300 // 5 minutes

	FldJWTKeysFetchTimeoutSecs = "jwt-keys-fetch-timeout-secs"
	EnvJWTKeysFetchTimeoutSecs = "JWT_KEYS_FETCH_TIMEOUT_SECS"
	DefJWTKeysFetchTimeoutSecs = 15
)

// AuthServiceClientName is the name every service registers its authservice
// client under via app.NewServiceClient / app.GetServiceClient.
const AuthServiceClientName = "authservice"

// JWTKeysConfigFields returns the standard JWT_KEYS_* config fields. Pass
// to WithConfigFields on any service using a jwtkeys.Cache.
func JWTKeysConfigFields() []ConfigField {
	return []ConfigField{
		NewIntConfigField(
			FldJWTKeysRefreshIntervalSecs, EnvJWTKeysRefreshIntervalSecs,
			"How often (seconds) to refresh JWT verification public keys from authservice",
			DefJWTKeysRefreshIntervalSecs),
		NewIntConfigField(
			FldJWTKeysFetchTimeoutSecs, EnvJWTKeysFetchTimeoutSecs,
			"Upper bound (seconds) on a single public-key fetch from authservice",
			DefJWTKeysFetchTimeoutSecs),
	}
}

// JWTKeysInitializer returns an initializer Callback that tunes cache from
// parsed JWT_KEYS_* config. Must run as an initializer (before
// JWTKeysFetcher's background routine starts ticking, since initializers
// complete before background routines start — see Run's execution order),
// i.e. pass to WithInitializers.
func JWTKeysInitializer(cache *jwtkeys.Cache) Callback {
	return func(a App) error {
		refreshSecs := GetConfigField[int](a.Config(), FldJWTKeysRefreshIntervalSecs)
		timeoutSecs := GetConfigField[int](a.Config(), FldJWTKeysFetchTimeoutSecs)
		cache.Configure(
			time.Duration(refreshSecs)*time.Second,
			time.Duration(timeoutSecs)*time.Second)
		return nil
	}
}

// JWTKeysFetcher returns a background routine that resolves the authservice
// client (only available once the App starts running — see
// app.GetServiceClient) and runs cache's refresh loop until shutdown. Pass
// to WithBackgroundRoutines alongside JWTKeysInitializer.
func JWTKeysFetcher(cache *jwtkeys.Cache) BackgroundRoutine {
	return func(a App) {
		lg := a.Logger().Derive(log.WithFunction("JWTKeysFetcher"))
		defer a.BackgroundWaitGroup().Done()

		ctx := a.BackgroundContext()
		clnt := GetServiceClient[*authclient.Client](a, AuthServiceClientName)

		lg.Infoln("starting JWT public key refresh loop")
		cache.Run(ctx, clnt)
		lg.Infoln("stopping JWT public key refresh loop")
	}
}
