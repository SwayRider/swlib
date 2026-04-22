// Package app provides a service bootstrap framework for building microservices.
//
// The app package is the core of swlib, offering a fluent builder pattern for
// configuring and running microservices. It handles service lifecycle management,
// configuration parsing, database connections, gRPC/HTTP servers, service client
// management, and graceful shutdown.
//
// # Basic Usage
//
// Create a new application using the builder pattern:
//
//	app.New("myservice").
//	    WithDefaultConfigFields(app.BackendServiceFields | app.DatabaseConnectionFields).
//	    WithServiceClients(
//	        app.NewServiceClient("authservice", authClientCtor),
//	    ).
//	    WithDatabase(dbCtor, dbBootstrap).
//	    WithGrpc(grpcConfig).
//	    Run()
//
// # Configuration
//
// Configuration can be provided via environment variables or CLI flags. Pre-defined
// field groups simplify common configurations:
//
//   - BackendServiceFields: HTTP and gRPC port configuration
//   - DatabaseConnectionFields: PostgreSQL connection settings
//   - WebServiceFields: Static web server settings
//   - ClientCredentialsFields: OAuth client credentials
//   - HTMXServiceFields: HTMX-specific settings
//
// # Lifecycle
//
// The Run() method starts the application and blocks until an interrupt signal
// (SIGINT or SIGTERM) is received. It handles:
//   - Configuration parsing
//   - Database initialization
//   - Running initializers
//   - Starting background routines
//   - Starting gRPC and HTTP servers
//   - Graceful shutdown
package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	//"time"

	"google.golang.org/grpc"
	"github.com/swayrider/grpcclients"
	//"github.com/swayrider/swlib/flag"
	log "github.com/swayrider/swlib/logger"
	"github.com/swayrider/swlib/svcreg"
)

// Callback is a function type used for initializers and bootstrap functions.
// It receives the App instance and returns an error if the operation fails.
type Callback func(App) error

// BackgroundRoutine is a function type for long-running background tasks.
// Background routines run in separate goroutines and should respect the
// application's context for graceful shutdown.
//
// Example:
//
//	func(a app.App) {
//	    ticker := time.NewTicker(5 * time.Minute)
//	    defer ticker.Stop()
//	    for {
//	        select {
//	        case <-a.BackgroundContext().Done():
//	            a.BackgroundWaitGroup().Done()
//	            return
//	        case <-ticker.C:
//	            // Perform periodic task
//	        }
//	    }
//	}
type BackgroundRoutine func(App)

// App is the main application interface for configuring and running a microservice.
// It provides a fluent builder API for configuration and access to runtime components.
type App interface {
	WithDefaultConfigFields(DefaultFlagGroup, FlagGroupOverrides) App
	WithServiceClients(...*ServiceClient) App
	WithConfigFields(...ConfigField) App
	//WithBackendServices(...string) App
	WithDatabase(DBCtor, Callback) App
	WithAppData(string, any) App
	WithInitializers(...Callback) App
	WithBackgroundRoutines(...BackgroundRoutine) App
	WithGrpc(*GrpcConfig) App
	WithHTTP(StartHttpServerFn, StopHttpServerFn) App

	Config() *Config
	Logger() *log.Logger
	Database() DB
	//ServiceResolver() *svcreg.Resolver
	BackgroundContext() context.Context
	BackgroundWaitGroup() *sync.WaitGroup
	ServiceClient(string) grpcclients.Client
	SetAppData(string, any)
	AppData(string) any

	//GetServiceEntry(svcreg.ServiceQuery) *svcreg.ServiceEntry
	//GetServiceField(svcreg.ServiceQuery, svcreg.ServiceMetaKey, FieldKey) any

	SetStaticHttpServer(HttpServer)
	GetStaticHttpServer() HttpServer

	Run()
}

// HttpServer is an interface for HTTP server implementations that can be
// registered with the application for lifecycle management.
type HttpServer interface {
	Server() *http.Server
}

// New creates a new application instance with the given service name.
// The service name is used for logging and identification purposes.
//
// Example:
//
//	app := app.New("authservice")
func New(
	serviceName string, // Name of the service
) App {
	a := new(app)
	a.serviceName = serviceName
	a.cfg = NewConfig()
	a.lg = log.New(
		log.WithComponent(serviceName),
		log.WithFunction("app.New"),
	)
	a.serviceClients = make(map[string]*ServiceClient)
	a.initializers = []Callback{}
	a.backendServices = []string{}
	a.appdataMux = &sync.RWMutex{}
	a.appdata = make(map[string]any)
	return a
}

type app struct {
	serviceName string
	cfg         *Config
	lg          *log.Logger
	//serviceClients  []string
	serviceClients  map[string]*ServiceClient
	initializers    []Callback
	backendServices []string

	appdataMux *sync.RWMutex
	appdata    map[string]any

	dbCtor      DBCtor
	dbBootstrap Callback

	backgroundRoutines []BackgroundRoutine
	backgroundContext  context.Context
	cancelFunc         context.CancelFunc
	wg                 *sync.WaitGroup

	grpcConfig *GrpcConfig

	httpStartFn StartHttpServerFn
	httpStopFn  StopHttpServerFn

	dbconn      DB
	svcResolver *svcreg.Resolver
	grpcServer  *grpc.Server
	httpGateway  *http.Server
	httpStaticServer HttpServer
}

func (a *app) WithDefaultConfigFields(
	groups DefaultFlagGroup,
	groupsDefaultOverrides FlagGroupOverrides,
) App {
	fields := make([]ConfigField, 0)
	/*if (groups & ConsulFields) == ConsulFields {
		fields = append(
			fields,
			defaultConsulFields(fields, groupsDefaultOverrides)...)
	}*/
	if (groups & BackendServiceFields) == BackendServiceFields {
		fields = append(
			fields,
			defaultBackendServiceFields(fields, groupsDefaultOverrides)...)
	}
	if (groups & HTMXServiceFields) == HTMXServiceFields {
		fields = append(
			fields,
			defaultHTMXServiceFields(fields, groupsDefaultOverrides)...)
	}
	if (groups & ClientCredentialsFields) == ClientCredentialsFields {
		fields = append(
			fields,
			defaultClientCredentialsFields(fields, groupsDefaultOverrides)...)
	}
	if (groups & DatabaseConnectionFields) == DatabaseConnectionFields {
		fields = append(
			fields,
			defaultDatabaseConnectionFields(fields, groupsDefaultOverrides)...)
	}
	if (groups & WebServiceFields) == WebServiceFields {
		fields = append(
			fields,
			defaultWebServiceFields(fields, groupsDefaultOverrides)...)
	}
	if (groups & LoggerFields) == LoggerFields {
		fields = append(
			fields,
			defaultLoggerFields(fields, groupsDefaultOverrides)...)
	}
	if len(fields) > 0 {
		a.cfg.AddFields(fields)
	}
	return a
}

func (a *app) WithServiceClients(
	serviceClients ...*ServiceClient,
) App {
	fields := make([]ConfigField, 0)
	for _, svc := range serviceClients {
		fields = append(fields, serviceClientFields(svc.Name)...)
	}
	if len(fields) > 0 {
		a.cfg.AddFields(fields)
	}

	for _, svc := range serviceClients {
		a.configureSericeClient(svc)
	}
	return a
}

func (a *app) WithConfigFields(
	fields ...ConfigField,
) App {
	a.cfg.AddFields(fields)
	return a
}

func (a *app) WithBackendServices(
	services ...string,
) App {
	a.backendServices = services
	return a
}

func (a *app) WithDatabase(dbCtor DBCtor, dbBootstrap Callback) App {
	a.dbCtor = dbCtor
	a.dbBootstrap = dbBootstrap
	return a
}

func (a *app) WithAppData(
	key string,
	value any,
) App {
	a.appdata[key] = value
	return a
}

func (a *app) WithInitializers(
	initializers ...Callback,
) App {
	a.initializers = initializers
	return a
}

func (a *app) WithBackgroundRoutines(routines ...BackgroundRoutine) App {
	a.backgroundRoutines = routines
	return a
}

func (a *app) WithGrpc(grpcConfig *GrpcConfig) App {
	a.grpcConfig = grpcConfig
	return a
}

func (a *app) WithHTTP(startFn StartHttpServerFn, stopFn StopHttpServerFn) App {
	a.httpStartFn = startFn
	a.httpStopFn = stopFn
	return a
}

func (a app) Config() *Config {
	return a.cfg
}

func (a app) Logger() *log.Logger {
	return a.lg
}

func (a app) Database() DB {
	return a.dbconn
}

/*func (a app) ServiceResolver() *svcreg.Resolver {
	return a.svcResolver
}*/

func (a app) BackgroundContext() context.Context {
	return a.backgroundContext
}

func (a app) BackgroundWaitGroup() *sync.WaitGroup {
	return a.wg
}

func (a app) ServiceClient(name string) grpcclients.Client {
	return a.serviceClients[name].client
}

/*func (a app) GetServiceEntry(query svcreg.ServiceQuery) *svcreg.ServiceEntry {
	if a.svcResolver != nil {
		entry, err := a.svcResolver.Get(query)
		if err == nil {
			return &entry
		}
	}
	return nil
}

func (a app) GetServiceField(
	query svcreg.ServiceQuery,
	key svcreg.ServiceMetaKey,
	fieldKey string,
) any {
	entry := a.GetServiceEntry(query)
	if entry != nil {
		if value, ok := entry.MetaData[key]; ok {
			return value
		}
		if a.cfg.Get(KeyConsulExternal).Value().(bool) {
			if value, ok := entry.MetaData["external_"+key]; ok {
				return value
			}
		} else {
			if value, ok := entry.MetaData["internal_"+key]; ok {
				return value
			}
		}
	}
	return GetConfigFieldAsString(a.cfg, fieldKey)
}*/

func (a *app) Run() {
	// Parse configuration
	if err := a.cfg.Parse(); err != nil {
		a.lg.Fatalf("Failed to parse config: %v", err)
	}

	// Configure logger level from config
	logLevel := GetConfigField[string](a.cfg, KeyLogLevel)
	if logLevel != "" {
		if err := log.SetLogLevel(logLevel); err != nil {
			a.lg.Warnf("invalid log level '%s', using default 'info': %v", logLevel, err)
		} else {
			a.lg.Debugf("log level set to: %s", logLevel)
		}
	}

	// Initialize interrupt handler
	ctx, stop := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a.backgroundContext, a.cancelFunc = context.WithCancel(context.Background())
	a.wg = new(sync.WaitGroup)

	/*consulHostsField := a.cfg.Get(KeyConsulHosts)
	if consulHostsField != nil {
		consulHosts, ok := consulHostsField.Value().(flag.StringArr)
		if ok && len(consulHosts) > 0 {
			consulInsecureField := a.cfg.Get(KeyConsulInsecure)
			consulInsecure := consulInsecureField.Value().(bool)

			svcList := make([]string, 0, len(a.serviceClients)+len(a.backendServices))
			for svc := range a.serviceClients {
				svcList = append(svcList, svc)
			}
			svcList = append(svcList, a.backendServices...)

			a.svcResolver = svcreg.NewResolver(
				consulHosts, consulInsecure, svcList, 5*time.Second, a.lg)
			a.svcResolver.Start()
		}
	}

	defer func() {
		a.closeServiceClients()
		if a.svcResolver != nil {
			a.svcResolver.Stop()
		}
	}()*/

	a.initDatabase()

	a.runInitializers()
	a.runBackgroundRoutines()

	a.startGrpc()
	a.startHttpServer()
	<-ctx.Done()

	a.lg.Infoln("Shutting down")

	a.stopGrpcServer()
	a.stopHttpServer()
	a.cancelFunc()
	a.wg.Wait()
	a.lg.Infoln("Shutdown complete")
}

func (a *app) runInitializers() {
	for _, routine := range a.initializers {
		routine(a)
	}
}

func (a *app) runBackgroundRoutines() {
	a.wg.Add(len(a.backgroundRoutines))
	for _, routine := range a.backgroundRoutines {
		go routine(a)
	}
}

// GetServiceClient retrieves a typed service client by name from the application.
// It uses Go generics to return the client with the correct type.
//
// Example:
//
//	authClient := app.GetServiceClient[*authclient.Client](a, "authservice")
func GetServiceClient[T grpcclients.Client](a App, name string) T {
	return a.ServiceClient(name).(T)
}

func (a app) SetAppData(key string, value any) {
	a.appdataMux.Lock()
	defer a.appdataMux.Unlock()
	a.appdata[key] = value
}

func (a app) AppData(key string) any {
	a.appdataMux.RLock()
	defer a.appdataMux.RUnlock()
	return a.appdata[key]
}

// SetAppData stores a typed value in the application's thread-safe data store.
// This is a generic helper that provides type safety when storing data.
//
// Example:
//
//	app.SetAppData(a, "cache-manager", cacheManager)
func SetAppData[T any](a App, key string, value T) {
	a.SetAppData(key, value)
}

// GetAppData retrieves a typed value from the application's thread-safe data store.
// It uses Go generics to return the value with the correct type.
//
// Example:
//
//	cm := app.GetAppData[*CacheManager](a, "cache-manager")
func GetAppData[T any](a App, key string) T {
	return a.AppData(key).(T)
}
