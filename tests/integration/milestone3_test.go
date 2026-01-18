//go:build cli_integration

package integration

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMileStone3_ApplicationManagement(t *testing.T) {
	resetDB(t)
	startTestServer(t, shen)
	xdgDirs := setTempXDG(t)
	_ = xdgDirs // used for session storage

	// Login as admin first
	adminLogin := exec.Command(shenctl, "auth", "login", "--username", "admin", "--password", "admin")
	require.NoError(t, adminLogin.Run(), "admin login should succeed")

	t.Run("admin creates application", func(t *testing.T) {
		createCmd := exec.Command(shenctl, "application", "create", "testapp")
		output, err := createCmd.CombinedOutput()
		require.NoError(t, err, "create application should succeed: %s", string(output))

		// Verify application exists in DB
		app, err := queries.GetApplicationByName(t.Context(), "testapp")
		require.NoError(t, err, "application should exist in DB")
		assert.Equal(t, "testapp", app.Name)
		assert.True(t, app.Active)
	})

	t.Run("application names normalized to lowercase", func(t *testing.T) {
		createCmd := exec.Command(shenctl, "application", "create", "MyMixedCaseApp")
		output, err := createCmd.CombinedOutput()
		require.NoError(t, err, "create application should succeed: %s", string(output))

		// Verify application was stored lowercase
		app, err := queries.GetApplicationByName(t.Context(), "mymixedcaseapp")
		require.NoError(t, err, "application should exist in DB with lowercase name")
		assert.Equal(t, "mymixedcaseapp", app.Name)
		assert.True(t, app.Active)

		// Verify uppercase name doesn't exist
		_, err = queries.GetApplicationByName(t.Context(), "MyMixedCaseApp")
		assert.Error(t, err, "application with uppercase name should not exist")
	})

	t.Run("admin lists applications with pagination", func(t *testing.T) {
		// Create additional applications for pagination testing
		for _, name := range []string{"alpha", "beta", "gamma", "delta"} {
			createCmd := exec.Command(shenctl, "application", "create", name)
			require.NoError(t, createCmd.Run(), "create application %s should succeed", name)
		}

		// List with default limit
		listCmd := exec.Command(shenctl, "application", "list")
		output, err := listCmd.CombinedOutput()
		require.NoError(t, err, "list applications should succeed: %s", string(output))

		outputStr := string(output)
		// Verify some applications appear (alphabetically ordered)
		assert.Contains(t, outputStr, "alpha")
		assert.Contains(t, outputStr, "beta")

		// Test pagination with limit
		listLimitCmd := exec.Command(shenctl, "application", "list", "--limit", "2")
		output, err = listLimitCmd.CombinedOutput()
		require.NoError(t, err, "list applications with limit should succeed: %s", string(output))

		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		assert.LessOrEqual(t, len(lines), 2, "should return at most 2 applications")

		// Test pagination with cursor
		listCursorCmd := exec.Command(shenctl, "application", "list", "--cursor", "beta", "--limit", "2")
		output, err = listCursorCmd.CombinedOutput()
		require.NoError(t, err, "list applications with cursor should succeed: %s", string(output))

		outputStr = string(output)
		assert.NotContains(t, outputStr, "alpha", "alpha should not appear after beta cursor")
		assert.NotContains(t, outputStr, "beta", "beta should not appear (cursor is exclusive)")
	})

	t.Run("admin soft deletes application", func(t *testing.T) {
		deleteCmd := exec.Command(shenctl, "application", "delete", "gamma")
		output, err := deleteCmd.CombinedOutput()
		require.NoError(t, err, "delete application should succeed: %s", string(output))

		// Verify application is inactive in DB
		app, err := queries.GetApplicationByName(t.Context(), "gamma")
		require.NoError(t, err, "application should still exist in DB")
		assert.False(t, app.Active, "application should be inactive (soft deleted)")

		// Verify it doesn't appear in list
		listCmd := exec.Command(shenctl, "application", "list", "--all")
		output, err = listCmd.CombinedOutput()
		require.NoError(t, err, "list applications should succeed: %s", string(output))
		assert.NotContains(t, string(output), "gamma", "deleted application should not appear in list")
	})

	t.Run("non-admin cannot access application management endpoints", func(t *testing.T) {
		// Create a regular user
		createUserCmd := exec.Command(shenctl, "user", "create", "regularuser", "user", "--password", "userpass")
		require.NoError(t, createUserCmd.Run())

		// Login as regular user
		userLogin := exec.Command(shenctl, "auth", "login", "--username", "regularuser", "--password", "userpass")
		require.NoError(t, userLogin.Run(), "user login should succeed")

		// Attempt to create an application
		createCmd := exec.Command(shenctl, "application", "create", "shouldfail")
		err := createCmd.Run()
		assert.Error(t, err, "non-admin should not be able to create applications")

		// Attempt to list applications
		listCmd := exec.Command(shenctl, "application", "list")
		err = listCmd.Run()
		assert.Error(t, err, "non-admin should not be able to list applications")

		// Attempt to delete an application
		deleteCmd := exec.Command(shenctl, "application", "delete", "testapp")
		err = deleteCmd.Run()
		assert.Error(t, err, "non-admin should not be able to delete applications")

		// Login back as admin
		adminLogin := exec.Command(shenctl, "auth", "login", "--username", "admin", "--password", "admin")
		require.NoError(t, adminLogin.Run())
	})

	t.Run("duplicate application name fails", func(t *testing.T) {
		// Attempt to create application with existing name
		createCmd := exec.Command(shenctl, "application", "create", "testapp")
		output, err := createCmd.CombinedOutput()
		assert.Error(t, err, "creating duplicate application should fail")
		assert.Contains(t, string(output), "already exists")
	})

	t.Run("delete non-existent application fails", func(t *testing.T) {
		deleteCmd := exec.Command(shenctl, "application", "delete", "nonexistent")
		output, err := deleteCmd.CombinedOutput()
		assert.Error(t, err, "deleting non-existent application should fail")
		assert.Contains(t, string(output), "not found")
	})
}
