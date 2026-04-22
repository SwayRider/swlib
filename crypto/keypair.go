package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"time"
)

// RSA key generation parameters.
const (
	// rsaKeySize is the bit size for generated RSA keys
	rsaKeySize = 2048
	// maxAge is the recommended validity period for keypairs (30 days)
	maxAge = 30 * 24 * time.Hour
)

// CreateKeypair generates a new RSA keypair for JWT signing.
// The keypair uses 2048-bit RSA and returns keys in PEM format.
//
// Returns:
//   - privatePEM: The private key in PEM format (for signing tokens)
//   - publicPEM: The public key in PEM format (for verifying tokens)
//   - validUntil: Recommended expiration time (30 days from creation)
//   - err: An error if the keypair could not be generated
//
// Example:
//
//	privateKey, publicKey, expiresAt, err := crypto.CreateKeypair()
//	if err != nil {
//	    log.Fatalf("Failed to create keypair: %v", err)
//	}
//	// Store privateKey securely for signing
//	// Distribute publicKey for verification
func CreateKeypair() (privatePEM, publicPEM string, validUntil *time.Time, err error) {
	key, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return
	}
	privBytes := x509.MarshalPKCS1PrivateKey(key)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
	pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	validUntil = new(time.Time)
	*validUntil = time.Now().Add(maxAge)

	privatePEM = string(privPEM)
	publicPEM = string(pubPEM)
	return
}
