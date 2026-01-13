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

type Server struct {
	cmd *exec.Cmd
}

func startServer(binaryPath string) *Server {
	cmd := exec.Command(binaryPath)
	s := &Server{cmd: cmd}

	if err := cmd.Start(); err != nil {
		panic(fmt.Errorf("could not start the test server: %w", err))
	}

	// Wait for server to be ready
	if err := waitForServer("http://localhost:8080"); err != nil {
		s.Stop()
		panic(fmt.Errorf("server failed to become ready: %w", err))
	}

	return s
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

func (s *Server) Stop() {
	if s == nil || s.cmd == nil || s.cmd.Process == nil {
		return
	}
	s.cmd.Process.Kill()
	s.cmd.Wait()
}

func setTempXDGConfig(t *testing.T) string {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "shen-test-config-*")
	if err != nil {
		t.Fatalf("could not create temp config dir: %v", err)
	}

	originalXDG := os.Getenv("XDG_CONFIG_HOME")

	if err := os.Setenv("XDG_CONFIG_HOME", tempDir); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("could not set XDG_CONFIG_HOME: %v", err)
	}

	t.Cleanup(func() {
		if originalXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
		os.RemoveAll(tempDir)
	})

	return tempDir
}
