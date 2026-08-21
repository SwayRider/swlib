# SwLib - SwayRider Shared Library

SwLib is a comprehensive Go library providing reusable components for building microservices in the SwayRider platform. It offers a complete toolkit for service lifecycle management, security, logging, configuration, and common utilities.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Packages](#packages)
  - [app - Service Bootstrap Framework](#app---service-bootstrap-framework)
  - [cache - In-Memory Caching](#cache---in-memory-caching)
  - [compression - Data Compression](#compression---data-compression)
  - [crypto - Cryptographic Utilities](#crypto---cryptographic-utilities)
  - [env - Environment Variables](#env---environment-variables)
  - [flag - CLI Flag Parsing](#flag---cli-flag-parsing)
  - [grpc - gRPC Utilities](#grpc---grpc-utilities)
  - [hibp - Pwned Passwords Breach Checks](#hibp---pwned-passwords-breach-checks)
  - [http - HTTP Middleware](#http---http-middleware)
  - [jwt - JWT Token Management](#jwt---jwt-token-management)
  - [jwtkeys - JWT Public Key Cache](#jwtkeys---jwt-public-key-cache)
  - [logger - Structured Logging](#logger---structured-logging)
  - [math - Mathematical Utilities](#math---mathematical-utilities)
  - [ratelimit - Request Rate Limiting](#ratelimit---request-rate-limiting)
  - [security - Authorization Framework](#security---authorization-framework)
  - [str - String Utilities](#str---string-utilities)
- [Architecture Overview](#architecture-overview)
- [Best Practices](#best-practices)

## Installation

```bash
go get github.com/swayrider/swlib
```

## Quick Start

Here's a minimal example of an HTTP-only microservice using swlib (see [HTTP-only Services](#http-only-services) below; for a gRPC service see [gRPC and HTTP Gateway](#grpc-and-http-gateway)):

```go
package main

import (
    "github.com/swayrider/swlib/app"
)

func main() {
    application := app.New("myservice").
        WithDefaultConfigFields(app.BackendServiceFields, app.FlagGroupOverrides{}).
        WithHTTP(startHTTPServer, stopHTTPServer).
        Run()
    _ = application
}

func startHTTPServer(a app.App) error {
    // Start your HTTP server here
    return nil
}

func stopHTTPServer(a app.App) {
    // Gracefully stop your HTTP server here
}
```

---

## Packages

### app - Service Bootstrap Framework

The `app` package is the core of swlib, providing a fluent builder pattern for configuring and running microservices. `app.New(...)` returns an `App` interface value; it handles service lifecycle, configuration, database connections, gRPC/HTTP servers, and graceful shutdown.

#### Basic Usage

```go
package main

import (
    "github.com/swayrider/swlib/app"
)

func main() {
    application := app.New("myservice").
        WithDefaultConfigFields(app.BackendServiceFields|app.DatabaseConnectionFields, app.FlagGroupOverrides{}).
        Run()
    _ = application
}
```

#### Configuration Field Groups

The app package provides pre-defined configuration field groups that can be combined using bitwise OR:

| Field Group | Description | Environment Variables |
|-------------|-------------|----------------------|
| `BackendServiceFields` | Basic service configuration | `LOG_LEVEL`, `HTTP_PORT`, `GRPC_PORT` |
| `DatabaseConnectionFields` | PostgreSQL connection | `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` |
| `WebServiceFields` | Web server configuration | `WEB_PORT`, `WEB_PATH_PREFIX` |
| `ClientCredentialsFields` | OAuth client credentials | `CLIENT_ID`, `CLIENT_SECRET` |
| `HTMXServiceFields` | HTMX-specific configuration | Various HTMX settings |

`WithDefaultConfigFields` takes a second `FlagGroupOverrides` argument used to override individual field defaults/env names per group; pass `app.FlagGroupOverrides{}` when no overrides are needed.

#### Custom Configuration Fields

```go
func main() {
    application := app.New("myservice").
        WithDefaultConfigFields(app.BackendServiceFields, app.FlagGroupOverrides{}).
        WithConfigFields(
            app.NewStringConfigField(
                "api-key", "API_KEY", "External API key", ""),
            app.NewIntConfigField(
                "cache-ttl", "CACHE_TTL", "Cache time-to-live in seconds", 300),
        ).
        Run()

    // Access configuration values
    apiKey := app.GetConfigField[string](application.Config(), "api-key")
    cacheTTL := app.GetConfigField[int](application.Config(), "cache-ttl")
    _, _ = apiKey, cacheTTL
}
```

#### Database Integration

```go
import (
    "database/sql"
    _ "github.com/lib/pq"

    "github.com/swayrider/swlib/app"
)

// pgDB is a minimal app.DB implementation wrapping *sql.DB.
type pgDB struct{ conn *sql.DB }

func (d *pgDB) SqlDB() *sql.DB { return d.conn }

func main() {
    application := app.New("myservice").
        WithDefaultConfigFields(app.BackendServiceFields|app.DatabaseConnectionFields, app.FlagGroupOverrides{}).
        WithDatabase(
            // Database constructor
            func(a app.App) app.DB {
                connStr := fmt.Sprintf(
                    "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
                    app.GetConfigField[string](a.Config(), app.KeyDBHost),
                    app.GetConfigField[int](a.Config(), app.KeyDBPort),
                    app.GetConfigField[string](a.Config(), app.KeyDBUser),
                    app.GetConfigField[string](a.Config(), app.KeyDBPassword),
                    app.GetConfigField[string](a.Config(), app.KeyDBName),
                )
                conn, err := sql.Open("postgres", connStr)
                if err != nil {
                    a.Logger().Fatal(err.Error())
                }
                return &pgDB{conn: conn}
            },
            // Bootstrap function (optional migrations, etc.)
            func(a app.App) error {
                // Run migrations or initial setup
                return nil
            },
        ).
        Run()

    // Access database
    sqlDB := application.Database().SqlDB()
    _ = sqlDB
}
```

#### gRPC and HTTP Gateway

```go
import (
    "context"

    "google.golang.org/grpc"
    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

    "github.com/swayrider/swlib/app"
)

func main() {
    grpcConfig := app.NewGrpcConfig(
        app.AuthInterceptor|app.ClientInfoInterceptor,
        nil, // JWT public-keys fn — see the jwtkeys package if the service validates tokens
        app.GrpcServiceHooks{
            ServiceRegistrar:   registerMyService,
            ServiceHTTPHandler: myServiceGateway(nil), // pass the application instance here
        },
    )

    application := app.New("myservice").
        WithDefaultConfigFields(app.BackendServiceFields, app.FlagGroupOverrides{}).
        WithGrpc(grpcConfig).
        Run()
    _ = application
}

// registerMyService registers the gRPC service implementation with the server.
func registerMyService(r grpc.ServiceRegistrar, a app.App) {
    pb.RegisterMyServiceServer(r, &MyServiceServer{app: a})
}

// myServiceGateway returns an HTTP handler that proxies REST requests to gRPC.
func myServiceGateway(a app.App) app.ServiceHTTPHandler {
    return func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
        return pb.RegisterMyServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
    }
}
```

#### HTTP-only Services

Services that expose a plain HTTP API with no gRPC server at all (e.g. tilesservice) use `WithHTTP` instead of `WithGrpc`:

```go
func main() {
    application := app.New("myservice").
        WithDefaultConfigFields(app.BackendServiceFields, app.FlagGroupOverrides{}).
        WithHTTP(startHTTPServer, stopHTTPServer).
        Run()
    _ = application
}

func startHTTPServer(a app.App) error {
    // Start and own the http.Server; return once it's listening
    return nil
}

func stopHTTPServer(a app.App) {
    // Gracefully shut the server down
}
```

#### Service Clients

For inter-service communication, use the service client pattern:

```go
import (
    "github.com/swayrider/grpcclients"
    "github.com/swayrider/grpcclients/authclient"
    "github.com/swayrider/swlib/app"
)

func authServiceClientCtor(a app.App) grpcclients.Client {
    clnt, err := authclient.New(app.ServiceClientHostAndPort(a, "authservice"))
    if err != nil {
        a.Logger().Fatal(err.Error())
    }
    return clnt
}

func main() {
    application := app.New("myservice").
        WithDefaultConfigFields(app.BackendServiceFields, app.FlagGroupOverrides{}).
        WithServiceClients(
            app.NewServiceClient("authservice", authServiceClientCtor),
        ).
        Run()

    // Access service client
    authClient := app.GetServiceClient[*authclient.Client](application, "authservice")
    _ = authClient
}
```

#### Background Routines

```go
func main() {
    application := app.New("myservice").
        WithDefaultConfigFields(app.BackendServiceFields, app.FlagGroupOverrides{}).
        WithBackgroundRoutines(cleanupRoutine).
        Run()
    _ = application
}

func cleanupRoutine(a app.App) {
    ctx := a.BackgroundContext()
    defer a.BackgroundWaitGroup().Done()

    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // Periodic task
            cleanupExpiredSessions(a)
        }
    }
}
```

#### Application Data Store

Thread-safe key-value store for sharing data across components:

```go
func main() {
    application := app.New("myservice").
        WithDefaultConfigFields(app.BackendServiceFields, app.FlagGroupOverrides{}).
        Run()

    // Store data
    application.SetAppData("cache-manager", cacheManager)

    // Retrieve data
    cm := app.GetAppData[*CacheManager](application, "cache-manager")
    _ = cm
}
```

---

### cache - In-Memory Caching

Simple thread-safe local cache for storing arbitrary values.

#### Usage

```go
import "github.com/swayrider/swlib/cache"

// Define cache keys
const (
    UserCacheKey    cache.LocalCacheKey = "user"
    SessionCacheKey cache.LocalCacheKey = "session"
)

func main() {
    // Store a value
    cache.LCSet(UserCacheKey, &User{ID: "123", Name: "John"})

    // Retrieve a value
    if value, ok := cache.LCGet(UserCacheKey); ok {
        user := value.(*User)
        fmt.Println(user.Name)
    }

    // Check existence
    if cache.LCHas(UserCacheKey) {
        // Key exists
    }

    // Delete a value
    cache.LCDel(UserCacheKey)
}
```

---

### compression - Data Compression

Utilities for working with compressed files.

#### Extracting ZIP Archives

```go
import "github.com/swayrider/swlib/compression"

func extractData() error {
    err := compression.UnZip("/path/to/archive.zip", "/path/to/destination")
    if err != nil {
        return fmt.Errorf("failed to extract archive: %w", err)
    }
    return nil
}
```

---

### crypto - Cryptographic Utilities

Secure password hashing, random string generation, and RSA keypair management.

#### Password Hashing

Uses Argon2id, the winner of the Password Hashing Competition:

```go
import "github.com/swayrider/swlib/crypto"

func registerUser(password string) (string, error) {
    // Hash password for storage
    hash, err := crypto.CalculatePasswordHash(password)
    if err != nil {
        return "", fmt.Errorf("failed to hash password: %w", err)
    }
    return hash, nil
}

func authenticateUser(storedHash, password string) bool {
    // Verify password against stored hash
    return crypto.VerifyPassword(storedHash, password)
}
```

#### Secure Random Strings

```go
import "github.com/swayrider/swlib/crypto"

func generateToken() (string, error) {
    // Generate a 32-character cryptographically secure random string
    token, err := crypto.GenerateSecureRandomString(32)
    if err != nil {
        return "", err
    }
    return token, nil
}
```

#### RSA Keypair Generation

```go
import "github.com/swayrider/swlib/crypto"

func rotateKeys() error {
    privateKeyPEM, publicKeyPEM, expiresAt, err := crypto.CreateKeypair()
    if err != nil {
        return fmt.Errorf("failed to create keypair: %w", err)
    }

    // Store keys securely
    // privateKeyPEM: PEM-encoded private key
    // publicKeyPEM: PEM-encoded public key
    // expiresAt: Recommended expiration (30 days from now)

    return nil
}
```

---

### env - Environment Variables

Simplified environment variable access with type conversion and fallbacks.

#### Usage

```go
import "github.com/swayrider/swlib/env"

func loadConfig() {
    // String with fallback
    apiHost := env.Get("API_HOST", "localhost")

    // Integer with fallback
    port := env.GetAsInt("PORT", 8080)

    // Float with fallback
    timeout := env.GetAsFloat64("TIMEOUT_SECONDS", 30.0)

    // Boolean with fallback
    debug := env.GetAsBool("DEBUG", false)

    // String array (comma-separated)
    allowedOrigins := env.GetAsStringArr("ALLOWED_ORIGINS", []string{"http://localhost:3000"})

    // Integer array
    retryDelays := env.GetAsIntArr("RETRY_DELAYS", []int{1, 2, 5, 10})
}
```

---

### flag - CLI Flag Parsing

Custom flag types for parsing array values from command-line arguments.

#### String Array Flags

```go
import (
    "flag"
    "github.com/swayrider/swlib/swflag"
)

func main() {
    var hosts swflag.StringArr

    flag.Var(&hosts, "host", "Host addresses (comma-separated or multiple flags)")
    flag.Parse()

    // Usage:
    // ./myapp -host "host1,host2,host3"
    // ./myapp -host host1 -host host2 -host host3

    for _, host := range hosts {
        fmt.Println("Host:", host)
    }
}
```

#### With Custom FlagSet

```go
import "github.com/swayrider/swlib/swflag"

func parseFlags() {
    fs := flag.NewFlagSet("myapp", flag.ExitOnError)

    // Get parser function
    parseStringArr := swflag.StringArrayParser()

    // Define and parse flags
    hosts := parseStringArr(fs, "hosts", "Host addresses")

    fs.Parse(os.Args[1:])

    for _, host := range *hosts {
        fmt.Println(host)
    }
}
```

#### Other Array Types

```go
import "github.com/swayrider/swlib/swflag"

func main() {
    var ports swflag.IntArr
    var weights swflag.Float64Arr
    var features swflag.BoolArr

    flag.Var(&ports, "port", "Port numbers")
    flag.Var(&weights, "weight", "Weight values")
    flag.Var(&features, "feature", "Feature flags")
    flag.Parse()
}
```

---

### grpc - gRPC Utilities

Interceptors and helpers for gRPC service communication.

#### Authentication Interceptor

```go
import (
    "google.golang.org/grpc"
    "github.com/swayrider/swlib/grpc/interceptors"
)

func setupGrpcServer() *grpc.Server {
    // Get public keys function
    getPublicKeys := func() ([]string, error) {
        // Return list of valid public keys for JWT verification
        return []string{publicKeyPEM}, nil
    }

    server := grpc.NewServer(
        grpc.UnaryInterceptor(interceptors.UnaryAuthInterceptor(getPublicKeys)),
        grpc.StreamInterceptor(interceptors.StreamAuthInterceptor(getPublicKeys)),
    )

    return server
}
```

#### Client Info Interceptor

Extracts and propagates client information through the request chain:

```go
import "github.com/swayrider/swlib/grpc/interceptors"

func setupGrpcServer() *grpc.Server {
    server := grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            interceptors.ClientInfoInterceptor(),
            // ... other interceptors
        ),
    )
    return server
}
```

#### Service-to-Service Calls with Auto-Retry

```go
import "github.com/swayrider/swlib/grpc/s2s"

func callOtherService(ctx context.Context, authClient *authclient.Client) (*pb.Response, error) {
    // Automatically retries with fresh token on Unauthenticated error
    response, err := s2s.Call(ctx, func(ctx context.Context) (*pb.Response, error) {
        return authClient.SomeMethod(ctx, &pb.Request{})
    }, func() error {
        // Token refresh function
        return authClient.RefreshToken()
    })

    return response, err
}
```

---

### hibp - Pwned Passwords Breach Checks

Privacy-preserving client for the Have I Been Pwned [Pwned Passwords](https://haveibeenpwned.com/Passwords) range API. Used by authservice to reject passwords that have appeared in a known data breach.

```go
import "github.com/swayrider/swlib/hibp"

// enabled=true, 3s timeout, reject passwords seen at least once in breach data
client := hibp.New(true, 3*time.Second, 1, l)

breached, count, err := client.IsBreached(ctx, password)
```

Uses the **k-anonymity range protocol**: only the first 5 characters of the uppercase SHA-1 hash are sent to the API, so the password (and its full hash) never leave the server. The API is free and requires no API key. Any API error (timeout, rate limit, non-200) is returned to the caller, which is expected to **fail open** — an HIBP outage must never block users. `hibp.New(false, ...)` short-circuits every check without a network call.

---

### http - HTTP Middleware

HTTP middleware components for authentication, content validation, and request context.

#### JWT Authentication Middleware

```go
import (
    "net/http"
    "github.com/swayrider/swlib/http/middlewares"
)

func setupRoutes() http.Handler {
    mux := http.NewServeMux()

    // Get public keys function
    getPublicKeys := func() ([]string, error) {
        return []string{publicKeyPEM}, nil
    }

    // Protected route
    mux.Handle("/api/protected", middlewares.Auth(getPublicKeys)(protectedHandler))

    return mux
}
```

#### Web Authentication with Redirects

For web applications that need to redirect to login pages:

```go
import "github.com/swayrider/swlib/http/middlewares"

func setupWebRoutes() http.Handler {
    mux := http.NewServeMux()

    getPublicKeys := func() ([]string, error) {
        return []string{publicKeyPEM}, nil
    }

    // Redirects to /login if not authenticated
    // Redirects to /verify if email not verified
    mux.Handle("/dashboard", middlewares.WebAuth(getPublicKeys)(dashboardHandler))

    return mux
}
```

#### Content-Type Validation

```go
import "github.com/swayrider/swlib/http/middlewares"

func setupRoutes() http.Handler {
    mux := http.NewServeMux()

    // Only allow JSON content
    jsonOnly := middlewares.RequireMimeType("application/json")
    mux.Handle("/api/data", jsonOnly(dataHandler))

    return mux
}
```

#### Source Info Middleware

Extracts client IP, user agent, and other request metadata:

```go
import "github.com/swayrider/swlib/http/middlewares"

func setupRoutes() http.Handler {
    mux := http.NewServeMux()

    // Adds client info to request context
    mux.Handle("/api/", middlewares.SourceInfo()(apiHandler))

    return mux
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
    // Access source info from context
    clientIP := security.GetOrigIp(r.Context())
    userAgent := security.GetUserAgent(r.Context())
}
```

#### Secure File Serving

Prevents directory traversal attacks when serving static files:

```go
import "github.com/swayrider/swlib/http/middlewares"

func setupStaticFiles() http.Handler {
    fs := http.Dir("./static")
    fileServer := http.FileServer(middlewares.NeuterFS{Fs: fs})
    return fileServer
}
```

#### Cookie Utilities

```go
import "github.com/swayrider/swlib/http/cookies"

func setUserCookie(w http.ResponseWriter, userData map[string]string) error {
    // Encode data into cookie
    encoded, err := cookies.Encode("user_data", userData)
    if err != nil {
        return err
    }

    http.SetCookie(w, &http.Cookie{
        Name:     "user_data",
        Value:    encoded,
        HttpOnly: true,
        Secure:   true,
    })
    return nil
}

func getUserCookie(r *http.Request) (map[string]string, error) {
    cookie, err := r.Cookie("user_data")
    if err != nil {
        return nil, err
    }

    var userData map[string]string
    err = cookies.Decode("user_data", cookie.Value, &userData)
    return userData, err
}
```

#### HTTP Gzip Compression

Helpers for gzip-encoding HTTP responses (distinct from the top-level `compression` package, which only extracts ZIP archives):

```go
import "github.com/swayrider/swlib/http/compression"

func writeResponse(w http.ResponseWriter, r *http.Request, body []byte) error {
    if !compression.SupportsGzip(r) {
        _, err := w.Write(body)
        return err
    }

    compressed, err := compression.CompressGzip(body, gzip.DefaultCompression)
    if err != nil {
        return err
    }
    w.Header().Set("Content-Encoding", "gzip")
    _, err = w.Write(compressed)
    return err
}
```

---

### jwt - JWT Token Management

Comprehensive JWT handling with custom claims for user and service authentication.

#### Configuration

```go
import "github.com/swayrider/swlib/jwt"

func init() {
    // Configure JWT issuer and audience (typically done once at startup)
    jwt.Configure("https://auth.swayrider.com", "swayrider-services")
}
```

#### Generating Tokens

```go
import "github.com/swayrider/swlib/jwt"

func createUserToken(userID string, email string, isAdmin bool) (*jwt.AccessToken, error) {
    // OpenID Connect claims
    openIDClaims := &jwt.OpenIDClaims{
        Email:         email,
        EmailVerified: true,
        Name:          "John Doe",
    }

    // SwayRider-specific claims
    swayRiderClaims := &jwt.SwayRiderUserClaims{
        IsAdmin:      isAdmin,
        AccountLevel: "premium",
    }

    // Generate token with 1-hour TTL
    tokenID, accessToken, expiresAt, err := jwt.GenerateToken(
        userID,
        openIDClaims,
        swayRiderClaims,
        privateKeyPEM,
        time.Hour,
    )
    if err != nil {
        return nil, err
    }

    return accessToken, nil
}
```

#### Generating Service Client Tokens

```go
import "github.com/swayrider/swlib/jwt"

func createServiceToken(clientID string, scopes []string) (*jwt.AccessToken, error) {
    serviceClaims := &jwt.SwayRiderServiceClaims{
        Scopes: scopes,
    }

    _, accessToken, _, err := jwt.GenerateToken(
        clientID,
        nil, // No OpenID claims for service clients
        serviceClaims,
        privateKeyPEM,
        24*time.Hour,
    )

    return accessToken, err
}
```

#### Verifying Tokens

```go
import "github.com/swayrider/swlib/jwt"

func verifyUserToken(tokenString string) (*jwt.Claims, error) {
    claims, err := jwt.VerifyToken(tokenString, publicKeyPEM, jwt.VerifyOptions{
        ValidateExpiration: true,
    })
    if err != nil {
        return nil, fmt.Errorf("invalid token: %w", err)
    }

    // Access claims
    fmt.Println("User ID:", claims.Subject)
    fmt.Println("Email:", claims.OpenID.Email)
    fmt.Println("Is Admin:", claims.SwayRider.User.IsAdmin)

    return claims, nil
}
```

#### Working with Claims

```go
import "github.com/swayrider/swlib/jwt"

// Serialize claims to map
func claimsToMap(claims *jwt.Claims) map[string]interface{} {
    return claims.MapClaims()
}

// Deserialize claims from map
func mapToClaims(m map[string]interface{}) (*jwt.OpenIDClaims, error) {
    var claims jwt.OpenIDClaims
    err := claims.FromMapClaims(m)
    return &claims, err
}
```

---

### jwtkeys - JWT Public Key Cache

A hardened, configurable cache of the JWT verification public keys published by authservice. It refreshes in the background on a configurable interval, retains the last-known-good keys if a refresh fails, and bounds each fetch with a timeout. This replaces the old `authclient.PublicKeyFetcher` helper (removed — see `grpcclients`' README).

`app` provides ready-made glue (`JWTKeysConfigFields`, `JWTKeysInitializer`, `JWTKeysFetcher`) so most services never call `jwtkeys` directly:

```go
import (
    "github.com/swayrider/swlib/app"
    "github.com/swayrider/swlib/jwtkeys"
)

func main() {
    jwtKeyCache := jwtkeys.New(nil) // pass the service's *log.Logger

    grpcConfig := app.NewGrpcConfig(
        app.AuthInterceptor,
        jwtKeyCache.GetPublicKeys,
        // ...service hooks
    )

    application := app.New("myservice").
        WithDefaultConfigFields(app.BackendServiceFields, app.FlagGroupOverrides{}).
        WithServiceClients(app.NewServiceClient("authservice", authServiceClientCtor)).
        WithConfigFields(app.JWTKeysConfigFields()...).
        WithInitializers(app.JWTKeysInitializer(jwtKeyCache)).
        WithBackgroundRoutines(app.JWTKeysFetcher(jwtKeyCache)).
        WithGrpc(grpcConfig).
        Run()
    _ = application
}
```

| Environment Variable | Default | Description |
|---|---|---|
| `JWT_KEYS_REFRESH_INTERVAL_SECS` | `300` | How often the cache refreshes public keys from authservice |
| `JWT_KEYS_FETCH_TIMEOUT_SECS` | `15` | Timeout for a single public-key fetch |

`jwtKeyCache.Verify(token)` verifies a token directly against the cached keys; `jwtKeyCache.Keys()` returns the current key set.

---

### logger - Structured Logging

Context-aware logging with component and function tracking.

#### Basic Usage

```go
import "github.com/swayrider/swlib/logger"

func main() {
    // Package-level logging
    logger.Infof("Application starting on port %d", 8080)
    logger.Debugf("Debug mode enabled")
    logger.Warnf("Cache size approaching limit: %d%%", 85)
    logger.Errorf("Failed to connect to database: %v", err)
    logger.Successf("Migration completed successfully")
}
```

#### Component-Scoped Logging

```go
import "github.com/swayrider/swlib/logger"

type UserService struct {
    log *logger.Logger
}

func NewUserService() *UserService {
    return &UserService{
        log: logger.New("user-service"),
    }
}

func (s *UserService) CreateUser(email string) error {
    // Create function-scoped logger
    log := s.log.Derive("CreateUser")

    log.Infof("Creating user with email: %s", email)

    // ... user creation logic

    log.Successf("User created successfully")
    return nil
}
```

#### Log Levels

| Level | Method | Icon | Use Case |
|-------|--------|------|----------|
| Info | `Infof`, `Infoln` | info | General information |
| Debug | `Debugf`, `Debugln` | bug | Debugging details |
| Warn | `Warnf`, `Warnln` | warning | Warnings |
| Error | `Errorf`, `Errorln` | x | Errors |
| Fatal | `Fatalf`, `Fatalln` | x | Fatal errors (exits program) |
| Success | `Successf`, `Successln` | check | Success messages |

---

### math - Mathematical Utilities

Geometric calculations and floating-point utilities.

#### Zoom Level to Radius Conversion

```go
import "github.com/swayrider/swlib/math/geo"

func calculateSearchArea(zoomLevel int) {
    // Get search radius in meters
    radiusMeters := geo.Zoom2Radius(zoomLevel)

    // Get search radius in kilometers
    radiusKm := geo.Zoom2RadiusKm(zoomLevel)

    fmt.Printf("Zoom %d: %.0fm / %.2fkm radius\n", zoomLevel, radiusMeters, radiusKm)
}

// Zoom level reference:
// Zoom 0:  20000km (world view)
// Zoom 5:  1000km  (country)
// Zoom 10: 50km    (city)
// Zoom 15: 1.5km   (neighborhood)
// Zoom 18: 200m    (street)
```

#### Float Comparisons

```go
import "github.com/swayrider/swlib/math/floats"

func compareFloats() {
    a := 0.1 + 0.2
    b := 0.3

    // Direct comparison fails due to floating-point precision
    fmt.Println(a == b) // false

    // Use epsilon comparison
    fmt.Println(floats.Equal(a, b)) // true

    // Round to specific precision
    rounded := floats.Round(3.14159, 2) // 3.14
}
```

---

### ratelimit - Request Rate Limiting

A per-key token-bucket rate limiter, used to apply per-source-IP request limits at the HTTP gateway (and, as a raw-port fallback, at the gRPC layer via `RateLimitInterceptor`).

```go
import "github.com/swayrider/swlib/ratelimit"

// requestsPerSecond=50, burst=100, idle entries evicted after 5 minutes
limiter := ratelimit.New(50, 100, 5*time.Minute)

if !limiter.Allow(sourceIP) {
    // reject with 429 / ResourceExhausted
}

// Periodically evict idle per-key entries
limiter.Evict()
```

`app` wires this up automatically for services that opt in via `app.RateLimitConfigFields()`, `app.RateLimitInterceptor`, and `app.RateLimitEvictor` (see the `RATE_LIMIT_*` env vars documented in each service's README).

---

### security - Authorization Framework

Endpoint-based authorization and JWT context extraction.

#### Defining Endpoint Profiles

```go
import "github.com/swayrider/swlib/security"

func init() {
    // Public endpoint - no authentication required
    security.SetEndpointProfile("/api/health", security.PublicEndpoint())

    // Protected endpoint - requires valid JWT
    security.SetEndpointProfile("/api/users", &security.EndpointProfile{})

    // Admin-only endpoint
    security.SetEndpointProfile("/api/admin", security.AdminEndpoint())

    // Endpoint for unverified users (e.g., email verification)
    security.SetEndpointProfile("/api/verify-email", security.UnverifiedEndpoint())

    // Service-to-service endpoint
    security.SetEndpointProfile("/api/internal", security.ServiceClientEndpoint("read:users", "write:users"))

    // Custom profile
    security.SetEndpointProfile("/api/premium", &security.EndpointProfile{
        AllowedAccountTypes: []string{"premium", "enterprise"},
    })

    // Method-specific profiles
    security.SetEndpointProfile("/api/posts", &security.EndpointProfile{
        AllowPublic: true,
    }, "GET")
    security.SetEndpointProfile("/api/posts", &security.EndpointProfile{}, "POST", "PUT", "DELETE")
}
```

#### Endpoint Profile Options

| Option | Type | Description |
|--------|------|-------------|
| `AllowPublic` | `bool` | Allow unauthenticated access |
| `AllowUnverified` | `bool` | Allow users with unverified email |
| `DenyVerified` | `bool` | Deny users with verified email (for verification endpoints) |
| `AllowExpiredJwt` | `bool` | Accept expired tokens (for refresh endpoints) |
| `RequiresAdmin` | `bool` | Require admin privileges |
| `AllowedAccountTypes` | `[]string` | Whitelist specific account levels |
| `AllowService` | `bool` | Allow service client tokens |
| `AllowedScopes` | `[]string` | Required scopes for service clients |

#### Extracting JWT Data from Context

```go
import "github.com/swayrider/swlib/security"

func protectedHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Get JWT claims
    claims := security.GetClaims(ctx)
    userID := claims.Subject
    email := claims.OpenID.Email

    // Get raw JWT token
    token := security.GetJwt(ctx)

    // Get refresh token (if present)
    refreshToken := security.GetRefreshToken(ctx)

    // Get client information
    clientIP := security.GetOrigIp(ctx)
    userAgent := security.GetUserAgent(ctx)
    host := security.GetHost(ctx)
    isSecure := security.GetSecure(ctx)
}
```

#### Evaluating Endpoint Profiles

```go
import "github.com/swayrider/swlib/security"

func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        profile := security.GetEndpointProfileForMethod(r.URL.Path, r.Method)
        if profile == nil {
            // No profile defined - deny by default
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }

        token := extractToken(r)

        result := profile.Evaluate(token, getPublicKeys, logger)
        if !result.Allowed {
            http.Error(w, result.Reason, http.StatusUnauthorized)
            return
        }

        // Add claims to context
        ctx := context.WithValue(r.Context(), security.ClaimsKey, result.Claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

---

### str - String Utilities

Simple string manipulation helpers.

```go
import "github.com/swayrider/swlib/str"

func processString() {
    // Remove null terminators from byte slice
    data := []byte("hello\x00\x00\x00")
    cleaned := str.NullTerm(data) // []byte("hello")

    // Convert string to pointer (nil for empty strings)
    name := "John"
    namePtr := str.ToPtr(name) // *string pointing to "John"

    empty := ""
    emptyPtr := str.ToPtr(empty) // nil
}
```

---

## Architecture Overview

```
+-------------------------------------------------------------+
|                        Application                          |
|  +-----------------------------------------------------+    |
|  |                    app.App                          |    |
|  |  +----------+ +----------+ +--------------------+   |    |
|  |  |  Config  | | Database | |  Service Clients   |   |    |
|  |  +----------+ +----------+ +--------------------+   |    |
|  |  +----------+ +----------+ +--------------------+   |    |
|  |  |   gRPC   | |   HTTP   | | Background Routines|   |    |
|  |  +----------+ +----------+ +--------------------+   |    |
|  +-----------------------------------------------------+    |
+-------------------------------------------------------------+
           |              |              |
           v              v              v
+---------------+ +-------------+ +-------------+
|   security/   | |    jwt/     | |   logger/   |
| Authorization | |   Tokens    | |   Logging   |
+---------------+ +-------------+ +-------------+
           |              |              |
           v              v              v
+---------------+ +-------------+ +-------------+
|    crypto/    | |    http/    | |    grpc/    |
|   Security    | |  Middleware | | Interceptors|
+---------------+ +-------------+ +-------------+
           |              |              |
           v              v              v
+---------------+ +-------------+ +-------------+
|     env/      | |    flag/    | |   cache/    |
|  Environment  | |  CLI Flags  | |   Caching   |
+---------------+ +-------------+ +-------------+
```

### Request Flow

```
HTTP Request
     |
     v
+-------------+
| SourceInfo  |  <- Extracts client IP, user agent
| Middleware  |
+-------------+
     |
     v
+-------------+
|    Auth     |  <- Validates JWT, checks endpoint profile
| Middleware  |
+-------------+
     |
     v
+-------------+
|   Handler   |  <- Access claims via security.GetClaims(ctx)
+-------------+
```

---

## Best Practices

### 1. Service Initialization

Always use the builder pattern with `app.New()`:

```go
app.New("servicename").
    WithDefaultConfigFields(app.BackendServiceFields | app.DatabaseConnectionFields).
    WithServiceClients(...).
    WithDatabase(...).
    WithGrpc(...).
    Run()
```

### 2. Configuration

- Use environment variables for all configuration
- Define required fields explicitly
- Provide sensible defaults where appropriate
- Use the pre-defined field groups for consistency

### 3. Security

- Always define endpoint profiles for authorization
- Use `security.GetClaims()` to access user information
- Never store passwords - use `crypto.CalculatePasswordHash()`
- Rotate JWT keys regularly using `crypto.CreateKeypair()`

### 4. Logging

- Create component-scoped loggers for traceability
- Use `Derive()` for function-level context
- Use appropriate log levels consistently

### 5. Error Handling

- Return errors up the call stack
- Log errors at the appropriate level
- Use `logger.Fatalf()` only for unrecoverable errors

### 6. Inter-Service Communication

- Use typed service clients from `grpcclients/`
- Handle authentication failures with `s2s.Call()`
- Configure service endpoints via environment variables
