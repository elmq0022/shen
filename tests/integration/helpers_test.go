//go:build cli_integration

package integration

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func compileBinary(pkgPath, outputName string) string {
	cmd := exec.Command("go", "build", "-o", outputName, pkgPath)
	if err := cmd.Run(); err != nil {
		panic(fmt.Sprintf("could not compile binary for %s: %v", pkgPath, err))
	}
	absPath, err := filepath.Abs(outputName)
	if err != nil {
		panic(fmt.Sprintf("could not get path for binary package %s: %v", pkgPath, err))
	}
	return absPath
}

func resetDB(t *testing.T) {
	t.Helper()

	const dbURLEnv = "DATABASE_URL"

	dbURL, ok := os.LookupEnv(dbURLEnv)
	if !ok {
		t.Fatalf("could not find the database url env: %s", dbURLEnv)
	}

	downCmd := exec.Command("migrate", "-database", dbURL, "-path", "../../db/migrations", "down", "-all")
	if err := downCmd.Run(); err != nil {
		t.Fatalf("could not migrate db down: %v", err)
	}

	upCmd := exec.Command("migrate", "-database", dbURL, "-path", "../../db/migrations", "up")
	if err := upCmd.Run(); err != nil {
		t.Fatalf("could not migrate db up: %v", err)
	}
}

func startTestServer(t *testing.T, binaryPath string) {
	t.Helper()

	cmd := exec.Command(binaryPath)

	if err := cmd.Start(); err != nil {
		t.Fatalf("could not start the test server: %v", err)
	}

	// Wait for server to be ready
	if err := waitForServer("http://localhost:8080"); err != nil {
		stopServer(cmd)
		t.Fatalf("server failed to become ready: %v", err)
	}

	t.Cleanup(func() {
		stopServer(cmd)
	})
}

func waitForServer(baseURL string) error {
	maxAttempts := 50 // 5 seconds max (50 * 100ms)
	for i := 0; i < maxAttempts; i++ {
		resp, err := http.Get(baseURL + "/api/v1/healthz")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("server failed to start within 5 seconds")
}

func stopServer(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	cmd.Process.Kill()
	cmd.Wait()
}

// XDGDirs holds temporary XDG directory paths for testing
type XDGDirs struct {
	ConfigDir string
	CacheDir  string
}

func setTempXDG(t *testing.T) *XDGDirs {
	t.Helper()

	configDir, err := os.MkdirTemp("", "shen-test-config-*")
	if err != nil {
		t.Fatalf("could not create temp config dir: %v", err)
	}

	cacheDir, err := os.MkdirTemp("", "shen-test-cache-*")
	if err != nil {
		os.RemoveAll(configDir)
		t.Fatalf("could not create temp cache dir: %v", err)
	}

	originalConfig := os.Getenv("XDG_CONFIG_HOME")
	originalCache := os.Getenv("XDG_CACHE_HOME")

	if err := os.Setenv("XDG_CONFIG_HOME", configDir); err != nil {
		os.RemoveAll(configDir)
		os.RemoveAll(cacheDir)
		t.Fatalf("could not set XDG_CONFIG_HOME: %v", err)
	}

	if err := os.Setenv("XDG_CACHE_HOME", cacheDir); err != nil {
		os.RemoveAll(configDir)
		os.RemoveAll(cacheDir)
		t.Fatalf("could not set XDG_CACHE_HOME: %v", err)
	}

	t.Cleanup(func() {
		if originalConfig != "" {
			os.Setenv("XDG_CONFIG_HOME", originalConfig)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
		if originalCache != "" {
			os.Setenv("XDG_CACHE_HOME", originalCache)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
		os.RemoveAll(configDir)
		os.RemoveAll(cacheDir)
	})

	return &XDGDirs{
		ConfigDir: configDir,
		CacheDir:  cacheDir,
	}
}
