package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	tokenLen = 32 // 32 bytes = 256 bits
)

// GenerateSessionToken generates a cryptographically secure random session token.
// Returns the token as a hex-encoded string.
func GenerateSessionToken() (string, error) {
	bytes := make([]byte, tokenLen)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// HashToken hashes a session token using SHA-256.
// This hash should be stored in the database, not the original token.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
