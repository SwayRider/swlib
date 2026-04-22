package app

import (
	"fmt"
	"strings"
)

// DefaultFlagGroup represents a bitmask for selecting pre-defined configuration
// field groups. Multiple groups can be combined using bitwise OR.
//
// Example:
//
//	app.New("myservice").
//	    WithDefaultConfigFields(
//	        app.BackendServiceFields | app.DatabaseConnectionFields,
//	        nil,
//	    )
type DefaultFlagGroup uint16

// FlagGroupOverrides allows overriding default values for pre-defined field groups.
// Keys are field names (e.g., KeyHttpPort), values are the new defaults.
type FlagGroupOverrides = map[string]any

// Pre-defined configuration field groups for common service configurations.
// These can be combined using bitwise OR when calling WithDefaultConfigFields.
const (
	// NoDefaultFields adds no default configuration fields
	NoDefaultFields DefaultFlagGroup = 0x0000
	//ConsulFields             DefaultFlagGroup = 0x0001

	// BackendServiceFields adds HTTP and gRPC port configuration and logger configuration
	// Fields: http-port (HTTP_PORT), grpc-port (GRPC_PORT), log-level (LOG_LEVEL)
	BackendServiceFields DefaultFlagGroup = 0x0002 | 0x0080

	// HTMXServiceFields adds configuration for HTMX-based web services
	// Fields: http-port (HTTP_PORT), server-name (SERVER_NAME), path-prefix (PATH_PREFIX)
	HTMXServiceFields DefaultFlagGroup = 0x0004

	// ClientCredentialsFields adds OAuth2 client credentials configuration
	// Fields: client-id (CLIENT_ID), client-secret (CLIENT_SECRET)
	ClientCredentialsFields DefaultFlagGroup = 0x0008

	// DatabaseConnectionFields adds PostgreSQL database connection configuration
	// Fields: db-host, db-port, db-name, db-user, db-password, db-ssl-mode
	DatabaseConnectionFields DefaultFlagGroup = 0x0010

	// WebServiceFields adds static web server configuration
	// Fields: web-port (WEB_PORT), web-path-prefix (WEB_PATH_PREFIX)
	WebServiceFields DefaultFlagGroup = 0x0040

	// LoggerFields adds logger configuration
	// Fields: log-level (LOG_LEVEL)
	LoggerFields DefaultFlagGroup = 0x0080
)

// FieldKey is a type alias for CLI flag names (e.g., "http-port").
type FieldKey = string

// FieldEnv is a type alias for environment variable names (e.g., "HTTP_PORT").
type FieldEnv = string

// Configuration field keys for CLI flags.
// These constants define the flag names used when configuring the application.
const (
	//KeyConsulHosts    FieldKey = "consul-hosts"
	//KeyConsulExternal FieldKey = "consul-external"
	//KeyConsulInsecure FieldKey = "consul-insecure"
	KeyHttpPort       FieldKey = "http-port"
	KeyGrpcPort       FieldKey = "grpc-port"
	KeyServerName     FieldKey = "server-name"
	KeyPathPrefix     FieldKey = "path-prefix"
	KeyClientId       FieldKey = "client-id"
	KeyClientSecret   FieldKey = "client-secret"
	KeyDBHost         FieldKey = "db-host"
	KeyDBPort         FieldKey = "db-port"
	KeyDBName         FieldKey = "db-name"
	KeyDBUser         FieldKey = "db-user"
	KeyDBPassword     FieldKey = "db-password"
	KeyDBSSLMode      FieldKey = "db-ssl-mode"
	KeyWebPort        FieldKey = "web-port"
	KeyWebPathPrefix  FieldKey = "web-path-prefix"
	KeyLogLevel       FieldKey = "log-level"
)

const (
	//EnvConsulHosts    FieldEnv = "CONSUL_HOSTS"
	//EnvConsulExternal FieldEnv = "CONSUL_EXTERNAL"
	//EnvConsulInsecure FieldEnv = "CONSUL_INSECURE"
	EnvHttpPort       FieldEnv = "HTTP_PORT"
	EnvGrpcPort       FieldEnv = "GRPC_PORT"
	EnvServerName     FieldEnv = "SERVER_NAME"
	EnvPathPrefix     FieldEnv = "PATH_PREFIX"
	EnvClientId       FieldEnv = "CLIENT_ID"
	EnvClientSecret   FieldEnv = "CLIENT_SECRET"
	EnvDBHost         FieldEnv = "DB_HOST"
	EnvDBPort         FieldEnv = "DB_PORT"
	EnvDBName         FieldEnv = "DB_NAME"
	EnvDBUser         FieldEnv = "DB_USER"
	EnvDBPassword     FieldEnv = "DB_PASSWORD"
	EnvDBSSLMode      FieldEnv = "DB_SSL_MODE"
	EnvWebPort        FieldEnv = "WEB_PORT"
	EnvWebPathPrefix  FieldEnv = "WEB_PATH_PREFIX"
	EnvLogLevel       FieldEnv = "LOG_LEVEL"
)

var (
	//DefConsulHosts      = []string{"127.0.0.1:8500"}
	//DefConsulExternal   = false
	//DefConsulInsecure   = false
	DefHttpPort         = 8080
	DefFrontentHttpPort = 9000
	DefGrpcPort         = 8081
	DefServerName       = "http://localhost"
	DefPathPrefix       = ""
	DefClientId         = ""
	DefClientSecret     = ""
	DefDBHost           = ""
	DefDBPort           = 0
	DefDBName           = ""
	DefDBUser           = ""
	DefDBPassword       = ""
	DefDBSSLMode        = "disable"
	DefWebPort          = 8000
	DefWebPathPrefix    = "/web"
	DefLogLevel         = "info"
)

// KeyServiceHost returns the CLI flag name for a service's host configuration.
// Example: KeyServiceHost("authservice") returns "authservice-host"
func KeyServiceHost(serviceName string) string {
	return fmt.Sprintf("%s-host", strings.ToLower(serviceName))
}

// KeyServicePort returns the CLI flag name for a service's port configuration.
// Example: KeyServicePort("authservice") returns "authservice-port"
func KeyServicePort(serviceName string) string {
	return fmt.Sprintf("%s-port", strings.ToLower(serviceName))
}

// EnvServiceHost returns the environment variable name for a service's host configuration.
// Example: EnvServiceHost("authservice") returns "AUTHSERVICE_HOST"
func EnvServiceHost(serviceName string) string {
	return fmt.Sprintf("%s_HOST", strings.ToUpper(serviceName))
}

// EnvServicePort returns the environment variable name for a service's port configuration.
// Example: EnvServicePort("authservice") returns "AUTHSERVICE_PORT"
func EnvServicePort(serviceName string) string {
	return fmt.Sprintf("%s_PORT", strings.ToUpper(serviceName))
}
