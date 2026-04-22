package app

import (
	"fmt"

	"github.com/swayrider/grpcclients"
	//"github.com/swayrider/swlib/svcreg"
)

// ServiceClientCtor is a constructor function type for creating gRPC service clients.
// It receives the App instance for accessing configuration values (host/port).
type ServiceClientCtor func(App) grpcclients.Client

// ServiceClient holds the configuration and instance of a gRPC service client.
// It manages the lifecycle of inter-service connections.
type ServiceClient struct {
	Name   string
	Ctor   ServiceClientCtor
	client grpcclients.Client
}

// NewServiceClient creates a new ServiceClient configuration.
// The constructor function will be called during application initialization.
//
// Example:
//
//	app.NewServiceClient("authservice", func(a app.App) grpcclients.Client {
//	    hostPort := app.ServiceClientHostAndPort(a, "authservice")
//	    host, port := hostPort()
//	    client, _ := authclient.New(host, port)
//	    return client
//	})
func NewServiceClient(
	name string,
	Ctor ServiceClientCtor,
) *ServiceClient {
	return &ServiceClient{
		Name: name,
		Ctor: Ctor,
	}
}

func (a *app) configureSericeClient(clnt *ServiceClient) {
	c := clnt.Ctor(a)
	if c == nil {
		panic(fmt.Sprintf("failed to create service client: %s", clnt.Name))
	}
	clnt.client = c
	a.serviceClients[clnt.Name] = clnt
}

func (a *app) closeServiceClients() {
	for _, clnt := range a.serviceClients {
		clnt.client.Close()
	}
}

// ServiceClientHostAndPort returns a function that retrieves the host and port
// for a service client. It first checks configuration values, then falls back
// to service discovery if available.
//
// Example:
//
//	hostPort := app.ServiceClientHostAndPort(a, "authservice")
//	host, port := hostPort()
//	client, _ := authclient.New(host, port)
func ServiceClientHostAndPort(a App, serviceName string, tags ...string) func() (string, int) {
	return func() (host string, port int) {
		host = GetConfigField[string](
			a.Config(), KeyServiceHost(serviceName))
		port = GetConfigField[int](
			a.Config(), KeyServicePort(serviceName))
		if host != "" && port != 0 {
			return
		}

		/*if a.ServiceResolver() != nil {
			consulExternal := GetConfigField[bool](
				a.Config(), KeyConsulExternal)

			serviceDesc, err := a.ServiceResolver().Get(
				svcreg.NewServiceQuery(serviceName, tags...))
			if err != nil {
				return
			}
			host = serviceDesc.ServiceHost(consulExternal)
			port = serviceDesc.ServicePort(consulExternal)
			return
		}*/
		return
	}
}
