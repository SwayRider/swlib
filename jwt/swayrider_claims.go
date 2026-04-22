package jwt

import (
	"errors"

	jwt5 "github.com/golang-jwt/jwt/v5"
)

// SwayRiderClaims is the interface for application-specific JWT claims.
// Implementations include SwayRiderUserClaims for user tokens and
// SwayRiderServiceClaims for service-to-service tokens.
type SwayRiderClaims interface {
	MapClaims() MapClaims
	FromMapClaims(m map[string]any) (err error)
}

// SwayRiderClaimsFromMapClaims parses SwayRiderClaims from a map structure.
// It determines the claim type from the "kind" field and returns the appropriate
// implementation (SwayRiderUserClaims or SwayRiderServiceClaims).
func SwayRiderClaimsFromMapClaims(m map[string]any) (c SwayRiderClaims, err error) {
	iface := m["swayrider"]
	if iface == nil {
		err = errors.New("missing swayrider claims")
		return
	}

	swMap, ok := iface.(map[string]any)
	if !ok {
		err = errors.New("invalid swayrider claims: expected map type")
		return
	}

	kind, err := mapField[string](swMap, "kind")
	if err != nil {
		return
	}

	switch kind {
	case "SwayRiderUserClaims":
		c = &SwayRiderUserClaims{}
	case "SwayRiderServiceClaims":
		c = &SwayRiderServiceClaims{}
	default:
		err = errors.New("invalid kind")
		return
	}

	return c, c.FromMapClaims(swMap)
}

// SwayRiderBaseClaims contains the discriminator field for claim type identification.
type SwayRiderBaseClaims struct {
	Kind string `json:"kind"` // "SwayRiderUserClaims" or "SwayRiderServiceClaims"
}

// SwayRiderUserClaims contains user-specific authorization claims.
type SwayRiderUserClaims struct {
	SwayRiderBaseClaims
	IsAdmin      bool   `json:"is_admin"`
	AccountLevel string `json:"account_level"`
}

// NewSwayRiderUserClaims creates user claims with the given admin status and account level.
// Account levels might include "free", "standard", "premium", etc.
func NewSwayRiderUserClaims(
	isAdmin bool,
	AccountLevel string,
) *SwayRiderUserClaims {
	return &SwayRiderUserClaims{
		SwayRiderBaseClaims: SwayRiderBaseClaims{
			Kind: "SwayRiderUserClaims",
		},
		IsAdmin:      isAdmin,
		AccountLevel: AccountLevel,
	}
}

// MapClaims converts SwayRiderUserClaims to a map for token generation.
func (c SwayRiderUserClaims) MapClaims() MapClaims {
	m := make(map[string]any)
	m["kind"] = "SwayRiderUserClaims"
	m["is_admin"] = c.IsAdmin
	m["account_level"] = c.AccountLevel
	return m
}

// FromMapClaims populates SwayRiderUserClaims from a map structure.
func (c *SwayRiderUserClaims) FromMapClaims(m map[string]any) (err error) {
	kind, err := mapField[string](m, "kind")
	if err != nil {
		return
	}
	if kind != "SwayRiderUserClaims" {
		err = errors.New("invalid kind")
		return
	}

	if c.IsAdmin, err = mapField[bool](m, "is_admin"); err != nil {
		return
	}
	if c.AccountLevel, err = mapField[string](m, "account_level"); err != nil {
		return
	}
	return nil
}

// SwayRiderServiceClaims contains service-to-service authorization claims.
// Services authenticate with scopes that define what operations they can perform.
type SwayRiderServiceClaims struct {
	SwayRiderBaseClaims
	Scopes jwt5.ClaimStrings `json:"scopes"` // List of allowed operation scopes
}

// NewSwayRiderServiceClaims creates service claims with the given permission scopes.
// Scopes define what operations the service can perform (e.g., "mail:send", "user:read").
func NewSwayRiderServiceClaims(
	scopes jwt5.ClaimStrings,
) *SwayRiderServiceClaims {
	return &SwayRiderServiceClaims{
		SwayRiderBaseClaims: SwayRiderBaseClaims{
			Kind: "SwayRiderServiceClaims",
		},
		Scopes: scopes,
	}
}

// MapClaims converts SwayRiderServiceClaims to a map for token generation.
func (c SwayRiderServiceClaims) MapClaims() MapClaims {
	m := make(map[string]any)
	m["kind"] = "SwayRiderServiceClaims"
	m["scopes"] = c.Scopes
	return m
}

// FromMapClaims populates SwayRiderServiceClaims from a map structure.
func (c *SwayRiderServiceClaims) FromMapClaims(m map[string]any) (err error) {
	kind, err := mapField[string](m, "kind")
	if err != nil {
		return
	}
	if kind != "SwayRiderServiceClaims" {
		err = errors.New("invalid kind")
		return
	}

	if c.Scopes, err = mapSlice[string](m, "scopes"); err != nil {
		return
	}

	return nil
}
