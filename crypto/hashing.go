// Package crypto provides cryptographic utilities for password hashing,
// secure random generation, and RSA keypair management.
//
// # Password Hashing
//
// Uses Argon2id, the winner of the Password Hashing Competition, for secure
// password storage. Argon2id is resistant to both GPU and side-channel attacks.
//
//	hash, err := crypto.CalculatePasswordHash("mypassword")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Store hash in database
//
//	valid, err := crypto.VerifyPassword(storedHash, "mypassword")
//
// # Random Generation
//
// Generates cryptographically secure random strings:
//
//	token, err := crypto.GenerateSecureRandomString(32)
//
// # RSA Keypairs
//
// Creates 2048-bit RSA keypairs for JWT signing:
//
//	privateKey, publicKey, expiresAt, err := crypto.CreateKeypair()
package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters for password hashing.
// These settings provide a good balance between security and performance:
// - 64MB memory usage
// - 1 iteration
// - 4 parallel threads
// - 32-byte hash output
const (
	argonMemory  = 64 * 1024
	argonTime    = 1
	argonThreads = 4
	argonHashLen = 32
)

// PasswordMinEntropy defines the minimum entropy bits required for passwords.
const PasswordMinEntropy = 80

// CalculatePasswordHash generates a password hash using Argon2id
//
// Parameters:
//   - password: the password to hash
//
// Returns:
//   - string: the hash of the password
//   - error: an error if the hash could not be generated
func CalculatePasswordHash(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonHashLen)
	b64sal := base64.URLEncoding.EncodeToString(salt)
	b64hash := base64.URLEncoding.EncodeToString(hash)
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads,
		b64sal, b64hash)
	return encoded, nil
}

// VerifyPassword verifies a password against a hash
//
// Parameters:
//   - encodedHash: the hash of the password
//   - password: the password to verify
//
// Returns:
//   - bool: true if the password is valid, false otherwise
//   - error: an error if the hash could not be verified
func VerifyPassword(encodedHash, password string) (bool, error) {
	parts := strings.Split(encodedHash, "$")

	switch parts[1] {
	case "argon2id":
		return verifyPasswordArgon2id(parts, password)
	default:
		return false, fmt.Errorf("unsupported hash type: %s", parts[1])
	}
}

// verifyPasswordArgon2id verifies a password against an Argon2id hash
//
// Parameters:
//   - hashParts: the parts of the hash
//   - password: the password to verify
//
// Returns:
//   - bool: true if the password is valid, false otherwise
//   - error: an error if the hash could not be verified
func verifyPasswordArgon2id(hashParts []string, password string) (bool, error) {
	if len(hashParts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}

	var version int
	var memory, time uint32
	var threads uint8

	if _, err := fmt.Sscanf(hashParts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("invalid version: %w", err)
	}

	if version != argon2.Version {
		return false, fmt.Errorf("invalid version")
	}

	for param := range strings.SplitSeq(hashParts[3], ",") {
		var value uint32
		if strings.HasPrefix(param, "m=") {
			if _, err := fmt.Sscanf(param, "m=%d", &value); err != nil {
				return false, fmt.Errorf("invalid memory param: %w", err)
			}
			memory = value
		} else if strings.HasPrefix(param, "t=") {
			if _, err := fmt.Sscanf(param, "t=%d", &value); err != nil {
				return false, fmt.Errorf("invalid time param: %w", err)
			}
			time = value
		} else if strings.HasPrefix(param, "p=") {
			if _, err := fmt.Sscanf(param, "p=%d", &value); err != nil {
				return false, fmt.Errorf("invalid threads param: %w", err)
			}
			threads = uint8(value)
		} else {
			return false, fmt.Errorf("invalid parameter: %s", param)
		}
	}

	if memory == 0 || time == 0 || threads == 0 {
		return false, fmt.Errorf("invalid parameters")
	}

	salt, err := base64.URLEncoding.DecodeString(hashParts[4])
	if err != nil {
		return false, fmt.Errorf("invalid salt: %w", err)
	}

	hash, err := base64.URLEncoding.DecodeString(hashParts[5])
	if err != nil {
		return false, fmt.Errorf("invalid hash: %w", err)
	}

	computedHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(hash)))

	if subtle.ConstantTimeCompare(hash, computedHash) == 1 {
		return true, nil
	}
	return false, nil
}
