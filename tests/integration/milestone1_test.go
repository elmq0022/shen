//go:build cli_integration

package integration

import (
	"testing"
)

// TestMilestone1_BootstrapAndAdminAuthentication orchestrates all Milestone 1 tests
// Goal: Set up the foundation - database, bootstrap process, and admin login via CLI
func TestMilestone1_BootstrapAndAdminAuthentication(t *testing.T) {
	// Setup: Reset database and start server
	resetDB(t)
	startTestServer(t, shen)
	setTempXDGConfig(t)

	t.Run("Bootstrap creates admin user and JWT keys", func(t *testing.T) {
		// TODO: Verify that bootstrap process created:
		// - Default admin user in database
		// - RSA key pair for JWT signing
		// - Keys are stored in database in PEM format
		t.Skip("Not implemented yet")
	})

	t.Run("Login with default admin credentials", func(t *testing.T) {
		// TODO: Test successful login with admin credentials
		// - Run: shenctl auth login
		// - Provide default admin username/password
		// - Verify successful authentication response
		// - Verify session token is returned
		t.Skip("Not implemented yet")
	})

	t.Run("Session token stored and reused by CLI", func(t *testing.T) {
		// TODO: Verify session token persistence
		// - Login as admin
		// - Verify session token is stored in ~/.cache/shenctl/session
		// - Make authenticated request
		// - Verify stored token is sent in request headers
		t.Skip("Not implemented yet")
	})

	t.Run("Invalid credentials return 401", func(t *testing.T) {
		// TODO: Test authentication failure
		// - Run: shenctl auth login with wrong credentials
		// - Verify 401 Unauthorized response
		// - Verify no session token is stored
		// - Verify appropriate error message displayed
		t.Skip("Not implemented yet")
	})

	t.Run("JWKS endpoint returns valid JWK", func(t *testing.T) {
		// TODO: Test JWKS endpoint
		// - Make GET request to /.well-known/jwks.json
		// - Verify response is valid JSON
		// - Verify response contains keys array
		// - Verify key format is valid JWK (includes kid, kty, use, alg, n, e)
		// - Verify public key can be parsed
		t.Skip("Not implemented yet")
	})

	t.Run("Logout revokes all of the user's active session tokens", func(t *testing.T) {
		// TODO: Test logout functionality
		// - Login as admin (create session)
		// - Verify session is valid (make authenticated request)
		// - Run: shenctl auth logout
		// - Verify logout was successful
		// - Verify session token is no longer valid (401 on authenticated request)
		// - Verify all user's sessions are revoked (if multiple sessions existed)
		t.Skip("Not implemented yet")
	})
}
