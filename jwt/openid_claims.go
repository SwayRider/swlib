package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/swayrider/swlib/math/floats"
)

// OpenIDClaims contains standard OpenID Connect claims for user profile information.
// All fields are optional (pointers) as per the OpenID Connect specification.
// See: https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims
type OpenIDClaims struct {
	// Full name
	Name *string `json:"name,omitempty"`

	// givven name(s) or first name(s)
	GivenName *string `json:"given_name,omitempty"`

	// Surname(s) or last name(s)
	FamilyName *string `json:"family_name,omitempty"`

	// Middle name(s)
	MiddleName *string `json:"middle_name,omitempty"`

	// Casual name
	Nickname *string `json:"nickname,omitempty"`

	// Shorthand name by which the End-User wishes to be referred to
	PreferredUsername *string `json:"preferred_username,omitempty"`

	// Profile page URL
	Profile *string `json:"profile,omitempty"`

	// Profile picture URL
	Picture *string `json:"picture,omitempty"`

	// Website or blog URL
	Website *string `json:"website,omitempty"`

	// Preferred e-mail address
	Email *string `json:"email,omitempty"`

	// True if the e-mail address has been verified; otherwise false
	EmailVerified *bool `json:"email_verified,omitempty"`

	// Gender
	Gender *string `json:"gender,omitempty"`

	// Birthdate
	Birthdate *string `json:"birthdate,omitempty"`

	// Time zone
	ZoneInfo *string `json:"zoneinfo,omitempty"`

	// Locale
	Locale *string `json:"locale,omitempty"`

	// Preferred telephone number
	PhoneNumber *string `json:"phone_number,omitempty"`

	// True if the phone number has been verified; otherwise false
	PhoneNumberVerified *bool `json:"phone_number_verified,omitempty"`

	// Preferred postal address
	Address *string `json:"address,omitempty"`

	// Time the information was last updated
	UpdatedTime *jwt.NumericDate `json:"updated_time,omitempty"`

	// Time when the authentication occurred
	AuthTime *jwt.NumericDate `json:"auth_time,omitempty"`
}

// MapClaims converts OpenIDClaims to a map structure for token generation.
// Only non-nil fields are included in the output map.
func (oc OpenIDClaims) MapClaims() MapClaims {
	m := make(map[string]any)
	if oc.Name != nil {
		m["name"] = *oc.Name
	}
	if oc.GivenName != nil {
		m["given_name"] = *oc.GivenName
	}
	if oc.FamilyName != nil {
		m["family_name"] = *oc.FamilyName
	}
	if oc.MiddleName != nil {
		m["middle_name"] = *oc.MiddleName
	}
	if oc.Nickname != nil {
		m["nickname"] = *oc.Nickname
	}
	if oc.PreferredUsername != nil {
		m["preferred_username"] = *oc.PreferredUsername
	}
	if oc.Profile != nil {
		m["profile"] = *oc.Profile
	}
	if oc.Picture != nil {
		m["picture"] = *oc.Picture
	}
	if oc.Website != nil {
		m["website"] = *oc.Website
	}
	if oc.Email != nil {
		m["email"] = *oc.Email
	}
	if oc.EmailVerified != nil {
		m["email_verified"] = *oc.EmailVerified
	}
	if oc.Gender != nil {
		m["gender"] = *oc.Gender
	}
	if oc.Birthdate != nil {
		m["birthdate"] = *oc.Birthdate
	}
	if oc.ZoneInfo != nil {
		m["zoneinfo"] = *oc.ZoneInfo
	}
	if oc.Locale != nil {
		m["locale"] = *oc.Locale
	}
	if oc.PhoneNumber != nil {
		m["phone_number"] = *oc.PhoneNumber
	}
	if oc.PhoneNumberVerified != nil {
		m["phone_number_verified"] = *oc.PhoneNumberVerified
	}
	if oc.Address != nil {
		m["address"] = *oc.Address
	}
	if oc.UpdatedTime != nil {
		m["updated_time"] = oc.UpdatedTime
	}
	if oc.AuthTime != nil {
		m["auth_time"] = oc.AuthTime
	}
	return m
}

// FromMapClaims populates OpenIDClaims from a map-based claims structure.
func (oc *OpenIDClaims) FromMapClaims(m map[string]any) (err error) {
	var ts *float64
	if oc.Name, err = mapField[*string](m, "name"); err != nil {
		return
	}
	if oc.Name, err = mapField[*string](m, "name"); err != nil {
		return
	}
	if oc.GivenName, err = mapField[*string](m, "given_name"); err != nil {
		return
	}
	if oc.FamilyName, err = mapField[*string](m, "family_name"); err != nil {
		return
	}
	if oc.MiddleName, err = mapField[*string](m, "middle_name"); err != nil {
		return
	}
	if oc.Nickname, err = mapField[*string](m, "nickname"); err != nil {
		return
	}
	if oc.PreferredUsername, err = mapField[*string](m, "preferred_username"); err != nil {
		return
	}
	if oc.Profile, err = mapField[*string](m, "profile"); err != nil {
		return
	}
	if oc.Picture, err = mapField[*string](m, "picture"); err != nil {
		return
	}
	if oc.Website, err = mapField[*string](m, "website"); err != nil {
		return
	}
	if oc.Email, err = mapField[*string](m, "email"); err != nil {
		return
	}
	if oc.EmailVerified, err = mapField[*bool](m, "email_verified"); err != nil {
		return
	}
	if oc.Gender, err = mapField[*string](m, "gender"); err != nil {
		return
	}
	if oc.Birthdate, err = mapField[*string](m, "birthdate"); err != nil {
		return
	}
	if oc.ZoneInfo, err = mapField[*string](m, "zoneinfo"); err != nil {
		return
	}
	if oc.Locale, err = mapField[*string](m, "locale"); err != nil {
		return
	}
	if oc.PhoneNumber, err = mapField[*string](m, "phone_number"); err != nil {
		return
	}
	if oc.PhoneNumberVerified, err = mapField[*bool](m, "phone_number_verified"); err != nil {
		return
	}
	if oc.Address, err = mapField[*string](m, "address"); err != nil {
		return
	}
	if ts, err = mapField[*float64](m, "updated_time"); err != nil {
		return
	} else if ts != nil && !floats.IsZero64(*ts) {
		oc.UpdatedTime = &jwt.NumericDate{Time: time.Unix(int64(*ts), 0)}
	}
	if ts, err = mapField[*float64](m, "auth_time"); err != nil {
		return
	} else if ts != nil && !floats.IsZero64(*ts) {
		oc.AuthTime = &jwt.NumericDate{Time: time.Unix(int64(*ts), 0)}
	}
	return nil
}

// SetUpdatedTime sets the time when the user profile was last updated.
func (oc *OpenIDClaims) SetUpdatedTime(t time.Time) {
	oc.UpdatedTime = &jwt.NumericDate{Time: t}
}

// SetAuthTime sets the time when the user authentication occurred.
func (oc *OpenIDClaims) SetAuthTime(t time.Time) {
	oc.AuthTime = &jwt.NumericDate{Time: t}
}
