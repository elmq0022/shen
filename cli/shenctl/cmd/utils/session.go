package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// getSessionFilePath returns the path to the session file.
func getSessionFilePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user cache directory: %w", err)
	}
	return filepath.Join(cacheDir, "shenctl", "session"), nil
}

// GetAuthHeader reads the session token from the cache directory and returns
// the full Authorization header value (e.g., "Bearer <token>").
// Returns an error if no session exists.
func GetAuthHeader() (string, error) {
	sessionFile, err := getSessionFilePath()
	if err != nil {
		return "", err
	}

	sessionToken, err := os.ReadFile(sessionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no active session found - please login first")
		}
		return "", fmt.Errorf("failed to read session file: %w", err)
	}

	return "Bearer " + strings.TrimSpace(string(sessionToken)), nil
}

// ClearSession removes the session file from the cache directory.
func ClearSession() error {
	sessionFile, err := getSessionFilePath()
	if err != nil {
		return err
	}
	return os.Remove(sessionFile)
}
