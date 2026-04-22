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
  - [http - HTTP Middleware](#http---http-middleware)
  - [jwt - JWT Token Management](#jwt---jwt-token-management)
  - [logger - Structured Logging](#logger---structured-logging)
  - [math - Mathematical Utilities](#math---mathematical-utilities)
  - [security - Authorization Framework](#security---authorization-framework)
  - [str - String Utilities](#str---string-utilities)
- [Architecture Overview](#architecture-overview)
- [Best Practices](#best-practices)

## Installation

```bash
go get github.com/swayrider/swayrider/backend/swlib
```

## Quick Start

Here's a minimal example of creating a microservice using swlib:

```go
package main

import (
    "github.com/swayrider/swayrider/backend/swlib/app"
)

func main() {
    application := app.New("myservice").
        WithDefaultConfigFields(app.BackendServiceFields).
        WithGrpc(grpcSetup).
        Run()
}

func grpcSetup(a *app.App, grpcServer *grpc.Server, mux *runtime.ServeMux) {
    // Register your gRPC service handlers here
}
```

---

## Packages

### app - Service Bootstrap Framework

The `app` package is the core of swlib, providing a fluent builder pattern for configuring and running microservices. It handles service lifecycle, configuration, database connections, gRPC/HTTP servers, and graceful shutdown.

#### Basic Usage

```go
package main

import (
    "github.com/swayrider/swayrider/backend/swlib/app"
)

func main() {
    application := app.New("myservice").
        WithDefaultConfigFields(app.BackendServiceFields | app.DatabaseConnectionFields).
        Run()
}
```

#### Configuration Field Groups

The app package provides pre-defined configuration field groups that can be combined using bitwise OR:

| Field Group | Description | Environment Variables |
|-------------|-------------|----------------------|
| `BackendServiceFields` | Basic service configuration | `LOG_LEVEL`, `HTTP_PORT`, `GRPC_PORT` |
| `DatabaseConnectionFields` | PostgreSQL connection | `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` |
| `MinioConnectionFields` | MinIO object storage | `MINIO_HOST`, `MINIO_PORT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `MINIO_SECURE` |
| `WebServiceFields` | Web server configuration | `WEB_PORT`, `WEB_ROOT` |
| `ClientCredentialsFields` | OAuth client credentials | `CLIENT_ID`, `CLIENT_SECRET` |
| `HTMXServiceFields` | HTMX-specific configuration | Various HTMX settings |

#### Custom Configuration Fields

```go
func main() {
    application := app.New("myservice").
        WithDefaultConfigFields(app.BackendServiceFields).
        WithConfigFields(
            app.ConfigField{
                Name:        "api-key",
                Description: "External API key",
                EnvName:     "API_KEY",
                Required:    true,
            },
            app.ConfigField{
                Name:        "cache-ttl",
                Description: "Cache time-to-live in seconds",
                EnvName:     "CACHE_TTL",
                Default:     "300",
            },
        ).
        Run()

    // Access configuration values
    apiKey := application.GetConfig("api-key")
    cacheTTL := application.GetConfigAsInt("cache-ttl")
}
```

#### Database Integration

```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

func main() {
    application := app.New("myservice").
        WithDefaultConfigFields(app.BackendServiceFields | app.DatabaseConnectionFields).
        WithDatabase(
            // Database constructor
            func(a *app.App) (*sql.DB, error) {
                connStr := fmt.Sprintf(
                    "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
                    a.GetConfig("db-host"),
                    a.GetConfig("db-port"),
                    a.GetConfig("db-user"),
                    a.GetConfig("db-password"),
                    a.GetConfig("db-name"),
                )
                return sql.Open("postgres", connStr)
            },
            // Bootstrap function (optional migrations, etc.)
            func(a *app.App, db *sql.DB) error {
                // Run migrations or initial setup
                return nil
            },
        ).
        Run()

    // Access database
    db := app.GetDatabase[*sql.DB](application)
}
```

#### Object Store (MinIO) Integration

```go
import "github.com/minio/minio-go/v7"

func main() {
    application := app.New("myservice").
        WithDefaultConfigFields(app.BackendServiceFields | app.MinioConnectionFields).
        WithObjectStore(
            func(a *app.App) (*minio.Client, error) {
                return minio.New(
                    fmt.Sprintf("%s:%s", a.GetConfig("minio-host"), a.GetConfig("minio-port")),
                    &minio.Options{
                        Creds:  credentials.NewStaticV4(a.GetConfig("minio-access-key"), a.GetConfig("minio-secret-key"), ""),
                        Secure: a.GetConfigAsBool("minio-secure"),
                    },
                )
            },
            func(a *app.App, client *minio.Client) error {
                // Initialize buckets, etc.
                return nil
            },
        ).
        Run()

    // Access object store
    minioClient := app.GetObjectStore[*minio.Client](application)
}
```

#### gRPC and HTTP Gateway

```go
import (
    "google.golang.org/grpc"
    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func main() {
    application := app.New("myservice").
        WithDefaultConfigFields(app.BackendServiceFields).
        WithGrpc(setupGrpc).
        Run()
}

func setupGrpc(a *app.App, grpcServer *grpc.Server, mux *runtime.ServeMux) {
    // Register gRPC service
    pb.RegisterMyServiceServer(grpcServer, &MyServiceServer{app: a})

    // Register HTTP gateway (optional)
    pb.RegisterMyServiceHandlerServer(context.Background(), mux, &MyServiceServer{app: a})
}
```

#### Service Clients

For inter-service communication, use the service client pattern:

```go
import "github.com/swayrider/swayrider/backend/grpcclients/authclient"

func main() {
    application := app.New("myservice").
        WithDefaultConfigFields(app.BackendServiceFields).
        WithServiceClients(
            app.NewServiceClient("authservice", func(host, port string) (*authclient.Client, error) {
                return authclient.New(host, port)
            }),
        ).
        Run()

    // Access service client
    authClient := app.GetServiceClient[*authclient.Client](application, "authservice")
}
```

#### Background Routines

```go
func main() {
    application := app.New("myservice").
        WithDefaultConfigFields(app.BackendServiceFields).
        WithBackgroundRoutines(
            func(a *app.App) {
                ticker := time.NewTicker(5 * time.Minute)
                defer ticker.Stop()

                for {
                    select {
                    case <-a.Done():
                        return
                    case <-ticker.C:
                        // Periodic task
                        cleanupExpiredSessions(a)
                    }
                }
            },
        ).
        Run()
}
```

#### Application Data Store

Thread-safe key-value store for sharing data across components:

```go
func main() {
    application := app.New("myservice").
        WithDefaultConfigFields(app.BackendServiceFields).
        Run()

    // Store data
    application.SetData("cache-manager", cacheManager)

    // Retrieve data
    cm := app.GetData[*CacheManager](application, "cache-manager")
}
```

---

### cache - In-Memory Caching

Simple thread-safe local cache for storing arbitrary values.

#### Usage

```go
import "github.com/swayrider/swayrider/backend/swlib/cache"

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
import "github.com/swayrider/swayrider/backend/swlib/compression"

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
import "github.com/swayrider/swayrider/backend/swlib/crypto"

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
import "github.com/swayrider/swayrider/backend/swlib/crypto"

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
import "github.com/swayrider/swayrider/backend/swlib/crypto"

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
import "github.com/swayrider/swayrider/backend/swlib/env"

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
    "github.com/swayrider/swayrider/backend/swlib/swflag"
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
import "github.com/swayrider/swayrider/backend/swlib/swflag"

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
import "github.com/swayrider/swayrider/backend/swlib/swflag"

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
    "github.com/swayrider/swayrider/backend/swlib/grpc/interceptors"
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
import "github.com/swayrider/swayrider/backend/swlib/grpc/interceptors"

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
import "github.com/swayrider/swayrider/backend/swlib/grpc/s2s"

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

### http - HTTP Middleware

HTTP middleware components for authentication, content validation, and request context.

#### JWT Authentication Middleware

```go
import (
    "net/http"
    "github.com/swayrider/swayrider/backend/swlib/http/middlewares"
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
import "github.com/swayrider/swayrider/backend/swlib/http/middlewares"

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
import "github.com/swayrider/swayrider/backend/swlib/http/middlewares"

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
import "github.com/swayrider/swayrider/backend/swlib/http/middlewares"

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
import "github.com/swayrider/swayrider/backend/swlib/http/middlewares"

func setupStaticFiles() http.Handler {
    fs := http.Dir("./static")
    fileServer := http.FileServer(middlewares.NeuterFS{Fs: fs})
    return fileServer
}
```

#### Cookie Utilities

```go
import "github.com/swayrider/swayrider/backend/swlib/http/cookies"

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

---

### jwt - JWT Token Management

Comprehensive JWT handling with custom claims for user and service authentication.

#### Configuration

```go
import "github.com/swayrider/swayrider/backend/swlib/jwt"

func init() {
    // Configure JWT issuer and audience (typically done once at startup)
    jwt.Configure("https://auth.swayrider.com", "swayrider-services")
}
```

#### Generating Tokens

```go
import "github.com/swayrider/swayrider/backend/swlib/jwt"

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
import "github.com/swayrider/swayrider/backend/swlib/jwt"

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
import "github.com/swayrider/swayrider/backend/swlib/jwt"

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
import "github.com/swayrider/swayrider/backend/swlib/jwt"

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

### logger - Structured Logging

Context-aware logging with component and function tracking.

#### Basic Usage

```go
import "github.com/swayrider/swayrider/backend/swlib/logger"

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
import "github.com/swayrider/swayrider/backend/swlib/logger"

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
import "github.com/swayrider/swayrider/backend/swlib/math/geo"

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
import "github.com/swayrider/swayrider/backend/swlib/math/floats"

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

### security - Authorization Framework

Endpoint-based authorization and JWT context extraction.

#### Defining Endpoint Profiles

```go
import "github.com/swayrider/swayrider/backend/swlib/security"

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
import "github.com/swayrider/swayrider/backend/swlib/security"

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
import "github.com/swayrider/swayrider/backend/swlib/security"

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
import "github.com/swayrider/swayrider/backend/swlib/str"

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
