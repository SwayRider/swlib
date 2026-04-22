package crypto

import (
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"
	"time"
)

// =============================================================================
// Password Hashing Tests
// =============================================================================

func TestCalculatePasswordHash(t *testing.T) {
	password := "mySecurePassword123!"

	hash, err := CalculatePasswordHash(password)
	if err != nil {
		t.Fatalf("CalculatePasswordHash failed: %v", err)
	}

	// Verify hash format
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("expected hash to start with '$argon2id$', got: %s", hash[:20])
	}

	// Verify hash has correct number of parts
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		t.Errorf("expected 6 parts in hash, got %d", len(parts))
	}

	// Verify version
	if !strings.HasPrefix(parts[2], "v=") {
		t.Errorf("expected version field, got: %s", parts[2])
	}

	// Verify parameters
	if !strings.Contains(parts[3], "m=") || !strings.Contains(parts[3], "t=") || !strings.Contains(parts[3], "p=") {
		t.Errorf("expected m=, t=, p= parameters, got: %s", parts[3])
	}
}

func TestCalculatePasswordHash_UniqueHashes(t *testing.T) {
	password := "samePassword"

	hash1, err := CalculatePasswordHash(password)
	if err != nil {
		t.Fatalf("first hash failed: %v", err)
	}

	hash2, err := CalculatePasswordHash(password)
	if err != nil {
		t.Fatalf("second hash failed: %v", err)
	}

	// Same password should produce different hashes (different salts)
	if hash1 == hash2 {
		t.Error("expected different hashes for same password due to random salt")
	}
}

func TestVerifyPassword_Valid(t *testing.T) {
	password := "correctPassword"

	hash, err := CalculatePasswordHash(password)
	if err != nil {
		t.Fatalf("CalculatePasswordHash failed: %v", err)
	}

	valid, err := VerifyPassword(hash, password)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if !valid {
		t.Error("expected password to be valid")
	}
}

func TestVerifyPassword_Invalid(t *testing.T) {
	password := "correctPassword"
	wrongPassword := "wrongPassword"

	hash, err := CalculatePasswordHash(password)
	if err != nil {
		t.Fatalf("CalculatePasswordHash failed: %v", err)
	}

	valid, err := VerifyPassword(hash, wrongPassword)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if valid {
		t.Error("expected password to be invalid")
	}
}

func TestVerifyPassword_EmptyPassword(t *testing.T) {
	password := ""

	hash, err := CalculatePasswordHash(password)
	if err != nil {
		t.Fatalf("CalculatePasswordHash failed: %v", err)
	}

	valid, err := VerifyPassword(hash, password)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if !valid {
		t.Error("expected empty password to verify correctly")
	}

	// Wrong password should fail
	valid, err = VerifyPassword(hash, "notEmpty")
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if valid {
		t.Error("expected wrong password to fail")
	}
}

func TestVerifyPassword_UnsupportedHashType(t *testing.T) {
	// Simulate a bcrypt-style hash
	fakeHash := "$bcrypt$somedata$morestuff"

	_, err := VerifyPassword(fakeHash, "password")
	if err == nil {
		t.Error("expected error for unsupported hash type")
	}
	if !strings.Contains(err.Error(), "unsupported hash type") {
		t.Errorf("expected 'unsupported hash type' error, got: %v", err)
	}
}

func TestVerifyPassword_InvalidHashFormat(t *testing.T) {
	// Hash with wrong number of parts
	invalidHash := "$argon2id$v=19$toofewparts"

	_, err := VerifyPassword(invalidHash, "password")
	if err == nil {
		t.Error("expected error for invalid hash format")
	}
}

func TestVerifyPassword_InvalidVersion(t *testing.T) {
	// Hash with invalid version format
	invalidHash := "$argon2id$v=invalid$m=65536,t=1,p=4$salt$hash"

	_, err := VerifyPassword(invalidHash, "password")
	if err == nil {
		t.Error("expected error for invalid version")
	}
}

func TestVerifyPassword_WrongVersion(t *testing.T) {
	// Hash with wrong version number (not matching argon2.Version)
	invalidHash := "$argon2id$v=1$m=65536,t=1,p=4$c2FsdA==$aGFzaA=="

	_, err := VerifyPassword(invalidHash, "password")
	if err == nil {
		t.Error("expected error for wrong version")
	}
	if !strings.Contains(err.Error(), "invalid version") {
		t.Errorf("expected 'invalid version' error, got: %v", err)
	}
}

func TestVerifyPassword_InvalidMemoryParam(t *testing.T) {
	invalidHash := "$argon2id$v=19$m=invalid,t=1,p=4$c2FsdA==$aGFzaA=="

	_, err := VerifyPassword(invalidHash, "password")
	if err == nil {
		t.Error("expected error for invalid memory param")
	}
}

func TestVerifyPassword_InvalidTimeParam(t *testing.T) {
	invalidHash := "$argon2id$v=19$m=65536,t=invalid,p=4$c2FsdA==$aGFzaA=="

	_, err := VerifyPassword(invalidHash, "password")
	if err == nil {
		t.Error("expected error for invalid time param")
	}
}

func TestVerifyPassword_InvalidThreadsParam(t *testing.T) {
	invalidHash := "$argon2id$v=19$m=65536,t=1,p=invalid$c2FsdA==$aGFzaA=="

	_, err := VerifyPassword(invalidHash, "password")
	if err == nil {
		t.Error("expected error for invalid threads param")
	}
}

func TestVerifyPassword_UnknownParam(t *testing.T) {
	invalidHash := "$argon2id$v=19$m=65536,t=1,p=4,x=unknown$c2FsdA==$aGFzaA=="

	_, err := VerifyPassword(invalidHash, "password")
	if err == nil {
		t.Error("expected error for unknown param")
	}
	if !strings.Contains(err.Error(), "invalid parameter") {
		t.Errorf("expected 'invalid parameter' error, got: %v", err)
	}
}

func TestVerifyPassword_MissingParams(t *testing.T) {
	// Missing time parameter (t=0)
	invalidHash := "$argon2id$v=19$m=65536,p=4$c2FsdA==$aGFzaA=="

	_, err := VerifyPassword(invalidHash, "password")
	if err == nil {
		t.Error("expected error for missing params")
	}
}

func TestVerifyPassword_InvalidSalt(t *testing.T) {
	// Invalid base64 in salt
	invalidHash := "$argon2id$v=19$m=65536,t=1,p=4$not-valid-base64!@#$aGFzaA=="

	_, err := VerifyPassword(invalidHash, "password")
	if err == nil {
		t.Error("expected error for invalid salt")
	}
}

func TestVerifyPassword_InvalidHash(t *testing.T) {
	// Invalid base64 in hash
	invalidHash := "$argon2id$v=19$m=65536,t=1,p=4$c2FsdA==$not-valid-base64!@#"

	_, err := VerifyPassword(invalidHash, "password")
	if err == nil {
		t.Error("expected error for invalid hash")
	}
}

func TestVerifyPassword_SpecialCharacters(t *testing.T) {
	passwords := []string{
		"password with spaces",
		"пароль", // Russian
		"密码",    // Chinese
		"🔐🔑🗝️",  // Emojis
		"pass\nword\ttab",
		`pass"word'quote`,
	}

	for _, password := range passwords {
		hash, err := CalculatePasswordHash(password)
		if err != nil {
			t.Fatalf("CalculatePasswordHash failed for '%s': %v", password, err)
		}

		valid, err := VerifyPassword(hash, password)
		if err != nil {
			t.Fatalf("VerifyPassword failed for '%s': %v", password, err)
		}
		if !valid {
			t.Errorf("expected password '%s' to verify correctly", password)
		}
	}
}

func TestVerifyPassword_LongPassword(t *testing.T) {
	// Test with a very long password
	password := strings.Repeat("a", 10000)

	hash, err := CalculatePasswordHash(password)
	if err != nil {
		t.Fatalf("CalculatePasswordHash failed: %v", err)
	}

	valid, err := VerifyPassword(hash, password)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if !valid {
		t.Error("expected long password to verify correctly")
	}
}

// =============================================================================
// Keypair Tests
// =============================================================================

func TestCreateKeypair(t *testing.T) {
	privatePEM, publicPEM, validUntil, err := CreateKeypair()
	if err != nil {
		t.Fatalf("CreateKeypair failed: %v", err)
	}

	// Check private key format
	if !strings.Contains(privatePEM, "-----BEGIN RSA PRIVATE KEY-----") {
		t.Error("private key should be in PEM format with RSA PRIVATE KEY header")
	}
	if !strings.Contains(privatePEM, "-----END RSA PRIVATE KEY-----") {
		t.Error("private key should have RSA PRIVATE KEY footer")
	}

	// Check public key format
	if !strings.Contains(publicPEM, "-----BEGIN PUBLIC KEY-----") {
		t.Error("public key should be in PEM format with PUBLIC KEY header")
	}
	if !strings.Contains(publicPEM, "-----END PUBLIC KEY-----") {
		t.Error("public key should have PUBLIC KEY footer")
	}

	// Check validUntil
	if validUntil == nil {
		t.Fatal("validUntil should not be nil")
	}
	expectedExpiry := time.Now().Add(30 * 24 * time.Hour)
	if validUntil.Before(expectedExpiry.Add(-time.Minute)) || validUntil.After(expectedExpiry.Add(time.Minute)) {
		t.Errorf("validUntil should be about 30 days from now, got: %v", validUntil)
	}
}

func TestCreateKeypair_PrivateKeyParseable(t *testing.T) {
	privatePEM, _, _, err := CreateKeypair()
	if err != nil {
		t.Fatalf("CreateKeypair failed: %v", err)
	}

	block, _ := pem.Decode([]byte(privatePEM))
	if block == nil {
		t.Fatal("failed to decode private key PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	// Verify key size (should be 2048 bits)
	if privateKey.N.BitLen() != 2048 {
		t.Errorf("expected 2048-bit key, got %d bits", privateKey.N.BitLen())
	}
}

func TestCreateKeypair_PublicKeyParseable(t *testing.T) {
	_, publicPEM, _, err := CreateKeypair()
	if err != nil {
		t.Fatalf("CreateKeypair failed: %v", err)
	}

	block, _ := pem.Decode([]byte(publicPEM))
	if block == nil {
		t.Fatal("failed to decode public key PEM")
	}

	_, err = x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse public key: %v", err)
	}
}

func TestCreateKeypair_UniqueKeys(t *testing.T) {
	priv1, pub1, _, err := CreateKeypair()
	if err != nil {
		t.Fatalf("first CreateKeypair failed: %v", err)
	}

	priv2, pub2, _, err := CreateKeypair()
	if err != nil {
		t.Fatalf("second CreateKeypair failed: %v", err)
	}

	// Each call should generate unique keys
	if priv1 == priv2 {
		t.Error("expected different private keys")
	}
	if pub1 == pub2 {
		t.Error("expected different public keys")
	}
}

func TestCreateKeypair_KeyPairMatch(t *testing.T) {
	privatePEM, publicPEM, _, err := CreateKeypair()
	if err != nil {
		t.Fatalf("CreateKeypair failed: %v", err)
	}

	// Parse private key
	privBlock, _ := pem.Decode([]byte(privatePEM))
	privateKey, err := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	// Parse public key
	pubBlock, _ := pem.Decode([]byte(publicPEM))
	publicKeyInterface, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		t.Fatalf("failed to parse public key: %v", err)
	}

	// Verify the public key matches the private key's public component
	derivedPubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal derived public key: %v", err)
	}

	parsedPubBytes, err := x509.MarshalPKIXPublicKey(publicKeyInterface)
	if err != nil {
		t.Fatalf("failed to marshal parsed public key: %v", err)
	}

	if string(derivedPubBytes) != string(parsedPubBytes) {
		t.Error("public key doesn't match private key's public component")
	}
}

// =============================================================================
// Random String Tests
// =============================================================================

func TestGenerateSecureRandomString(t *testing.T) {
	lengths := []int{1, 8, 16, 32, 64, 128}

	for _, length := range lengths {
		result, err := GenerateSecureRandomString(length)
		if err != nil {
			t.Fatalf("GenerateSecureRandomString(%d) failed: %v", length, err)
		}

		if len(result) != length {
			t.Errorf("expected length %d, got %d", length, len(result))
		}

		// Verify it's valid hex
		for _, c := range result {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("expected hex characters, got: %c", c)
			}
		}
	}
}

func TestGenerateSecureRandomString_ZeroLength(t *testing.T) {
	_, err := GenerateSecureRandomString(0)
	if err == nil {
		t.Error("expected error for zero length")
	}
	if !strings.Contains(err.Error(), "length must be greater than 0") {
		t.Errorf("expected 'length must be greater than 0' error, got: %v", err)
	}
}

func TestGenerateSecureRandomString_NegativeLength(t *testing.T) {
	_, err := GenerateSecureRandomString(-5)
	if err == nil {
		t.Error("expected error for negative length")
	}
}

func TestGenerateSecureRandomString_OddLength(t *testing.T) {
	// Odd lengths should work correctly
	result, err := GenerateSecureRandomString(7)
	if err != nil {
		t.Fatalf("GenerateSecureRandomString(7) failed: %v", err)
	}

	if len(result) != 7 {
		t.Errorf("expected length 7, got %d", len(result))
	}
}

func TestGenerateSecureRandomString_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)

	// Generate 100 random strings and verify they're all unique
	for i := 0; i < 100; i++ {
		result, err := GenerateSecureRandomString(32)
		if err != nil {
			t.Fatalf("GenerateSecureRandomString failed: %v", err)
		}

		if seen[result] {
			t.Error("generated duplicate random string")
		}
		seen[result] = true
	}
}

func TestGenerateSecureRandomString_LargeLength(t *testing.T) {
	result, err := GenerateSecureRandomString(1000)
	if err != nil {
		t.Fatalf("GenerateSecureRandomString(1000) failed: %v", err)
	}

	if len(result) != 1000 {
		t.Errorf("expected length 1000, got %d", len(result))
	}
}

// =============================================================================
// Constants Tests
// =============================================================================

func TestPasswordMinEntropy(t *testing.T) {
	if PasswordMinEntropy != 80 {
		t.Errorf("expected PasswordMinEntropy to be 80, got %d", PasswordMinEntropy)
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkCalculatePasswordHash(b *testing.B) {
	password := "benchmarkPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CalculatePasswordHash(password)
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "benchmarkPassword123!"
	hash, _ := CalculatePasswordHash(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = VerifyPassword(hash, password)
	}
}

func BenchmarkCreateKeypair(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _, _ = CreateKeypair()
	}
}

func BenchmarkGenerateSecureRandomString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateSecureRandomString(32)
	}
}
