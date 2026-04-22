package app

import log "github.com/swayrider/swlib/logger"

// StartHttpServerFn is a function type for starting a custom HTTP server.
// It receives the App instance and should return an error if startup fails.
// The function should start the server asynchronously (in a goroutine).
type StartHttpServerFn func(App) error

// StopHttpServerFn is a function type for stopping a custom HTTP server.
// It receives the App instance and should perform graceful shutdown.
type StopHttpServerFn func(App)

func (a *app) startHttpServer() {
	lg := a.lg.Derive(log.WithFunction("startHttpServer"))
	if a.httpStartFn == nil {
		return
	}

	if err := a.httpStartFn(a); err != nil {
		lg.Fatalf("Http servier failed to start: %v", err)
	}
}

func (a *app) stopHttpServer() {
	if a.httpStopFn != nil {
		a.httpStopFn(a)
	}
}

func (a* app) SetStaticHttpServer(s HttpServer) {
	a.httpStaticServer = s
}

func (a app) GetStaticHttpServer() HttpServer {
	return a.httpStaticServer
}

