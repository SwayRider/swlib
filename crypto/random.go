package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
)

// GenerateSecureRandomString generates a secure random string of the given length
//
// Parameters:
//   - length: the length of the string to generate
//
// Returns:
//   - string: the generated string
//   - error: an error if the string could not be generated
func GenerateSecureRandomString(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be greater than 0")
	}

	byteLen := length / 2
	if length%2 != 0 {
		byteLen++
	}

	b := make([]byte, byteLen)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}
	return hex.EncodeToString(b)[:length], nil
}
