package client

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/elmq0022/shen/cli/shenctl/utils"
)

func Logout() error {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get user cache directory: %w", err)
	}

	sessionFile := filepath.Join(cacheDir, "shenctl", "session")
	sessionToken, err := os.ReadFile(sessionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no active session found - please login first")
		}
		return fmt.Errorf("failed to read session file: %w", err)
	}

	req, err := utils.NewRequestBuilder(http.MethodPost, "/api/v1/auth/logout").
		WithHeader("Authorization", "Bearer "+strings.TrimSpace(string(sessionToken))).
		Build()

	if err != nil {
		return fmt.Errorf("failed to build logout request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("logout failed: server returned status %d", resp.StatusCode)
	}

	if err := os.Remove(sessionFile); err != nil {
		return fmt.Errorf("logout successful but failed to remove local session file: %w", err)
	}

	return nil
}
