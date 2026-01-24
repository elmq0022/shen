//go:build cli_integration

package integration

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/crypto"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMilestone6_PersonalAccessTokens(t *testing.T) {
	resetDB(t)
	startTestServer(t, shen)
	xdgDirs := setTempXDG(t)
	_ = xdgDirs

	// Login as admin first
	adminLogin := exec.Command(shenctl, "auth", "login", "--username", "admin", "--password", "admin")
	require.NoError(t, adminLogin.Run(), "admin login should succeed")

	// Create test application
	createAppCmd := exec.Command(shenctl, "application", "create", "testapp")
	require.NoError(t, createAppCmd.Run(), "create application testapp should succeed")

	// Create test group
	createGroupCmd := exec.Command(shenctl, "group", "create", "testgroup")
	require.NoError(t, createGroupCmd.Run(), "create group testgroup should succeed")

	// Add viewer role to testgroup for testapp
	addRoleCmd := exec.Command(shenctl, "group", "add-role", "testgroup", "testapp", "viewer")
	require.NoError(t, addRoleCmd.Run(), "add viewer role should succeed")

	// Create test user
	createUserCmd := exec.Command(shenctl, "user", "create", "testuser", "user", "--password", "testpass")
	require.NoError(t, createUserCmd.Run(), "create user testuser should succeed")

	// Add testuser to testgroup
	addUserCmd := exec.Command(shenctl, "group", "add-users", "testgroup", "testuser")
	require.NoError(t, addUserCmd.Run(), "add user to group should succeed")

	// Store PAT for reuse in tests
	var capturedPAT string

	t.Run("user creates PAT for application via CLI", func(t *testing.T) {
		// Login as testuser
		userLogin := exec.Command(shenctl, "auth", "login", "--username", "testuser", "--password", "testpass")
		require.NoError(t, userLogin.Run(), "testuser login should succeed")

		// Create PAT via CLI
		createCmd := exec.Command(shenctl, "token", "create", "my-token", "testapp")
		output, err := createCmd.CombinedOutput()
		require.NoError(t, err, "create token should succeed: %s", string(output))

		// Verify token starts with expected prefix
		capturedPAT = strings.TrimSpace(string(output))
		assert.True(t, strings.HasPrefix(capturedPAT, "shen_pat_"), "PAT should have correct prefix")
	})

	t.Run("PAT is returned only once - list does not show plaintext", func(t *testing.T) {
		// List tokens should show metadata, not the plaintext PAT
		listCmd := exec.Command(shenctl, "token", "list")
		output, err := listCmd.CombinedOutput()
		require.NoError(t, err, "list tokens should succeed: %s", string(output))

		outputStr := string(output)
		assert.Contains(t, outputStr, "my-token", "should contain token name")
		assert.Contains(t, outputStr, "testapp", "should contain application name")
		assert.NotContains(t, outputStr, "shen_pat_", "should not show plaintext PAT in list")
	})

	t.Run("PAT is hashed in database", func(t *testing.T) {
		// Get the user
		user, err := queries.GetUserByUsername(t.Context(), "testuser")
		require.NoError(t, err)

		// Get the token from DB
		tokens, err := queries.ListTokensByUser(t.Context(), db.ListTokensByUserParams{
			UserID:   user.ID,
			Limit:    10,
			CursorID: 0,
		})
		require.NoError(t, err)
		require.NotEmpty(t, tokens, "should have at least one token")

		// Get the full token record to check hash
		token, err := queries.GetTokenByID(t.Context(), tokens[0].ID)
		require.NoError(t, err)

		// Verify the stored token is a SHA-256 hash (64 hex characters)
		assert.Len(t, token.HashedToken, 64, "stored token should be SHA-256 hash (64 hex chars)")
		assert.NotContains(t, token.HashedToken, "shen_pat_", "should not store plaintext token")

		// Verify we can match the captured PAT to the hash
		expectedHash := crypto.HashToken(capturedPAT)
		assert.Equal(t, expectedHash, token.HashedToken, "hash should match the captured PAT")
	})

	t.Run("exchange PAT for JWT", func(t *testing.T) {
		// Create a new token to capture the PAT for this test
		createCmd := exec.Command(shenctl, "token", "create", "jwt-test-token", "testapp")
		output, err := createCmd.CombinedOutput()
		require.NoError(t, err, "create token should succeed: %s", string(output))
		pat := strings.TrimSpace(string(output))

		// Call authorize endpoint directly with HTTP
		req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/authorize", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+pat)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode, "authorize should succeed")

		var authResp struct {
			Token string `json:"token"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&authResp))
		assert.NotEmpty(t, authResp.Token, "should return JWT")
	})

	t.Run("JWT contains correct claims", func(t *testing.T) {
		// Create a new token
		createCmd := exec.Command(shenctl, "token", "create", "claims-test-token", "testapp")
		output, err := createCmd.CombinedOutput()
		require.NoError(t, err)
		pat := strings.TrimSpace(string(output))

		// Exchange for JWT
		req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/authorize", nil)
		req.Header.Set("Authorization", "Bearer "+pat)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		var authResp struct {
			Token string `json:"token"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&authResp))

		// Parse JWT without verification to check claims
		token, _, err := jwt.NewParser().ParseUnverified(authResp.Token, jwt.MapClaims{})
		require.NoError(t, err)

		claims := token.Claims.(jwt.MapClaims)
		assert.Equal(t, "testuser", claims["sub"], "subject should be username")
		assert.Equal(t, "shen", claims["iss"], "issuer should be shen")
		assert.NotNil(t, claims["exp"], "should have expiration")
		assert.NotNil(t, claims["iat"], "should have issued at")
		assert.NotNil(t, claims["roles"], "should have roles claim")
		assert.NotNil(t, claims["groups"], "should have groups claim")

		// Check audience contains the app name
		aud, ok := claims["aud"]
		require.True(t, ok, "should have audience claim")
		// aud can be string or []string
		switch v := aud.(type) {
		case string:
			assert.Equal(t, "testapp", v)
		case []interface{}:
			found := false
			for _, a := range v {
				if a.(string) == "testapp" {
					found = true
					break
				}
			}
			assert.True(t, found, "audience should contain testapp")
		}
	})

	t.Run("JWT uses RS256 signing method", func(t *testing.T) {
		// Create a new token
		createCmd := exec.Command(shenctl, "token", "create", "sig-test-token", "testapp")
		output, err := createCmd.CombinedOutput()
		require.NoError(t, err, "create token should succeed: %s", string(output))
		pat := strings.TrimSpace(string(output))

		// Exchange for JWT
		req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/authorize", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+pat)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode, "authorize should succeed")

		var authResp struct {
			Token string `json:"token"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&authResp))

		// Parse and check signing method
		token, _, err := jwt.NewParser().ParseUnverified(authResp.Token, jwt.MapClaims{})
		require.NoError(t, err)
		assert.Equal(t, "RS256", token.Method.Alg(), "should use RS256 signing method")
	})

	t.Run("user in multiple groups gets deduplicated roles", func(t *testing.T) {
		// Login as admin
		adminLogin := exec.Command(shenctl, "auth", "login", "--username", "admin", "--password", "admin")
		require.NoError(t, adminLogin.Run())

		// Create second group with overlapping and different roles
		createGroup2 := exec.Command(shenctl, "group", "create", "secondgroup")
		require.NoError(t, createGroup2.Run())

		// Add viewer (duplicate) and operator (new) roles
		addRole1 := exec.Command(shenctl, "group", "add-role", "secondgroup", "testapp", "viewer")
		require.NoError(t, addRole1.Run())

		addRole2 := exec.Command(shenctl, "group", "add-role", "secondgroup", "testapp", "operator")
		require.NoError(t, addRole2.Run())

		// Add testuser to second group
		addUser := exec.Command(shenctl, "group", "add-users", "secondgroup", "testuser")
		require.NoError(t, addUser.Run())

		// Login as testuser
		userLogin := exec.Command(shenctl, "auth", "login", "--username", "testuser", "--password", "testpass")
		require.NoError(t, userLogin.Run())

		// Create PAT and exchange for JWT
		createCmd := exec.Command(shenctl, "token", "create", "multi-group-token", "testapp")
		output, err := createCmd.CombinedOutput()
		require.NoError(t, err, "create token should succeed: %s", string(output))
		pat := strings.TrimSpace(string(output))

		req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/authorize", nil)
		req.Header.Set("Authorization", "Bearer "+pat)
		resp, _ := http.DefaultClient.Do(req)
		defer resp.Body.Close()

		var authResp struct {
			Token string `json:"token"`
		}
		json.NewDecoder(resp.Body).Decode(&authResp)

		token, _, _ := jwt.NewParser().ParseUnverified(authResp.Token, jwt.MapClaims{})
		claims := token.Claims.(jwt.MapClaims)

		roles := claims["roles"].([]interface{})
		roleStrs := make([]string, len(roles))
		for i, r := range roles {
			roleStrs[i] = r.(string)
		}

		// Should have viewer and operator
		assert.Contains(t, roleStrs, "viewer", "should have viewer role")
		assert.Contains(t, roleStrs, "operator", "should have operator role")

		// Count occurrences of viewer - should be exactly 1 (deduplicated)
		viewerCount := 0
		for _, r := range roleStrs {
			if r == "viewer" {
				viewerCount++
			}
		}
		assert.Equal(t, 1, viewerCount, "viewer role should appear exactly once (deduplicated)")
	})

	t.Run("expired PAT cannot be exchanged", func(t *testing.T) {
		// Get user and app for direct DB token creation
		user, err := queries.GetUserByUsername(t.Context(), "testuser")
		require.NoError(t, err)

		app, err := queries.GetApplicationByName(t.Context(), "testapp")
		require.NoError(t, err)

		// Create an expired token directly in DB
		expiredTime := time.Now().Add(-24 * time.Hour)
		expiredPAT := "shen_pat_expiredtesttoken123456789"
		_, err = queries.CreateToken(t.Context(), db.CreateTokenParams{
			Name:          "expired-token",
			HashedToken:   crypto.HashToken(expiredPAT),
			UserID:        user.ID,
			ApplicationID: app.ID,
			ExpiresAt:     pgtype.Timestamptz{Time: expiredTime, Valid: true},
		})
		require.NoError(t, err)

		// Try to exchange the expired token
		req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/authorize", nil)
		req.Header.Set("Authorization", "Bearer "+expiredPAT)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expired PAT should return 401")
	})

	t.Run("PAT for inactive application fails", func(t *testing.T) {
		// Login as admin
		adminLogin := exec.Command(shenctl, "auth", "login", "--username", "admin", "--password", "admin")
		require.NoError(t, adminLogin.Run())

		// Create temp application
		createApp := exec.Command(shenctl, "application", "create", "tempapp")
		require.NoError(t, createApp.Run())

		// Add role so user can create PAT
		addRole := exec.Command(shenctl, "group", "add-role", "testgroup", "tempapp", "viewer")
		require.NoError(t, addRole.Run())

		// Login as user and create PAT
		userLogin := exec.Command(shenctl, "auth", "login", "--username", "testuser", "--password", "testpass")
		require.NoError(t, userLogin.Run())

		createToken := exec.Command(shenctl, "token", "create", "temp-app-token", "tempapp")
		output, err := createToken.CombinedOutput()
		require.NoError(t, err, "create token should succeed: %s", string(output))
		pat := strings.TrimSpace(string(output))

		// Admin deletes the application
		adminLogin = exec.Command(shenctl, "auth", "login", "--username", "admin", "--password", "admin")
		require.NoError(t, adminLogin.Run())

		deleteApp := exec.Command(shenctl, "application", "delete", "tempapp")
		require.NoError(t, deleteApp.Run())

		// Try to exchange PAT for deleted app
		req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/authorize", nil)
		req.Header.Set("Authorization", "Bearer "+pat)

		resp, _ := http.DefaultClient.Do(req)
		defer resp.Body.Close()

		// Should fail - app is inactive (could be 403 or 404 depending on implementation)
		assert.True(t, resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound,
			"inactive app should return 403 or 404, got %d", resp.StatusCode)
	})

	t.Run("user without group membership cannot create PAT", func(t *testing.T) {
		// Login as admin and create user not in any group
		adminLogin := exec.Command(shenctl, "auth", "login", "--username", "admin", "--password", "admin")
		require.NoError(t, adminLogin.Run())

		createUser := exec.Command(shenctl, "user", "create", "lonelyuser", "user", "--password", "lonelypass")
		require.NoError(t, createUser.Run())

		// Login as the new user
		userLogin := exec.Command(shenctl, "auth", "login", "--username", "lonelyuser", "--password", "lonelypass")
		require.NoError(t, userLogin.Run())

		// Try to create PAT - should fail (no roles for application)
		createToken := exec.Command(shenctl, "token", "create", "lonely-token", "testapp")
		err := createToken.Run()
		assert.Error(t, err, "user without group membership should not be able to create PAT")
	})

	t.Run("admin can list another user's tokens", func(t *testing.T) {
		// Login as admin
		adminLogin := exec.Command(shenctl, "auth", "login", "--username", "admin", "--password", "admin")
		require.NoError(t, adminLogin.Run())

		// List testuser's tokens
		listCmd := exec.Command(shenctl, "token", "list", "--user", "testuser")
		output, err := listCmd.CombinedOutput()
		require.NoError(t, err, "admin should be able to list another user's tokens: %s", string(output))

		outputStr := string(output)
		assert.Contains(t, outputStr, "my-token", "should show testuser's tokens")
	})

	t.Run("non-admin cannot list another user's tokens", func(t *testing.T) {
		// Login as testuser
		userLogin := exec.Command(shenctl, "auth", "login", "--username", "testuser", "--password", "testpass")
		require.NoError(t, userLogin.Run())

		// Try to list admin's tokens
		listCmd := exec.Command(shenctl, "token", "list", "--user", "admin")
		err := listCmd.Run()
		assert.Error(t, err, "non-admin should not be able to list another user's tokens")
	})

	t.Run("duplicate token name for same app fails", func(t *testing.T) {
		// Login as testuser
		userLogin := exec.Command(shenctl, "auth", "login", "--username", "testuser", "--password", "testpass")
		require.NoError(t, userLogin.Run())

		// Create a token
		createCmd := exec.Command(shenctl, "token", "create", "duplicate-token", "testapp")
		require.NoError(t, createCmd.Run(), "first token creation should succeed")

		// Try to create same token again
		createCmd2 := exec.Command(shenctl, "token", "create", "duplicate-token", "testapp")
		err := createCmd2.Run()
		assert.Error(t, err, "duplicate token name should fail")
	})
}
