package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Time    = 3
	argon2Memory  = 128 * 1024
	argon2Threads = 4
	argon2KeyLen  = 32
	saltLen       = 16
)

type HashFunc func(password string) (string, error)
type CheckFunc func(password, hashedPassword string) (bool, error)

func HashedPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	return fmt.Sprintf("%x.%x", salt, hash), nil
}

func CheckPassword(password, hashedPassword string) (bool, error) {
	parts := strings.Split(hashedPassword, ".")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid hash format")
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false, err
	}
	storedHash, err := hex.DecodeString(parts[1])
	if err != nil {
		return false, err
	}
	newHash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	return subtle.ConstantTimeCompare(storedHash, newHash) == 1, nil
}
