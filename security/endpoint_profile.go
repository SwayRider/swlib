package security

import (
	"errors"
	"slices"
	"strings"

	"github.com/swayrider/swlib/jwt"
	log "github.com/swayrider/swlib/logger"
)

// Authentication and authorization errors returned by endpoint profile evaluation.
var (
	ErrMetaDataNotFound           = errors.New("metadata not found")
	ErrNoAuthToken                = errors.New("no authorization token")
	ErrInvalidAuthHeader          = errors.New("invalid authorization header")
	ErrInvalidJwt                 = errors.New("invalid jwt")
	ErrUserNotAdmin               = errors.New("user is not admin")
	ErrUserMissingRequiredAccount = errors.New("user is missing required account")
	ErrUserNotVerified            = errors.New("user is not verified")
	ErrUserAlreadyVerified        = errors.New("user is already verified")
	ErrNoKeys                     = errors.New("no public keys")
	ErrServiceClientNotAllowed    = errors.New("service client not allowed")
)

// PublicKeysFn is a function that returns the public keys for JWT verification.
// Multiple keys support key rotation scenarios.
type PublicKeysFn func() ([]string, error)

// EndpointProfile definition
//
// Default constructed object results in a restrcition set that requires a valid
// jwt token with standard authenticaiton behaviour. Any exceptions to the rule,
// e.g. unrestricted access, unverified access, etc needs tp be set via a custom
// restriction object on the endpont.
type EndpointProfile struct {
	// Incase of a pure HTTP Endpoint, you can use this to put specific restrictions
	// on the http method
	// Leave empty to put the restriction on all variants
	// You can register an empty one and then override a specific method as well
	HttpMethod string

	// AllowPublic allows unauthenticated users
	// use to allow registration and some public available functionalities
	AllowPublic bool

	// AllowUnverified allows unverified users (email not yet verified)
	// use to deny the actual functionality besides login and account management
	AllowUnverified bool

	// DenyVerified denies verified users (email already verified)
	// use for only allowing verification of unverified users
	DenyVerified bool

	// Allow expired jwt's
	// use this for example for refresh tokens, we require a jwt to be present,
	// but we allow it to be expired as we are using the refresh token to Generate
	// a new jwt
	AllowExpiredJwt bool

	// Requires the caller to be an admin user
	RequiresAdmin bool

	// List of account types that are allowed, if empty all are allowed
	AllowedAccountTypes []string

	// -------------------- Sevice Clients ---------------------

	// AllowService allows service claims to call the endpoint
	AllowService bool

	// AllowedScopes whitelists certain scopes to call the endpoint
	// If no scopes are provided, no service clients are allowed
	// If "*" is provided, all service clients are allowed
	AllowedScopes []string
}

// Global registry of endpoint profiles.
var endpointProfiles map[string]EndpointProfile

// EndpointProfiles returns all registered endpoint profiles.
func EndpointProfiles() map[string]EndpointProfile {
	return endpointProfiles
}

// GetEndpointProfile returns the profile for the given endpoint path.
// Returns a zero-value EndpointProfile if not found (requires auth by default).
func GetEndpointProfile(endpoint string) EndpointProfile {
	return endpointProfiles[endpoint]
}

// GetEndpointProfileForMethod returns the profile for an endpoint with HTTP method specificity.
// It first checks for "METHOD /path" key, then falls back to "/path" key.
func GetEndpointProfileForMethod(endpoint string, method string) EndpointProfile {
	m := strings.ToUpper(method)
	methodEndpoint := m + " " + endpoint
	if p, ok := endpointProfiles[methodEndpoint]; ok {
		return p
	}
	return GetEndpointProfile(endpoint)
}

func init() {
	endpointProfiles = make(map[string]EndpointProfile)
}

func endpointKeys(endpoint string, method ...string) (keys []string, methods []string) {
	if len(method) > 0 {
		for _, m := range method {
			if m == "" {
				keys = append(keys, endpoint)
				methods = append(methods, "")
				continue
			}
			m = strings.ToUpper(m)
			keys = append(keys, m+" "+endpoint)
			methods = append(methods, m)
		}
		return
	}
	return []string{endpoint}, []string{""}
}

func setEndpointProfile(key, method string, profile EndpointProfile) {
	p := profile
	p.HttpMethod = method
	endpointProfiles[key] = p
}

func getEndpointProfile(key string) (profile EndpointProfile, ok bool) {
	profile, ok = endpointProfiles[key]
	return
}

// SetEndpointProfile registers a complete endpoint profile.
// The method parameter allows setting profiles for specific HTTP methods.
func SetEndpointProfile(endpoint string, profile EndpointProfile, method ...string) {
	keys, methods := endpointKeys(endpoint, method...)
	for i, k := range keys {
		setEndpointProfile(k, methods[i], profile)
	}
}

// PublicEndpoint marks an endpoint as publicly accessible (no authentication required).
// Use for registration, health checks, and other public APIs.
func PublicEndpoint(endpoint string, method ...string) {
	keys, methods := endpointKeys(endpoint, method...)
	for i, k := range keys {
		p, ok := getEndpointProfile(k)
		if ok {
			p.AllowPublic = true
			setEndpointProfile(k, methods[i], p)
			continue
		}
		setEndpointProfile(k, methods[i], EndpointProfile{AllowPublic: true})
	}
}

// AdminEndpoint marks an endpoint as requiring admin privileges.
func AdminEndpoint(endpoint string, method ...string) {
	keys, methods := endpointKeys(endpoint, method...)
	for i, k := range keys {
		p, ok := getEndpointProfile(k)
		if ok {
			p.RequiresAdmin = true
			setEndpointProfile(k, methods[i], p)
			continue
		}
		setEndpointProfile(k, methods[i], EndpointProfile{RequiresAdmin: true})
	}
}

// UnverifiedEndpoint allows unverified users (email not verified) to access the endpoint.
// Use for endpoints that unverified users need, like profile management.
func UnverifiedEndpoint(endpoint string, methos ...string) {
	keys, methods := endpointKeys(endpoint, methos...)
	for i, k := range keys {
		p, ok := getEndpointProfile(k)
		if ok {
			p.AllowUnverified = true
			setEndpointProfile(k, methods[i], p)
			continue
		}
		setEndpointProfile(k, methods[i], EndpointProfile{AllowUnverified: true})
	}
}

// DenyVerifiedEndpoint denies access to users who have already verified their email.
// Use for verification endpoints that should only be accessible to unverified users.
func DenyVerifiedEndpoint(endpoint string, methos ...string) {
	keys, methods := endpointKeys(endpoint, methos...)
	for i, k := range keys {
		p, ok := getEndpointProfile(k)
		if ok {
			p.DenyVerified = true
			setEndpointProfile(k, methods[i], p)
			continue
		}
		setEndpointProfile(k, methods[i], EndpointProfile{DenyVerified: true})
	}
}

// AllowExpiredJwtEndpoint allows requests with expired JWT tokens.
// Use for token refresh endpoints where an expired token is expected.
func AllowExpiredJwtEndpoint(endpoint string, methos ...string) {
	keys, methods := endpointKeys(endpoint, methos...)
	for i, k := range keys {
		p, ok := getEndpointProfile(k)
		if ok {
			p.AllowExpiredJwt = true
			setEndpointProfile(k, methods[i], p)
			continue
		}
		setEndpointProfile(k, methods[i], EndpointProfile{AllowExpiredJwt: true})
	}
}

// ServiceClientEndpoint allows service clients to call the endpoint
//
// If no scopes are provided: "*" is added, allowing unscoped access
func ServiceClientEndpoint(endpoint string, scopes []string, method ...string) {
	if len(scopes) == 0 {
		scopes = append(scopes, "*")
	}

	keys, methods := endpointKeys(endpoint, method...)
	for i, k := range keys {
		p, ok := getEndpointProfile(k)
		if ok {
			p.AllowService = true
			p.AllowedScopes = scopes
			setEndpointProfile(k, methods[i], p)
			continue
		}
		setEndpointProfile(k, methods[i], EndpointProfile{AllowService: true, AllowedScopes: scopes})
	}
}

// Evaluate validates a request against this endpoint's security profile.
// It verifies the JWT token and checks all authorization requirements.
//
// Returns the parsed claims on success, or an error if authorization fails.
// For public endpoints with no token, returns (nil, nil).
func (p EndpointProfile) Evaluate(
	token *string,
	publicKeysFn PublicKeysFn,
	l *log.Logger,
) (claims *jwt.Claims, err error) {
	lg := l.Derive(log.WithFunction("EndpointProfile.Evaluate"))

	if token == nil {
		if !p.AllowPublic {
			err = ErrNoAuthToken
			lg.Debugf("%v", err)
		}
		return
	}

	claims, err = p.verifyJwt(token, publicKeysFn, lg)
	if err != nil {
		if p.AllowPublic {
			err = nil
			return
		}
		return
	}

	switch v := claims.SwayRiderClaims.(type) {
	case *jwt.SwayRiderUserClaims:
		err = p.evaluateUserClaims(claims, v, lg)
	case *jwt.SwayRiderServiceClaims:
		err = p.evaluateServiceClaims(claims, v, lg)
	default:
		err = ErrInvalidJwt
		lg.Debugf("%v", err)
		return
	}
	return
}

// evaluateUserClaims checks user-specific authorization requirements.
func (p EndpointProfile) evaluateUserClaims(
	claims *jwt.Claims,
	userClaims *jwt.SwayRiderUserClaims,
	lg *log.Logger,
) (err error) {
	if p.RequiresAdmin && !userClaims.IsAdmin {
		err = ErrUserNotAdmin
		lg.Debugf("%v", err)
		return
	}

	if len(p.AllowedAccountTypes) > 0 && !p.AllowPublic {
		if !slices.Contains(p.AllowedAccountTypes, userClaims.AccountLevel) {
			err = ErrUserMissingRequiredAccount
			lg.Debugf("%v", err)
			return
		}
	}

	if !p.AllowUnverified && !p.AllowPublic {
		if claims.EmailVerified == nil || !*claims.EmailVerified {
			err = ErrUserNotVerified
			lg.Debugf("%v", err)
			return
		}
	}

	if p.DenyVerified && !p.AllowPublic {
		if claims.EmailVerified != nil && *claims.EmailVerified {
			err = ErrUserAlreadyVerified
			lg.Debugf("%v", err)
			return
		}
	}

	return
}

// evaluateServiceClaims checks service-to-service authorization requirements.
func (p EndpointProfile) evaluateServiceClaims(
	_ *jwt.Claims,
	serviceClaims *jwt.SwayRiderServiceClaims,
	lg *log.Logger,
) (err error) {
	if !p.AllowService || len(p.AllowedScopes) == 0 {
		err = ErrServiceClientNotAllowed
		lg.Debugf("Endpoint does not allow service clients: %v", err)
		return
	}

	if len(p.AllowedScopes) > 0 {
		if slices.Contains(p.AllowedScopes, "*") {
			return
		}

		found := false
		for _, scope := range p.AllowedScopes {
			found = slices.Contains(serviceClaims.Scopes, scope)
			break
		}
		if !found {
			err = ErrServiceClientNotAllowed
			lg.Debugf("Missing correct scope (%v): %v", p.AllowedScopes, err)
			return
		}
	}
	return
}

// verifyJwt validates the JWT token against all available public keys.
func (p EndpointProfile) verifyJwt(
	token *string,
	publicKeysFn PublicKeysFn,
	l *log.Logger,
) (claims *jwt.Claims, err error) {
	lg := l.Derive(log.WithFunction("EndpointProfile.VerifyJwt"))

	publicKeys, err := publicKeysFn()
	if err != nil {
		lg.Debugf("%v: %v", ErrNoKeys, err)
		err = ErrNoKeys
		return
	}

	jwtVerifyMethod := jwt.VerifyDefault
	if p.AllowExpiredJwt {
		jwtVerifyMethod = jwt.VerifyOmitClaimsValidation
	}
	for _, pk := range publicKeys {
		claims, err = jwt.VerifyToken(*token, pk, jwtVerifyMethod)
		if err != nil {
			continue
		}
		return
	}
	claims = nil
	lg.Debugf("%v: %v", ErrInvalidJwt, err)
	err = ErrInvalidJwt
	return
}
