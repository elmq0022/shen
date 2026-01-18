//go:build cli_integration

package integration

import (
	"os/exec"
	"strings"
	"testing"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMilestone5_GroupRoleAssignments(t *testing.T) {
	resetDB(t)
	startTestServer(t, shen)
	xdgDirs := setTempXDG(t)
	_ = xdgDirs // used for session storage

	// Login as admin first
	adminLogin := exec.Command(shenctl, "auth", "login", "--username", "admin", "--password", "admin")
	require.NoError(t, adminLogin.Run(), "admin login should succeed")

	// Create test groups and applications
	createGroupCmd := exec.Command(shenctl, "group", "create", "engineering")
	require.NoError(t, createGroupCmd.Run(), "create group engineering should succeed")

	createGroupCmd2 := exec.Command(shenctl, "group", "create", "ops")
	require.NoError(t, createGroupCmd2.Run(), "create group ops should succeed")

	createAppCmd := exec.Command(shenctl, "application", "create", "myapp")
	require.NoError(t, createAppCmd.Run(), "create application myapp should succeed")

	createAppCmd2 := exec.Command(shenctl, "application", "create", "otherapp")
	require.NoError(t, createAppCmd2.Run(), "create application otherapp should succeed")

	t.Run("admin assigns role to group for application", func(t *testing.T) {
		addRoleCmd := exec.Command(shenctl, "group", "add-role", "engineering", "myapp", "viewer")
		output, err := addRoleCmd.CombinedOutput()
		require.NoError(t, err, "add role should succeed: %s", string(output))

		// Verify role assignment in DB
		group, err := queries.GetGroupByName(t.Context(), "engineering")
		require.NoError(t, err)

		app, err := queries.GetApplicationByName(t.Context(), "myapp")
		require.NoError(t, err)

		viewerRole, err := queries.GetApplicationRoleByName(t.Context(), "viewer")
		require.NoError(t, err)

		_, err = queries.GetGroupApplicationRole(t.Context(), db.GetGroupApplicationRoleParams{
			GroupID:       group.ID,
			ApplicationID: app.ID,
			RoleID:        viewerRole.ID,
		})
		require.NoError(t, err, "role assignment should exist in DB")
	})

	t.Run("admin assigns multiple roles to same group for same application", func(t *testing.T) {
		// Add another role (operator) for same group-app
		addRoleCmd := exec.Command(shenctl, "group", "add-role", "engineering", "myapp", "operator")
		output, err := addRoleCmd.CombinedOutput()
		require.NoError(t, err, "add second role should succeed: %s", string(output))

		// Verify both roles exist
		group, err := queries.GetGroupByName(t.Context(), "engineering")
		require.NoError(t, err)

		count, err := queries.CountGroupApplicationRolesByGroup(t.Context(), group.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(2), "group should have at least 2 role assignments")
	})

	t.Run("duplicate role assignment is idempotent", func(t *testing.T) {
		// Try to add viewer role again (should succeed without error)
		addRoleCmd := exec.Command(shenctl, "group", "add-role", "engineering", "myapp", "viewer")
		output, err := addRoleCmd.CombinedOutput()
		require.NoError(t, err, "duplicate role assignment should succeed (idempotent): %s", string(output))
	})

	t.Run("admin lists roles for group", func(t *testing.T) {
		// Add roles for a different application too
		addRoleCmd := exec.Command(shenctl, "group", "add-role", "engineering", "otherapp", "admin")
		require.NoError(t, addRoleCmd.Run())

		// List all roles for the group
		listRolesCmd := exec.Command(shenctl, "group", "list-roles", "engineering")
		output, err := listRolesCmd.CombinedOutput()
		require.NoError(t, err, "list roles should succeed: %s", string(output))

		outputStr := string(output)
		assert.Contains(t, outputStr, "myapp", "should contain myapp")
		assert.Contains(t, outputStr, "viewer", "should contain viewer role")
		assert.Contains(t, outputStr, "operator", "should contain operator role")
		assert.Contains(t, outputStr, "otherapp", "should contain otherapp")
		assert.Contains(t, outputStr, "admin", "should contain admin role")
	})

	t.Run("admin lists roles filtered by application", func(t *testing.T) {
		listRolesCmd := exec.Command(shenctl, "group", "list-roles", "engineering", "myapp")
		output, err := listRolesCmd.CombinedOutput()
		require.NoError(t, err, "list roles filtered by app should succeed: %s", string(output))

		outputStr := string(output)
		assert.Contains(t, outputStr, "myapp", "should contain myapp")
		assert.Contains(t, outputStr, "viewer", "should contain viewer role")
		assert.Contains(t, outputStr, "operator", "should contain operator role")
		assert.NotContains(t, outputStr, "otherapp", "should not contain otherapp when filtering by myapp")
	})

	t.Run("admin removes role from group", func(t *testing.T) {
		removeRoleCmd := exec.Command(shenctl, "group", "remove-role", "engineering", "myapp", "operator")
		output, err := removeRoleCmd.CombinedOutput()
		require.NoError(t, err, "remove role should succeed: %s", string(output))

		// Verify role is removed from DB
		group, err := queries.GetGroupByName(t.Context(), "engineering")
		require.NoError(t, err)

		app, err := queries.GetApplicationByName(t.Context(), "myapp")
		require.NoError(t, err)

		operatorRole, err := queries.GetApplicationRoleByName(t.Context(), "operator")
		require.NoError(t, err)

		_, err = queries.GetGroupApplicationRole(t.Context(), db.GetGroupApplicationRoleParams{
			GroupID:       group.ID,
			ApplicationID: app.ID,
			RoleID:        operatorRole.ID,
		})
		assert.Error(t, err, "operator role should no longer exist in DB")

		// Verify viewer role still exists
		viewerRole, err := queries.GetApplicationRoleByName(t.Context(), "viewer")
		require.NoError(t, err)

		_, err = queries.GetGroupApplicationRole(t.Context(), db.GetGroupApplicationRoleParams{
			GroupID:       group.ID,
			ApplicationID: app.ID,
			RoleID:        viewerRole.ID,
		})
		require.NoError(t, err, "viewer role should still exist")
	})

	t.Run("invalid role name fails", func(t *testing.T) {
		addRoleCmd := exec.Command(shenctl, "group", "add-role", "engineering", "myapp", "invalidrole")
		output, err := addRoleCmd.CombinedOutput()
		assert.Error(t, err, "invalid role should fail")
		assert.Contains(t, string(output), "invalid role", "error should mention invalid role")
	})

	t.Run("role for non-existent application fails", func(t *testing.T) {
		addRoleCmd := exec.Command(shenctl, "group", "add-role", "engineering", "nonexistentapp", "viewer")
		output, err := addRoleCmd.CombinedOutput()
		assert.Error(t, err, "non-existent application should fail")
		assert.Contains(t, string(output), "not found", "error should mention not found")
	})

	t.Run("role for non-existent group fails", func(t *testing.T) {
		addRoleCmd := exec.Command(shenctl, "group", "add-role", "nonexistentgroup", "myapp", "viewer")
		output, err := addRoleCmd.CombinedOutput()
		assert.Error(t, err, "non-existent group should fail")
		assert.Contains(t, string(output), "not found", "error should mention not found")
	})

	t.Run("user in multiple groups gets all roles as array", func(t *testing.T) {
		// Create a test user
		createUserCmd := exec.Command(shenctl, "user", "create", "testuser", "user", "--password", "testpass")
		require.NoError(t, createUserCmd.Run(), "create user should succeed")

		// Add roles to ops group for myapp
		addRoleCmd := exec.Command(shenctl, "group", "add-role", "ops", "myapp", "auditor")
		require.NoError(t, addRoleCmd.Run())

		addRoleCmd2 := exec.Command(shenctl, "group", "add-role", "ops", "myapp", "operator")
		require.NoError(t, addRoleCmd2.Run())

		// engineering group already has viewer role for myapp

		// Add user to both groups
		addUsersCmd := exec.Command(shenctl, "group", "add-users", "engineering", "testuser")
		require.NoError(t, addUsersCmd.Run())

		addUsersCmd2 := exec.Command(shenctl, "group", "add-users", "ops", "testuser")
		require.NoError(t, addUsersCmd2.Run())

		// Query the user's roles directly from DB
		user, err := queries.GetUserByUsername(t.Context(), "testuser")
		require.NoError(t, err)

		app, err := queries.GetApplicationByName(t.Context(), "myapp")
		require.NoError(t, err)

		roles, err := queries.GetUserApplicationRoles(t.Context(), db.GetUserApplicationRolesParams{
			UserID:        user.ID,
			ApplicationID: app.ID,
		})
		require.NoError(t, err)

		// User should have roles from both groups: viewer (engineering), auditor (ops), operator (ops)
		assert.Len(t, roles, 3, "user should have 3 roles from 2 groups")

		roleNames := make([]string, len(roles))
		for i, r := range roles {
			roleNames[i] = r.Name
		}

		assert.Contains(t, roleNames, "viewer", "should have viewer from engineering")
		assert.Contains(t, roleNames, "auditor", "should have auditor from ops")
		assert.Contains(t, roleNames, "operator", "should have operator from ops")
	})

	t.Run("list-roles with pagination", func(t *testing.T) {
		// Create a group with many roles
		createGroupCmd := exec.Command(shenctl, "group", "create", "biggroup")
		require.NoError(t, createGroupCmd.Run())

		// Add all 5 roles for myapp
		for _, role := range []string{"authenticated", "viewer", "auditor", "operator", "admin"} {
			addRoleCmd := exec.Command(shenctl, "group", "add-role", "biggroup", "myapp", role)
			require.NoError(t, addRoleCmd.Run())
		}

		// List with limit
		listRolesCmd := exec.Command(shenctl, "group", "list-roles", "biggroup", "myapp", "--limit", "2")
		output, err := listRolesCmd.CombinedOutput()
		require.NoError(t, err, "list roles with limit should succeed: %s", string(output))

		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		assert.LessOrEqual(t, len(lines), 2, "should return at most 2 roles")

		// List all roles
		listAllCmd := exec.Command(shenctl, "group", "list-roles", "biggroup", "myapp", "--all")
		output, err = listAllCmd.CombinedOutput()
		require.NoError(t, err, "list all roles should succeed: %s", string(output))

		lines = strings.Split(strings.TrimSpace(string(output)), "\n")
		assert.Equal(t, 5, len(lines), "should return all 5 roles")
	})

	t.Run("non-admin cannot manage group roles", func(t *testing.T) {
		// Create a regular user
		createUserCmd := exec.Command(shenctl, "user", "create", "regularuser", "user", "--password", "userpass")
		require.NoError(t, createUserCmd.Run())

		// Login as regular user
		userLogin := exec.Command(shenctl, "auth", "login", "--username", "regularuser", "--password", "userpass")
		require.NoError(t, userLogin.Run(), "user login should succeed")

		// Attempt to add role
		addRoleCmd := exec.Command(shenctl, "group", "add-role", "engineering", "myapp", "admin")
		err := addRoleCmd.Run()
		assert.Error(t, err, "non-admin should not be able to add roles")

		// Attempt to remove role
		removeRoleCmd := exec.Command(shenctl, "group", "remove-role", "engineering", "myapp", "viewer")
		err = removeRoleCmd.Run()
		assert.Error(t, err, "non-admin should not be able to remove roles")

		// Attempt to list roles
		listRolesCmd := exec.Command(shenctl, "group", "list-roles", "engineering")
		err = listRolesCmd.Run()
		assert.Error(t, err, "non-admin should not be able to list roles")

		// Login back as admin
		adminLogin := exec.Command(shenctl, "auth", "login", "--username", "admin", "--password", "admin")
		require.NoError(t, adminLogin.Run())
	})
}
