package encryption

import (
	"bytes"
	"strings"
	"testing"
)

func testKey(t *testing.T, seed byte) []byte {
	t.Helper()
	key := make([]byte, KeySize)
	for i := range key {
		key[i] = seed
	}
	return key
}

// =============================================================================
// Encrypt / Decrypt Tests
// =============================================================================

func TestEncryptDecrypt_Roundtrip(t *testing.T) {
	key := testKey(t, 1)

	// Realistic ~1.7KB payload, matching a PEM-encoded 3072-bit RSA private key.
	pemLike := strings.Repeat("A", 1700)

	cases := map[string][]byte{
		"empty":        {},
		"short":        []byte("hello"),
		"pem-like-3kb": []byte(pemLike),
	}

	for name, plaintext := range cases {
		t.Run(name, func(t *testing.T) {
			blob, err := Encrypt(key, plaintext)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			got, err := Decrypt(key, blob)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if !bytes.Equal(got, plaintext) {
				t.Errorf("roundtrip mismatch: got %q, want %q", got, plaintext)
			}
		})
	}
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	key := testKey(t, 1)

	blob, err := Encrypt(key, []byte("sensitive data"))
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	tampered := append([]byte(nil), blob...)
	tampered[len(tampered)-1] ^= 0xFF

	if _, err := Decrypt(key, tampered); err == nil {
		t.Error("expected error decrypting tampered ciphertext")
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	keyA := testKey(t, 1)
	keyB := testKey(t, 2)

	blob, err := Encrypt(keyA, []byte("sensitive data"))
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if _, err := Decrypt(keyB, blob); err == nil {
		t.Error("expected error decrypting with the wrong key")
	}
}

func TestDecrypt_TruncatedInput(t *testing.T) {
	key := testKey(t, 1)

	if _, err := Decrypt(key, []byte("short")); err == nil {
		t.Error("expected error for input shorter than the nonce size")
	}
}

func TestEncrypt_InvalidKeyLength(t *testing.T) {
	lengths := []int{0, 16, 24, 31, 33, 64}

	for _, length := range lengths {
		key := make([]byte, length)
		if _, err := Encrypt(key, []byte("data")); err == nil {
			t.Errorf("expected error for %d-byte key", length)
		}
	}
}

func TestEncrypt_UniqueNonces(t *testing.T) {
	key := testKey(t, 1)
	plaintext := []byte("same plaintext every time")

	blob1, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("first Encrypt failed: %v", err)
	}
	blob2, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("second Encrypt failed: %v", err)
	}

	if bytes.Equal(blob1, blob2) {
		t.Error("expected different ciphertext for repeated calls (nonce reuse)")
	}
}

// =============================================================================
// ParseMasterKey / Fingerprint Tests
// =============================================================================

func TestParseMasterKey_Valid(t *testing.T) {
	// base64 of 32 zero bytes
	raw := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

	key, err := ParseMasterKey(raw)
	if err != nil {
		t.Fatalf("ParseMasterKey failed: %v", err)
	}
	if len(key) != KeySize {
		t.Errorf("expected %d-byte key, got %d", KeySize, len(key))
	}
}

func TestParseMasterKey_InvalidBase64(t *testing.T) {
	if _, err := ParseMasterKey("not-valid-base64!@#"); err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestParseMasterKey_WrongLength(t *testing.T) {
	// base64 of 16 zero bytes, decodes fine but is too short
	if _, err := ParseMasterKey("AAAAAAAAAAAAAAAAAAAAAA=="); err == nil {
		t.Error("expected error for a key that decodes to the wrong length")
	}
}

func TestFingerprint_Deterministic(t *testing.T) {
	key := testKey(t, 1)

	first := Fingerprint(key)
	second := Fingerprint(key)
	if first != second {
		t.Errorf("expected Fingerprint to be deterministic for the same key: got %q then %q", first, second)
	}
}

func TestFingerprint_DiffersPerKey(t *testing.T) {
	keyA := testKey(t, 1)
	keyB := testKey(t, 2)

	if Fingerprint(keyA) == Fingerprint(keyB) {
		t.Error("expected different fingerprints for different keys")
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkEncrypt(b *testing.B) {
	key := make([]byte, KeySize)
	plaintext := []byte(strings.Repeat("A", 1700))

	for b.Loop() {
		_, _ = Encrypt(key, plaintext)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	key := make([]byte, KeySize)
	plaintext := []byte(strings.Repeat("A", 1700))
	blob, _ := Encrypt(key, plaintext)

	for b.Loop() {
		_, _ = Decrypt(key, blob)
	}
}
