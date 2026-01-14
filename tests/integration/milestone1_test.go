//go:build cli_integration

package integration

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/handlers/auth"
	"github.com/elmq0022/shen/internal/keys"
	"github.com/stretchr/testify/assert"
)

// TestMilestone1_BootstrapAndAdminAuthentication orchestrates all Milestone 1 tests
// Goal: Set up the foundation - database, bootstrap process, and admin login via CLI
func TestMilestone1_BootstrapAndAdminAuthentication(t *testing.T) {
	// Setup: Reset database and start server
	resetDB(t)
	startTestServer(t, shen)
	xdgDirs := setTempXDG(t)

	t.Run("Bootstrap creates admin user and JWT keys", func(t *testing.T) {
		user, err := queries.GetUserByUsername(t.Context(), "admin")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "admin", user.Username)

		jwtKey, err := queries.GetActiveSigningKey(t.Context())
		if err != nil {
			t.Fatal(err)
		}

		assert.NotEqual(t, "", string(jwtKey.EncryptedPrivateKey))
		assert.Contains(t, jwtKey.PublicKey, "-----BEGIN PUBLIC KEY-----")
		assert.Contains(t, jwtKey.PublicKey, "-----END PUBLIC KEY-----")

		kek, err := keys.LoadKEK()
		if err != nil {
			t.Fatal(err)
		}

		decryptedPrivateKey, err := keys.DecryptPrivateKey(jwtKey.EncryptedPrivateKey, kek)
		assert.NoError(t, err, "Private key should decrypt successfully")
		assert.Contains(t, string(decryptedPrivateKey), "-----BEGIN PRIVATE KEY-----")
		assert.Contains(t, string(decryptedPrivateKey), "-----END PRIVATE KEY-----")
	})

	t.Run("Login with default admin credentials and verify session token is stored", func(t *testing.T) {
		login := exec.Command(shenctl, "auth", "login", "--username", "admin", "--password", "admin")
		if err := login.Run(); err != nil {
			t.Fatalf("failed to login %v", err)
		}

		cacheFile := filepath.Join(xdgDirs.CacheDir, "shenctl", "session")
		token, err := os.ReadFile(cacheFile)
		if err != nil {
			t.Fatal(err)
		}

		if string(token) == "" {
			t.Fatal("no session token")
		}
	})

	t.Run("Invalid credentials return 401", func(t *testing.T) {
		loginCmd := exec.Command(shenctl, "auth", "login", "--username", "bob", "--password", "bad-password")
		if err := loginCmd.Run(); err == nil {
			t.Fatal("logged in with bad creds")
		}
	})

	t.Run("JWKS endpoint returns valid JWK", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/.well-known/jwks.json")
		if err != nil {
			t.Fatalf("failed to fetch JWKS endpoint: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var jwks auth.JWKS
		if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
			t.Fatalf("failed to decode JWKS response: %v", err)
		}

		assert.NotEmpty(t, jwks.Keys)
		assert.Equal(t, "RSA", jwks.Keys[0].Kty)
		assert.Equal(t, "sig", jwks.Keys[0].Use)
	})

	t.Run("Logout revokes all of the active user session tokens", func(t *testing.T) {
		// TODO: Verify session token is no longer valid (401 on authenticated request)
		logoutCmd := exec.Command(shenctl, "auth", "logout")
		if err := logoutCmd.Run(); err != nil {
			t.Fatalf("logout failed: %v", err)
		}

		user, err := queries.GetUserByUsername(t.Context(), "admin")
		if err != nil {
			t.Fatalf("failed to get user: %v", err)
		}

		sessions, err := queries.ListActiveSessionsByUser(t.Context(), db.ListActiveSessionsByUserParams{
			Limit:    int32(10),
			UserID:   user.ID,
			CursorID: 0,
		})
		if err != nil {
			t.Fatalf("failed to list sessions: %v", err)
		}

		assert.Empty(t, sessions)
	})
}
