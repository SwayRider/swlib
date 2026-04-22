// Package jwt provides JWT token generation and verification using RS256 signing.
//
// It extends standard JWT claims with OpenID Connect claims and custom SwayRider
// claims for user authorization levels and service-to-service authentication.
//
// # Token Structure
//
// Tokens contain three claim types:
//   - RegisteredClaims: Standard JWT claims (iss, sub, aud, exp, iat, nbf, jti)
//   - OpenIDClaims: User profile information (name, email, etc.)
//   - SwayRiderClaims: Application-specific claims (admin status, account level, scopes)
//
// # Usage
//
//	// Configure issuer and audience
//	jwt.Configure("SwayRider", "SwayRider")
//
//	// Generate a token
//	openID := &jwt.OpenIDClaims{Email: &email}
//	swClaims := jwt.NewSwayRiderUserClaims(false, "standard")
//	jwtID, token, validUntil, err := jwt.GenerateToken(userID, openID, swClaims, privateKey, jwt.DefaultTTL)
//
//	// Verify a token
//	claims, err := jwt.VerifyToken(tokenString, publicKey, jwt.VerifyDefault)
//	if err != nil {
//	    // Handle invalid token
//	}
//	userID := claims.Subject
package jwt

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"time"

	jwt5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// DefaultTTL is the default token lifetime (15 minutes).
const (
	DefaultTTL = 15 * time.Minute
)

// Package-level configuration for JWT issuer and audience.
var jwtIssuer = "SwayRider"
var jwtAudience = "SwayRider"

// AccessToken is a string type representing a signed JWT token.
type AccessToken string

// MapClaims is an alias for jwt5.MapClaims for map-based claim access.
type MapClaims = jwt5.MapClaims

// CustomClaims contains legacy custom claim fields.
// Deprecated: Use SwayRiderUserClaims instead.
type CustomClaims struct {
	IsAdmin      bool   `json:"is_admin"`
	IsVerified   bool   `json:"is_verified"`
	AccountLevel string `json:"account_level"`
}

// Claims is the complete JWT claims structure containing registered claims,
// OpenID Connect claims, and SwayRider-specific claims.
type Claims struct {
	jwt5.RegisteredClaims
	OpenIDClaims
	SwayRiderClaims `json:"swayrider"`
}

// FromMapClaims populates Claims from a map-based claims structure.
// This is used internally during token verification to convert parsed claims.
func (c *Claims) FromMapClaims(m MapClaims) (err error) {
	if c.ExpiresAt, err = m.GetExpirationTime(); err != nil {
		return
	}
	if c.IssuedAt, err = m.GetIssuedAt(); err != nil {
		return
	}
	if c.NotBefore, err = m.GetNotBefore(); err != nil {
		return
	}
	if c.Subject, err = m.GetSubject(); err != nil {
		return
	}
	if c.Issuer, err = m.GetIssuer(); err != nil {
		return
	}
	if c.Audience, err = m.GetAudience(); err != nil {
		return
	}

	if c.ID, err = mapField[string](m, "jti"); err != nil {
		return
	}

	if err = c.OpenIDClaims.FromMapClaims(m); err != nil {
		return
	}

	if c.SwayRiderClaims, err = SwayRiderClaimsFromMapClaims(m); err != nil {
		return
	}
	return
}

// MapClaims converts Claims to a map structure for token generation.
func (c Claims) MapClaims() MapClaims {
	m := make(map[string]any)
	m["exp"] = c.ExpiresAt
	m["iat"] = c.IssuedAt
	m["nbf"] = c.NotBefore
	m["sub"] = c.Subject
	m["iss"] = c.Issuer
	m["aud"] = c.Audience
	m["jti"] = c.ID
	maps.Copy(m, c.OpenIDClaims.MapClaims())
	maps.Copy(m, c.SwayRiderClaims.MapClaims())
	return m
}

// Configure sets the JWT issuer and audience for token generation and verification.
// This should be called during application initialization.
func Configure(issuer string, audience string) {
	jwtIssuer = issuer
	jwtAudience = audience
}

// VerifyOpts configures token verification behavior.
type VerifyOpts uint8

// Verification options.
const (
	// VerifyDefault performs standard verification including expiration check.
	VerifyDefault VerifyOpts = 0
	// VerifyOmitClaimsValidation skips expiration validation (useful for refresh flows).
	VerifyOmitClaimsValidation VerifyOpts = 0x01
)

// VerifyToken validates a JWT token and returns its claims.
// It verifies the signature using RS256 and checks the audience claim.
//
// Parameters:
//   - token: The JWT token string to verify
//   - publicKeyPEM: The RSA public key in PEM format for signature verification
//   - opts: Verification options (VerifyDefault or VerifyOmitClaimsValidation)
//
// Returns an error if the token is invalid, expired, or has an invalid signature.
func VerifyToken(
	token string,
	publicKeyPEM string,
	opts VerifyOpts,
) (jwtClaims *Claims, err error) {
	key, err := jwt5.ParseRSAPublicKeyFromPEM([]byte(publicKeyPEM))
	if err != nil {
		return
	}

	tok, err := jwt5.Parse(
		token,
		func(t *jwt5.Token) (any, error) {
			if _, ok := t.Method.(*jwt5.SigningMethodRSA); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return key, nil
		},
		func() []jwt5.ParserOption {
			jwtOpts := []jwt5.ParserOption{
				jwt5.WithValidMethods([]string{jwt5.SigningMethodRS256.Name}),
				jwt5.WithAudience(jwtAudience),
				jwt5.WithIssuedAt(),
			}
			if opts&VerifyOmitClaimsValidation != VerifyOmitClaimsValidation {
				jwtOpts = append(jwtOpts, jwt5.WithExpirationRequired())
			} else {
				jwtOpts = append(jwtOpts, jwt5.WithoutClaimsValidation())
			}
			return jwtOpts
		}()...,
	)
	if err != nil {
		return
	}
	if !tok.Valid {
		err = errors.New("invalid token")
		return
	}

	mapClaims, ok := tok.Claims.(MapClaims)
	if !ok {
		err = errors.New("invalid claims")
		return
	}
	jwtClaims = &Claims{}
	jwtClaims.FromMapClaims(mapClaims)

	// Verify NotBefore
	if opts&VerifyOmitClaimsValidation == VerifyOmitClaimsValidation &&
		time.Now().Before(jwtClaims.NotBefore.Time) {
		err = errors.New("token not yet valid")
		return
	}

	return
}

// GenerateToken creates a new signed JWT token.
//
// Parameters:
//   - userId: The user ID to set as the token subject
//   - openIdClaims: Optional OpenID Connect claims (can be nil)
//   - swayRiderClaims: Application-specific claims (user or service claims)
//   - privateKeyPEM: The RSA private key in PEM format for signing
//   - ttl: Token time-to-live duration
//
// Returns:
//   - jwtID: Unique identifier for the token (useful for revocation)
//   - accessToken: The signed JWT token string
//   - validUntil: Token expiration time
//   - err: Any error that occurred
func GenerateToken(
	userId string,
	openIdClaims *OpenIDClaims,
	swayRiderClaims SwayRiderClaims,
	privateKeyPEM string,
	ttl time.Duration,
) (jwtID string, accessToken AccessToken, validUntil time.Time, err error) {
	key, err := jwt5.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return
	}

	expiresAt := time.Now().Add(ttl)
	claims := &Claims{
		RegisteredClaims: jwt5.RegisteredClaims{
			ExpiresAt: jwt5.NewNumericDate(expiresAt),
			IssuedAt:  jwt5.NewNumericDate(time.Now()),
			NotBefore: jwt5.NewNumericDate(time.Now()),
			Subject:   userId,
			Issuer:    jwtIssuer,
			Audience:  jwt5.ClaimStrings{jwtAudience},
			ID:        uuid.New().String(),
		},
	}
	if openIdClaims != nil {
		claims.OpenIDClaims = *openIdClaims
	}
	if swayRiderClaims != nil {
		claims.SwayRiderClaims = swayRiderClaims
	}

	tok := jwt5.NewWithClaims(jwt5.SigningMethodRS256, claims)
	token, err := tok.SignedString(key)
	if err != nil {
		return
	}
	accessToken = AccessToken(token)
	jwtID = claims.ID
	validUntil = expiresAt
	return
}

// mapField extracts a typed field from a map, handling both value and pointer types.
func mapField[T any](m map[string]any, key string) (T, error) {
	var zero T
	tType := reflect.TypeOf(zero)
	required := tType.Kind() != reflect.Ptr

	v, ok := m[key]
	if !ok {
		if required {
			return zero, fmt.Errorf("%s not found", key)
		}
		return zero, nil
	}

	// Value afhandeling
	val, ok := v.(T)
	if ok {
		return val, nil
	}

	// Pointer afhandeling
	if !required {
		elemType := tType.Elem()
		if reflect.TypeOf(v) == elemType {
			ptr := reflect.New(elemType)
			ptr.Elem().Set(reflect.ValueOf(v))
			return ptr.Interface().(T), nil
		}
	}

	return zero, fmt.Errorf("invalid type for %s", key)
}

// mapSlice extracts a typed slice from a map's interface{} slice.
func mapSlice[T any](m map[string]any, key string) ([]T, error) {
	var zero []T

	v, ok := m[key]
	if !ok {
		return zero, nil
	}

	ifacearr, ok := v.([]any)
	if !ok {
		return zero, fmt.Errorf("invalid type for %s", key)
	}
	slice := make([]T, 0, len(ifacearr))
	for _, iface := range ifacearr {
		val, ok := iface.(T)
		if !ok {
			return zero, fmt.Errorf("invalid type for %s", key)
		}
		slice = append(slice, val)
	}
	return slice, nil
}
