package totp

import (
	"encoding/base32"
	"strings"
	"testing"
	"time"
)

// rfcSecret returns the base32 form of the RFC 6238 Appendix B test seed.
// The RFC vectors use the ASCII string "12345678901234567890" directly as
// the HMAC key; since GenerateCode base32-decodes its secret input, the
// equivalent base32 form decodes back to exactly those bytes.
func rfcSecret() string {
	return base32.StdEncoding.WithPadding(base32.NoPadding).
		EncodeToString([]byte("12345678901234567890"))
}

func TestGenerateCode_RFC6238Vectors8Digit(t *testing.T) {
	secret := rfcSecret()
	vectors := []struct {
		timeSec int64
		want    string
	}{
		{59, "94287082"},
		{1111111109, "07081804"},
		{1111111111, "14050471"},
		{1234567890, "89005924"},
		{2000000000, "69279037"},
		{20000000000, "65353130"},
	}

	for _, v := range vectors {
		got, err := GenerateCode(secret, time.Unix(v.timeSec, 0), Config{CodeLength: 8})
		if err != nil {
			t.Fatalf("GenerateCode(t=%d) failed: %v", v.timeSec, err)
		}
		if got != v.want {
			t.Errorf("GenerateCode(t=%d) = %q, want %q", v.timeSec, got, v.want)
		}
	}
}

func TestGenerateCode_RFC6238Vectors6DigitDefault(t *testing.T) {
	secret := rfcSecret()
	vectors := []struct {
		timeSec int64
		want    string
	}{
		{59, "287082"},
		{1111111109, "081804"},
		{1111111111, "050471"},
		{1234567890, "005924"},
		{2000000000, "279037"},
		{20000000000, "353130"},
	}

	for _, v := range vectors {
		got, err := GenerateCode(secret, time.Unix(v.timeSec, 0), Config{})
		if err != nil {
			t.Fatalf("GenerateCode(t=%d) failed: %v", v.timeSec, err)
		}
		if got != v.want {
			t.Errorf("GenerateCode(t=%d) = %q, want %q", v.timeSec, got, v.want)
		}
	}
}

func TestValidate_CurrentWindow(t *testing.T) {
	secret, err := GenerateSecret(0)
	if err != nil {
		t.Fatalf("GenerateSecret failed: %v", err)
	}

	now := time.Now()
	code, err := GenerateCode(secret, now, Config{})
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}

	ok, err := Validate(secret, code, now, Config{})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if !ok {
		t.Error("expected the current-window code to validate")
	}
}

func TestValidate_AcceptsGracePeriodWindows(t *testing.T) {
	secret, err := GenerateSecret(0)
	if err != nil {
		t.Fatalf("GenerateSecret failed: %v", err)
	}

	cfg := Config{GracePeriod: 1}
	now := time.Now()

	// A code from the previous and next windows must be accepted.
	for _, offset := range []time.Duration{-30 * time.Second, 30 * time.Second} {
		code, err := GenerateCode(secret, now.Add(offset), Config{})
		if err != nil {
			t.Fatalf("GenerateCode failed: %v", err)
		}
		ok, err := Validate(secret, code, now, cfg)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}
		if !ok {
			t.Errorf("expected code from %v to validate within the grace period", offset)
		}
	}
}

func TestValidate_RejectsOutsideGracePeriod(t *testing.T) {
	secret, err := GenerateSecret(0)
	if err != nil {
		t.Fatalf("GenerateSecret failed: %v", err)
	}

	cfg := Config{GracePeriod: 1}
	now := time.Now()

	// Two windows away either direction must be rejected.
	for _, offset := range []time.Duration{-90 * time.Second, 90 * time.Second} {
		code, err := GenerateCode(secret, now.Add(offset), Config{})
		if err != nil {
			t.Fatalf("GenerateCode failed: %v", err)
		}
		ok, err := Validate(secret, code, now, cfg)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}
		if ok {
			t.Errorf("expected code from %v to be rejected", offset)
		}
	}
}

func TestValidate_RejectsWrongCodeInValidWindow(t *testing.T) {
	secret, err := GenerateSecret(0)
	if err != nil {
		t.Fatalf("GenerateSecret failed: %v", err)
	}

	now := time.Now()
	code, err := GenerateCode(secret, now, Config{})
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}

	// Same length, but wrong digits, in an otherwise valid window.
	wrong := code
	if wrong[0] == '0' {
		wrong = "1" + wrong[1:]
	} else {
		wrong = "0" + wrong[1:]
	}

	ok, err := Validate(secret, wrong, now, Config{})
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if ok {
		t.Error("expected a wrong code in a valid window to be rejected")
	}
}

func TestValidate_NoGracePeriodOnlyAcceptsCurrentWindow(t *testing.T) {
	secret, err := GenerateSecret(0)
	if err != nil {
		t.Fatalf("GenerateSecret failed: %v", err)
	}

	cfg := Config{GracePeriod: 0}
	now := time.Now()

	prevCode, err := GenerateCode(secret, now.Add(-30*time.Second), Config{})
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	ok, err := Validate(secret, prevCode, now, cfg)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if ok {
		t.Error("expected the previous-window code to be rejected with GracePeriod 0")
	}
}

func TestValidate_InvalidBase32Secret(t *testing.T) {
	// '1' is not in the standard RFC 4648 base32 alphabet.
	if _, err := Validate("11111111", "123456", time.Now(), Config{}); err == nil {
		t.Error("expected an error for an invalid base32 secret")
	}
}

func TestGenerateSecret_LengthAndAlphabet(t *testing.T) {
	secret, err := GenerateSecret(0)
	if err != nil {
		t.Fatalf("GenerateSecret failed: %v", err)
	}
	if len(secret) != 32 {
		t.Errorf("secret length = %d, want 32 (20 bytes unpadded base32)", len(secret))
	}
	if secret != strings.ToUpper(secret) {
		t.Error("expected an uppercase secret")
	}
	if strings.Contains(secret, "=") {
		t.Error("expected an unpadded secret")
	}

	// The secret must round-trip through the base32 decoder.
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		t.Fatalf("secret is not valid base32: %v", err)
	}
	if len(decoded) != 20 {
		t.Errorf("decoded secret length = %d, want 20", len(decoded))
	}
}

func TestGenerateSecret_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		secret, err := GenerateSecret(0)
		if err != nil {
			t.Fatalf("GenerateSecret failed: %v", err)
		}
		if seen[secret] {
			t.Fatalf("duplicate secret generated: %q", secret)
		}
		seen[secret] = true
	}
}

func TestGenerateBackupCodes_CountLengthAlphabetUniqueness(t *testing.T) {
	codes, err := GenerateBackupCodes(10, 8)
	if err != nil {
		t.Fatalf("GenerateBackupCodes failed: %v", err)
	}
	if len(codes) != 10 {
		t.Fatalf("got %d codes, want 10", len(codes))
	}

	seen := make(map[string]bool, len(codes))
	for _, code := range codes {
		if len(code) != 8 {
			t.Errorf("code %q has length %d, want 8", code, len(code))
		}
		for _, c := range code {
			if !strings.ContainsRune(crockfordAlphabet, c) {
				t.Errorf("code %q contains %q, which is not in the Crockford alphabet", code, c)
			}
		}
		for _, banned := range []rune{'I', 'L', 'O', 'U'} {
			if strings.ContainsRune(code, banned) {
				t.Errorf("code %q contains ambiguous character %q", code, banned)
			}
		}
		if seen[code] {
			t.Errorf("duplicate backup code %q", code)
		}
		seen[code] = true
	}
}

func TestGenerateBackupCodes_EdgeInputs(t *testing.T) {
	codes, err := GenerateBackupCodes(0, 8)
	if err != nil {
		t.Fatalf("GenerateBackupCodes(0, 8) failed: %v", err)
	}
	if codes != nil {
		t.Errorf("expected nil for n=0, got %v", codes)
	}

	if _, err := GenerateBackupCodes(5, 0); err == nil {
		t.Error("expected an error for length 0")
	}
}

func TestGenerateOTPAuthURL_Structure(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	url := GenerateOTPAuthURL(secret, "user@example.com", "SwayRider")

	// ':' and '@' must be percent-encoded; secret passes through verbatim.
	want := "otpauth://totp/SwayRider:user%40example.com?secret=JBSWY3DPEHPK3PXP&issuer=SwayRider"
	if url != want {
		t.Errorf("GenerateOTPAuthURL = %q, want %q", url, want)
	}
	if !strings.HasPrefix(url, "otpauth://totp/") {
		t.Errorf("url %q does not start with otpauth://totp/", url)
	}
}

func TestGenerateOTPAuthURL_PercentEncodesSpecialChars(t *testing.T) {
	url := GenerateOTPAuthURL("SECRET", "john:smith@example.com", "ACME Corp")

	if !strings.Contains(url, "john%3Asmith%40example.com") {
		t.Errorf("account ':' and '@' not percent-encoded: %q", url)
	}
	// The issuer appears both in the path and the query; both must be encoded.
	if !strings.Contains(url, "ACME%20Corp") {
		t.Errorf("issuer space not percent-encoded: %q", url)
	}
	if !strings.Contains(url, "issuer=ACME%20Corp") {
		t.Errorf("query issuer not percent-encoded: %q", url)
	}
}

func TestGenerateCode_CodeLengthClamped(t *testing.T) {
	secret, err := GenerateSecret(0)
	if err != nil {
		t.Fatalf("GenerateSecret failed: %v", err)
	}

	// 12 clamps to 8.
	code, err := GenerateCode(secret, time.Now(), Config{CodeLength: 12})
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if len(code) != 8 {
		t.Errorf("code length = %d, want clamped 8", len(code))
	}

	// 0 falls back to 6.
	code, err = GenerateCode(secret, time.Now(), Config{CodeLength: 0})
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if len(code) != 6 {
		t.Errorf("code length = %d, want default 6", len(code))
	}
}
