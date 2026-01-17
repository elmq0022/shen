//go:build cli_integration

package integration

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMileStone2_UserAdministration(t *testing.T) {
	resetDB(t)
	startTestServer(t, shen)
	xdgDirs := setTempXDG(t)
	_ = xdgDirs // used for session storage

	// Login as admin first
	adminLogin := exec.Command(shenctl, "auth", "login", "--username", "admin", "--password", "admin")
	require.NoError(t, adminLogin.Run(), "admin login should succeed")

	t.Run("admin creates new user", func(t *testing.T) {
		createCmd := exec.Command(shenctl, "user", "create", "testuser", "user", "--password", "testpass")
		output, err := createCmd.CombinedOutput()
		require.NoError(t, err, "create user should succeed: %s", string(output))

		// Verify user exists in DB
		user, err := queries.GetUserByUsername(t.Context(), "testuser")
		require.NoError(t, err, "user should exist in DB")
		assert.Equal(t, "testuser", user.Username)
		assert.True(t, user.Active)

		// Verify role
		role, err := queries.GetRoleByName(t.Context(), "user")
		require.NoError(t, err)
		assert.Equal(t, role.ID, user.Role)
	})

	t.Run("admin lists users", func(t *testing.T) {
		listCmd := exec.Command(shenctl, "user", "list")
		output, err := listCmd.CombinedOutput()
		require.NoError(t, err, "list users should succeed: %s", string(output))

		outputStr := string(output)
		assert.Contains(t, outputStr, "admin")
		assert.Contains(t, outputStr, "testuser")
	})

	t.Run("admin updates user role", func(t *testing.T) {
		updateCmd := exec.Command(shenctl, "user", "update", "testuser", "--role", "admin")
		output, err := updateCmd.CombinedOutput()
		require.NoError(t, err, "update role should succeed: %s", string(output))

		// Verify role changed in DB
		user, err := queries.GetUserByUsername(t.Context(), "testuser")
		require.NoError(t, err)

		adminRole, err := queries.GetRoleByName(t.Context(), "admin")
		require.NoError(t, err)
		assert.Equal(t, adminRole.ID, user.Role)

		// Reset back to user role for subsequent tests
		resetCmd := exec.Command(shenctl, "user", "update", "testuser", "--role", "user")
		require.NoError(t, resetCmd.Run())
	})

	t.Run("user updates own password", func(t *testing.T) {
		// Login as testuser
		testUserLogin := exec.Command(shenctl, "auth", "login", "--username", "testuser", "--password", "testpass")
		require.NoError(t, testUserLogin.Run(), "testuser login should succeed")

		// Update own password
		updateCmd := exec.Command(shenctl, "user", "update", "testuser", "--password", "--new-password", "newpass")
		output, err := updateCmd.CombinedOutput()
		require.NoError(t, err, "update own password should succeed: %s", string(output))

		// Verify new password works by logging in
		newPassLogin := exec.Command(shenctl, "auth", "login", "--username", "testuser", "--password", "newpass")
		require.NoError(t, newPassLogin.Run(), "login with new password should succeed")

		// Verify old password fails
		oldPassLogin := exec.Command(shenctl, "auth", "login", "--username", "testuser", "--password", "testpass")
		err = oldPassLogin.Run()
		assert.Error(t, err, "login with old password should fail")

		// Login back as admin for subsequent tests
		adminLogin := exec.Command(shenctl, "auth", "login", "--username", "admin", "--password", "admin")
		require.NoError(t, adminLogin.Run())
	})

	t.Run("admin updates another users' password", func(t *testing.T) {
		// Admin is already logged in
		updateCmd := exec.Command(shenctl, "user", "update", "testuser", "--password", "--new-password", "adminsetpass")
		output, err := updateCmd.CombinedOutput()
		require.NoError(t, err, "admin should be able to update other user's password: %s", string(output))

		// Verify new password works
		newPassLogin := exec.Command(shenctl, "auth", "login", "--username", "testuser", "--password", "adminsetpass")
		require.NoError(t, newPassLogin.Run(), "login with admin-set password should succeed")

		// Login back as admin
		adminLogin := exec.Command(shenctl, "auth", "login", "--username", "admin", "--password", "admin")
		require.NoError(t, adminLogin.Run())
	})

	t.Run("user cannot update another user's password", func(t *testing.T) {
		// Create another user
		createCmd := exec.Command(shenctl, "user", "create", "anotheruser", "user", "--password", "anotherpass")
		require.NoError(t, createCmd.Run())

		// Login as testuser (non-admin)
		testUserLogin := exec.Command(shenctl, "auth", "login", "--username", "testuser", "--password", "adminsetpass")
		require.NoError(t, testUserLogin.Run())

		// Attempt to update another user's password
		updateCmd := exec.Command(shenctl, "user", "update", "anotheruser", "--password", "--new-password", "hackedpass")
		output, err := updateCmd.CombinedOutput()
		assert.Error(t, err, "non-admin should not be able to update another user's password")
		assert.True(t, strings.Contains(string(output), "Error") || err != nil)

		// Login back as admin
		adminLogin := exec.Command(shenctl, "auth", "login", "--username", "admin", "--password", "admin")
		require.NoError(t, adminLogin.Run())
	})

	t.Run("admin soft deletes user", func(t *testing.T) {
		deleteCmd := exec.Command(shenctl, "user", "delete", "anotheruser")
		output, err := deleteCmd.CombinedOutput()
		require.NoError(t, err, "delete user should succeed: %s", string(output))

		// Verify user is inactive in DB
		user, err := queries.GetUserByUsername(t.Context(), "anotheruser")
		require.NoError(t, err, "user should still exist in DB")
		assert.False(t, user.Active, "user should be inactive (soft deleted)")
	})

	t.Run("non-admin cannot access user management endpoints", func(t *testing.T) {
		// Login as testuser (non-admin)
		testUserLogin := exec.Command(shenctl, "auth", "login", "--username", "testuser", "--password", "adminsetpass")
		require.NoError(t, testUserLogin.Run())

		// Attempt to create a user
		createCmd := exec.Command(shenctl, "user", "create", "shouldfail", "user", "--password", "pass")
		err := createCmd.Run()
		assert.Error(t, err, "non-admin should not be able to create users")

		// Attempt to list users
		listCmd := exec.Command(shenctl, "user", "list")
		err = listCmd.Run()
		assert.Error(t, err, "non-admin should not be able to list users")

		// Attempt to delete a user
		deleteCmd := exec.Command(shenctl, "user", "delete", "testuser")
		err = deleteCmd.Run()
		assert.Error(t, err, "non-admin should not be able to delete users")

		// Login back as admin
		adminLogin := exec.Command(shenctl, "auth", "login", "--username", "admin", "--password", "admin")
		require.NoError(t, adminLogin.Run())
	})

	t.Run("service accounts cannot login to shen", func(t *testing.T) {
		// Create a service account
		createCmd := exec.Command(shenctl, "user", "create", "svc-account", "service")
		output, err := createCmd.CombinedOutput()
		require.NoError(t, err, "create service account should succeed: %s", string(output))

		// Verify service account exists
		user, err := queries.GetUserByUsername(t.Context(), "svc-account")
		require.NoError(t, err)
		serviceRole, err := queries.GetRoleByName(t.Context(), "service")
		require.NoError(t, err)
		assert.Equal(t, serviceRole.ID, user.Role)

		// Attempt to login as service account (should fail)
		loginCmd := exec.Command(shenctl, "auth", "login", "--username", "svc-account", "--password", "anything")
		err = loginCmd.Run()
		assert.Error(t, err, "service account should not be able to login")
	})
}
