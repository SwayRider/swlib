// Package encryption provides AES-256-GCM symmetric encryption for secrets
// that need to be stored at rest (e.g. JWT signing keys), plus a KeyRing
// helper for rotating the master key without losing access to data
// encrypted under a previous one.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

// KeySize is the required key length in bytes for AES-256.
const KeySize = 32

// Encrypt encrypts plaintext with AES-256-GCM under key, using a random
// nonce. The returned blob is nonce || ciphertext || tag, in the layout
// produced by cipher.AEAD.Seal — everything Decrypt needs is in one value.
func Encrypt(key, plaintext []byte) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt reverses Encrypt. It returns an error if key is the wrong length,
// blob is truncated, or GCM authentication fails (wrong key or tampered
// data).
func Decrypt(key, blob []byte) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	if len(blob) < gcm.NonceSize() {
		return nil, errors.New("encryption: ciphertext shorter than nonce size")
	}

	nonce, ciphertext := blob[:gcm.NonceSize()], blob[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func newGCM(key []byte) (cipher.AEAD, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("encryption: key must be %d bytes, got %d", KeySize, len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// ParseMasterKey decodes a base64-encoded master key (as produced by
// `openssl rand -base64 32`) and validates it is KeySize bytes.
func ParseMasterKey(raw string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("encryption: master key is not valid base64: %w", err)
	}
	if len(key) != KeySize {
		return nil, fmt.Errorf("encryption: master key must decode to %d bytes, got %d", KeySize, len(key))
	}
	return key, nil
}

// Fingerprint returns a short, non-reversible identifier for key, suitable
// for tagging encrypted data with which key it was encrypted under. It does
// not reveal the key: it is a truncated SHA-256 hash, not the key itself.
func Fingerprint(key []byte) string {
	sum := sha256.Sum256(key)
	return hex.EncodeToString(sum[:4])
}
