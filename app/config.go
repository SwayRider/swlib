package app

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application's configuration fields and provides methods
// for adding, retrieving, and parsing configuration values.
//
// Configuration values can be set via:
//   - Environment variables (highest priority when parsing)
//   - CLI flags
//   - Default values (lowest priority)
type Config struct {
	fields map[string]ConfigField
}

// NewConfig creates a new Config instance with the given initial fields.
// Fields can also be added later using AddFields.
func NewConfig(
	fields ...ConfigField,
) *Config {
	c := new(Config)
	c.fields = make(map[string]ConfigField)
	for _, f := range fields {
		c.fields[f.Name()] = f
	}
	return c
}

// AddFields adds new configuration fields to the Config.
// Fields that already exist (by name) are not overwritten.
func (c Config) AddFields(fields []ConfigField) {
	for _, f := range fields {
		if _, ok := c.fields[f.Name()]; !ok {
			c.fields[f.Name()] = f
		}
	}
}

// Get retrieves a configuration field by name.
// Returns nil if the field does not exist.
func (c Config) Get(name string) ConfigField {
	return c.fields[name]
}

// Parse parses configuration from environment variables and CLI flags.
// It loads .env files using godotenv, then processes all registered fields.
//
// The parsing order is:
//  1. Load .env file if present
//  2. Configure CLI flags with environment variable values as defaults
//  3. Parse CLI flags (overrides environment variables)
//  4. Set final values on all fields
func (c Config) Parse(flagSet ...*flag.FlagSet) (err error) {
	_ = godotenv.Load()

	for _, f := range c.fields {
		f.configureFlag()
	}

	args := []string{}
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}
	for _, fs := range flagSet {
		if fs != nil {
			fs.Parse(args)
		}
	}
	if len(flagSet) == 0 {
		flag.Parse()
	}

	for _, f := range c.fields {
		f.setFlagValue()
	}

	return nil
}

/*func defaultConsulFields(
	fields []ConfigField,
	groupsDefaultOverrides FlagGroupOverrides,
) []ConfigField {
	fields = append(fields, NewStringArrConfigField(
		KeyConsulHosts, EnvConsulHosts,
		"host and port numbers of the consul cluster (e.g. 127.0.0.1:8501,127.0.0.1:8502,127.0.0.1:8502)",
		groupFieldDefVal(
			KeyConsulHosts, groupsDefaultOverrides, DefConsulHosts),
	))
	fields = append(fields, NewBoolConfigField(
		KeyConsulExternal, EnvConsulExternal,
		"Service is running outside of consule network",
		groupFieldDefVal(
			KeyConsulExternal, groupsDefaultOverrides, DefConsulExternal),
	))
	fields = append(fields, NewBoolConfigField(
		KeyConsulInsecure, EnvConsulInsecure,
		"TLS certificates are not verified",
		groupFieldDefVal(
			KeyConsulInsecure, groupsDefaultOverrides, DefConsulInsecure),
	))
	return fields
}*/

func defaultBackendServiceFields(
	fields []ConfigField,
	groupsDefaultOverrides map[string]any,
) []ConfigField {
	fields = append(fields, NewIntConfigField(
		KeyHttpPort, EnvHttpPort,
		"Port of the grpc gateway server",
		groupFieldDefVal(
			KeyHttpPort, groupsDefaultOverrides, DefHttpPort),
	))
	fields = append(fields, NewIntConfigField(
		KeyGrpcPort, EnvGrpcPort,
		"Port of the gRPC server",
		groupFieldDefVal(
			KeyGrpcPort, groupsDefaultOverrides, DefGrpcPort),
	))
	return fields
}

func defaultHTMXServiceFields(
	fields []ConfigField,
	groupsDefaultOverrides map[string]any,
) []ConfigField {
	fields = append(fields, NewIntConfigField(
		KeyHttpPort, EnvHttpPort,
		"Port of the http server",
		groupFieldDefVal(
			KeyHttpPort, groupsDefaultOverrides, DefFrontentHttpPort),
	))
	fields = append(fields, NewStringConfigField(
		KeyServerName, EnvServerName,
		"Server name (e.g. https://example/com), used for redirects etc.",
		groupFieldDefVal(
			KeyServerName, groupsDefaultOverrides, DefServerName),
	))
	fields = append(fields, NewStringConfigField(
		KeyPathPrefix, EnvPathPrefix,
		"Path prefix (e.g. /my/subpath), used when running behind a proxy that modifies the path by prefixing it",
		groupFieldDefVal(
			KeyPathPrefix, groupsDefaultOverrides, DefPathPrefix),
	))
	return fields
}

func defaultClientCredentialsFields(
	fields []ConfigField,
	groupsDefaultOverrides map[string]any,
) []ConfigField {
	fields = append(fields, NewStringConfigField(
		KeyClientId, EnvClientId,
		"OAuth2 client id",
		groupFieldDefVal(
			KeyClientId, groupsDefaultOverrides, DefClientId),
	))
	fields = append(fields, NewStringConfigField(
		KeyClientSecret, EnvClientSecret,
		"OAuth2 client secret",
		groupFieldDefVal(
			KeyClientSecret, groupsDefaultOverrides, DefClientSecret),
	))
	return fields
}

func defaultDatabaseConnectionFields(
	fields []ConfigField,
	groupsDefaultOverrides map[string]any,
) []ConfigField {
	fields = append(fields, NewStringConfigField(
		KeyDBHost, EnvDBHost,
		"Database host (e.g. localhost)",
		groupFieldDefVal(
			KeyDBHost, groupsDefaultOverrides, DefDBHost),
	))
	fields = append(fields, NewIntConfigField(
		KeyDBPort, EnvDBPort,
		"Database port (e.g. 5432)",
		groupFieldDefVal(
			KeyDBPort, groupsDefaultOverrides, DefDBPort),
	))
	fields = append(fields, NewStringConfigField(
		KeyDBName, EnvDBName,
		"Database name (e.g. my-database)",
		groupFieldDefVal(
			KeyDBName, groupsDefaultOverrides, DefDBName),
	))
	fields = append(fields, NewStringConfigField(
		KeyDBUser, EnvDBUser,
		"Database user (e.g. admin)",
		groupFieldDefVal(
			KeyDBUser, groupsDefaultOverrides, DefDBUser),
	))
	fields = append(fields, NewStringConfigField(
		KeyDBPassword, EnvDBPassword,
		"Database password (e.g. abc123)",
		groupFieldDefVal(
			KeyDBPassword, groupsDefaultOverrides, DefDBPassword),
	))
	fields = append(fields, NewStringConfigField(
		KeyDBSSLMode, EnvDBSSLMode,
		"Database connection SSL model (e.g. disable)",
		groupFieldDefVal(
			KeyDBSSLMode, groupsDefaultOverrides, DefDBSSLMode),
	))
	return fields
}

func defaultWebServiceFields(
	fields []ConfigField,
	groupsDefaultOverrides map[string]any,
) []ConfigField {
	fields = append(fields, NewIntConfigField(
		KeyWebPort, EnvWebPort,
		"Port of the static web server",
		groupFieldDefVal(
			KeyWebPort, groupsDefaultOverrides, DefWebPort),
	))
	fields = append(fields, NewStringConfigField(
		KeyWebPathPrefix, EnvWebPathPrefix,
		"Path prefix (e.g. /my/subpath)",
		groupFieldDefVal(
			KeyWebPathPrefix, groupsDefaultOverrides, DefWebPathPrefix),
	))
	return fields
}

func defaultLoggerFields(
	fields []ConfigField,
	groupsDefaultOverrides map[string]any,
) []ConfigField {
	fields = append(fields, NewStringConfigField(
		KeyLogLevel, EnvLogLevel,
		"Log level: debug, info, warn, error (default: info)",
		groupFieldDefVal(
			KeyLogLevel, groupsDefaultOverrides, DefLogLevel),
	))
	return fields
}

func serviceClientFields(
	serviceName string,
) []ConfigField {
	hostFlag := KeyServiceHost(serviceName)
	portFlag := KeyServicePort(serviceName)
	hostEnv := EnvServiceHost(serviceName)
	portEnv := EnvServicePort(serviceName)

	fields := make([]ConfigField, 2)
	fields[0] = NewStringConfigField(
		hostFlag, hostEnv,
		fmt.Sprintf("%s host (e.g. localhost)", serviceName),
		"",
	)
	fields[1] = NewIntConfigField(
		portFlag, portEnv,
		fmt.Sprintf("%s port (e.g. 8081)", serviceName),
		0,
	)

	return fields
}

// GetConfigField retrieves a typed configuration value by name.
// It uses Go generics to return the value with the correct type.
//
// Example:
//
//	port := app.GetConfigField[int](cfg, "http-port")
//	hosts := app.GetConfigField[flag.StringArr](cfg, "allowed-hosts")
func GetConfigField[T ConfigFieldValue](c *Config, name string) T {
	return c.fields[name].Value().(T)
}

// GetConfigFieldAsString retrieves a configuration value as a string.
// It converts any value type to its string representation.
// Returns an empty string if the field does not exist or has no value.
func GetConfigFieldAsString(c *Config, name string) string {
	v := c.fields[name].Value()
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", c.fields[name].Value())
}

func groupFieldDefVal[T any](name string, groupsDefaultOverrides map[string]any, genericDefault T) T {
	if val, ok := groupsDefaultOverrides[name]; ok {
		return val.(T)
	}
	return genericDefault
}
