// Package totp implements time-based one-time passwords (RFC 6238) and
// related MFA primitives: random base32 secrets, otpauth:// URLs, and
// unambiguous backup codes.
//
// The package is deliberately stdlib-only (crypto/hmac, crypto/sha1,
// crypto/rand, encoding/base32, crypto/subtle) so the shared swlib module
// stays dependency-light -- the same precedent as swlib/hibp. QR rendering
// and other third-party-dependent pieces live in the services that need
// them, not here.
package totp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

const (
	// Defaults applied when the matching Config field is zero.
	defaultSecretSize  = 20 // bytes -> 32 base32 chars
	defaultCodeLength  = 6  // digits
	defaultTimeStep    = 30 * time.Second
	defaultGracePeriod = 1 // accept ±N windows around the current one

	// maxCodeLength is the largest TOTP truncation supports (RFC 4226 §5.4).
	maxCodeLength = 8
)

// crockfordAlphabet is the Crockford base32 alphabet: 0-9 and A-Z minus
// the ambiguous characters I, L, O and U. Backup codes drawn from it are
// unambiguous to type (no 1/I, 0/O, etc. confusion).
const crockfordAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// Config tunes TOTP generation/validation. Zero values are replaced with
// the defaults above.
type Config struct {
	SecretSize  int           // random secret bytes; default 20
	CodeLength  int           // digits; default 6
	TimeStep    time.Duration // default 30s
	GracePeriod int           // accept ±N windows around the current one; default 1
}

// withDefaults returns c with zero fields replaced by the package defaults.
// CodeLength is clamped to [1, 8] and GracePeriod to >= 0 as defensive
// guards against misconfiguration.
func (c Config) withDefaults() Config {
	if c.SecretSize <= 0 {
		c.SecretSize = defaultSecretSize
	}
	if c.CodeLength < 1 {
		c.CodeLength = defaultCodeLength
	}
	if c.CodeLength > maxCodeLength {
		c.CodeLength = maxCodeLength
	}
	if c.TimeStep <= 0 {
		c.TimeStep = defaultTimeStep
	}
	if c.GracePeriod < 0 {
		c.GracePeriod = 0
	}
	return c
}

// GenerateSecret returns size random bytes as an unpadded, uppercase
// base32 (RFC 4648) string. size <= 0 falls back to the default (20 bytes
// -> 32 chars). crypto/rand is the only entropy source.
func GenerateSecret(size int) (string, error) {
	if size <= 0 {
		size = defaultSecretSize
	}
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("totp: failed to read random bytes: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), nil
}

// GenerateCode computes the RFC 6238 code for the given base32 secret at
// time t. The secret is base32-decoded (unpadded, case-insensitive) and
// used as the HMAC-SHA1 key; the counter is floor(t.Unix() / TimeStep).
//
// t is taken as a parameter (not time.Now()) so callers and tests can pin
// the clock.
func GenerateCode(secret string, t time.Time, cfg Config) (string, error) {
	cfg = cfg.withDefaults()

	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return "", fmt.Errorf("totp: invalid base32 secret: %w", err)
	}

	counter := uint64(t.Unix()) / uint64(cfg.TimeStep.Seconds())

	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], counter)

	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(msg[:])
	sum := mac.Sum(nil)

	// Dynamic truncation (RFC 4226 §5.3): the low nibble of the last MAC
	// byte selects the 4-byte window; the high bit is cleared.
	offset := sum[len(sum)-1] & 0x0f
	bin := (uint32(sum[offset]&0x7f) << 24) |
		(uint32(sum[offset+1]) << 16) |
		(uint32(sum[offset+2]) << 8) |
		uint32(sum[offset+3])

	mod := uint32(1)
	for i := 0; i < cfg.CodeLength; i++ {
		mod *= 10
	}

	return fmt.Sprintf("%0*d", cfg.CodeLength, bin%mod), nil
}

// Validate reports whether code matches secret within the current window
// plus GracePeriod windows before and after it, all evaluated at time t.
// Each candidate window is compared with crypto/subtle so no early string
// comparison leaks timing information about how close a guess was.
func Validate(secret, code string, t time.Time, cfg Config) (bool, error) {
	cfg = cfg.withDefaults()

	for i := -cfg.GracePeriod; i <= cfg.GracePeriod; i++ {
		candidate, err := GenerateCode(secret, t.Add(time.Duration(i)*cfg.TimeStep), cfg)
		if err != nil {
			return false, err
		}
		if subtle.ConstantTimeCompare([]byte(candidate), []byte(code)) == 1 {
			return true, nil
		}
	}
	return false, nil
}

// GenerateOTPAuthURL builds an otpauth://totp/ URI suitable for feeding to
// an authenticator app (or a QR encoder in the caller):
//
//	otpauth://totp/{issuer}:{account}?secret={secret}&issuer={issuer}
//
// account and issuer are percent-encoded (RFC 3986 unreserved characters
// only), so values containing ':', '@', '/' or other special characters
// survive the round trip. secret is expected to already be uppercase
// base32 (as produced by GenerateSecret) and is passed through verbatim.
func GenerateOTPAuthURL(secret, account, issuer string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
		percentEncode(issuer), percentEncode(account), secret, percentEncode(issuer))
}

// GenerateBackupCodes returns n random codes of the given length drawn
// from the Crockford base32 alphabet (0-9 A-Z minus I, L, O, U), so they
// are unambiguous to type. Codes are unique within the returned set.
//
// Codes are returned without spaces or separators; display layers group
// them in blocks of four (e.g. "ABCD EFGH") for readability. length <= 0
// is an error; n <= 0 returns an empty slice.
func GenerateBackupCodes(n, length int) ([]string, error) {
	if n <= 0 {
		return nil, nil
	}
	if length <= 0 {
		return nil, fmt.Errorf("totp: backup code length must be positive, got %d", length)
	}

	codes := make([]string, 0, n)
	seen := make(map[string]struct{}, n)
	for len(codes) < n {
		code, err := randomBackupCode(length)
		if err != nil {
			return nil, err
		}
		if _, dup := seen[code]; dup {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	return codes, nil
}

// randomBackupCode returns one length-character Crockford code. Each
// character is a fresh crypto/rand byte masked to 5 bits: the alphabet is
// exactly 32 characters, so 256 % 32 == 0 and the mask is unbiased.
func randomBackupCode(length int) (string, error) {
	var b strings.Builder
	b.Grow(length)
	var buf [1]byte
	for i := 0; i < length; i++ {
		if _, err := rand.Read(buf[:]); err != nil {
			return "", fmt.Errorf("totp: failed to read random bytes: %w", err)
		}
		b.WriteByte(crockfordAlphabet[buf[0]&0x1f])
	}
	return b.String(), nil
}

// percentEncode escapes every byte that is not an RFC 3986 unreserved
// character (A-Z a-z 0-9 - . _ ~), so the result is safe in any URL
// component and always encodes ':' and '@' (which Go's url.PathEscape and
// url.QueryEscape deliberately leave alone).
func percentEncode(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isUnreserved(c) {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}

// isUnreserved reports whether c is an RFC 3986 unreserved character.
func isUnreserved(c byte) bool {
	return 'a' <= c && c <= 'z' ||
		'A' <= c && c <= 'Z' ||
		'0' <= c && c <= '9' ||
		c == '-' || c == '.' || c == '_' || c == '~'
}
