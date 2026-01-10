package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/elmq0022/shen/cli/shenctl/utils"
	"github.com/elmq0022/shen/internal/handlers/auth"
)

func Login() error {
	username, err := utils.ReadUsername()
	if err != nil {
		return fmt.Errorf("failed to read username: %w", err)
	}

	password, err := utils.ReadPassword()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	loginRequest := auth.LoginRequest{
		Username: username,
		Password: password,
	}

	req, err := utils.NewRequestBuilder(http.MethodPost, "http://localhost:8080/api/v1/auth/login").
		WithJSON(loginRequest).
		Build()

	if err != nil {
		return fmt.Errorf("failed to build login request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	var loginResponse auth.LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	sessionBytes := []byte(loginResponse.SessionToken)
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get user cache directory: %w", err)
	}

	shenCacheDir := filepath.Join(cacheDir, "shenctl")
	if err := os.MkdirAll(shenCacheDir, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cachePath := filepath.Join(shenCacheDir, "session")
	if err := os.WriteFile(cachePath, sessionBytes, 0600); err != nil {
		return fmt.Errorf("could not write session token to local file cache: %w", err)
	}

	return nil
}
