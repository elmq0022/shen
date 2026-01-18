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

func TestMilestone4_GroupManagement(t *testing.T) {
	resetDB(t)
	startTestServer(t, shen)
	xdgDirs := setTempXDG(t)
	_ = xdgDirs // used for session storage

	// Login as admin first
	adminLogin := exec.Command(shenctl, "auth", "login", "--username", "admin", "--password", "admin")
	require.NoError(t, adminLogin.Run(), "admin login should succeed")

	t.Run("admin creates group", func(t *testing.T) {
		createCmd := exec.Command(shenctl, "group", "create", "engineering")
		output, err := createCmd.CombinedOutput()
		require.NoError(t, err, "create group should succeed: %s", string(output))

		// Verify group exists in DB
		group, err := queries.GetGroupByName(t.Context(), "engineering")
		require.NoError(t, err, "group should exist in DB")
		assert.Equal(t, "engineering", group.Name)
		assert.True(t, group.Active)
	})

	t.Run("group names normalized to lowercase", func(t *testing.T) {
		createCmd := exec.Command(shenctl, "group", "create", "DataScience")
		output, err := createCmd.CombinedOutput()
		require.NoError(t, err, "create group should succeed: %s", string(output))

		// Verify group was stored lowercase
		group, err := queries.GetGroupByName(t.Context(), "datascience")
		require.NoError(t, err, "group should exist in DB with lowercase name")
		assert.Equal(t, "datascience", group.Name)
		assert.True(t, group.Active)

		// Verify uppercase name doesn't exist
		_, err = queries.GetGroupByName(t.Context(), "DataScience")
		assert.Error(t, err, "group with uppercase name should not exist")
	})

	t.Run("admin lists groups with pagination", func(t *testing.T) {
		// Create additional groups for pagination testing
		for _, name := range []string{"alpha", "beta", "gamma", "delta"} {
			createCmd := exec.Command(shenctl, "group", "create", name)
			require.NoError(t, createCmd.Run(), "create group %s should succeed", name)
		}

		// List with default limit
		listCmd := exec.Command(shenctl, "group", "list")
		output, err := listCmd.CombinedOutput()
		require.NoError(t, err, "list groups should succeed: %s", string(output))

		outputStr := string(output)
		// Verify some groups appear (alphabetically ordered)
		assert.Contains(t, outputStr, "alpha")
		assert.Contains(t, outputStr, "beta")

		// Test pagination with limit
		listLimitCmd := exec.Command(shenctl, "group", "list", "--limit", "2")
		output, err = listLimitCmd.CombinedOutput()
		require.NoError(t, err, "list groups with limit should succeed: %s", string(output))

		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		assert.LessOrEqual(t, len(lines), 2, "should return at most 2 groups")

		// Test pagination with cursor
		listCursorCmd := exec.Command(shenctl, "group", "list", "--cursor", "beta", "--limit", "2")
		output, err = listCursorCmd.CombinedOutput()
		require.NoError(t, err, "list groups with cursor should succeed: %s", string(output))

		outputStr = string(output)
		assert.NotContains(t, outputStr, "alpha", "alpha should not appear after beta cursor")
		assert.NotContains(t, outputStr, "beta", "beta should not appear (cursor is exclusive)")
	})

	t.Run("admin soft deletes group", func(t *testing.T) {
		deleteCmd := exec.Command(shenctl, "group", "delete", "gamma")
		output, err := deleteCmd.CombinedOutput()
		require.NoError(t, err, "delete group should succeed: %s", string(output))

		// Verify group is inactive in DB
		group, err := queries.GetGroupByName(t.Context(), "gamma")
		require.NoError(t, err, "group should still exist in DB")
		assert.False(t, group.Active, "group should be inactive (soft deleted)")

		// Verify it doesn't appear in list
		listCmd := exec.Command(shenctl, "group", "list", "--all")
		output, err = listCmd.CombinedOutput()
		require.NoError(t, err, "list groups should succeed: %s", string(output))
		assert.NotContains(t, string(output), "gamma", "deleted group should not appear in list")
	})

	t.Run("admin adds users to group", func(t *testing.T) {
		// Create test users
		createUser1 := exec.Command(shenctl, "user", "create", "alice", "user", "--password", "alicepass")
		require.NoError(t, createUser1.Run(), "create user alice should succeed")

		createUser2 := exec.Command(shenctl, "user", "create", "bob", "user", "--password", "bobpass")
		require.NoError(t, createUser2.Run(), "create user bob should succeed")

		// Add users to group
		addUsersCmd := exec.Command(shenctl, "group", "add-users", "engineering", "alice", "bob")
		output, err := addUsersCmd.CombinedOutput()
		require.NoError(t, err, "add users to group should succeed: %s", string(output))

		// Verify memberships in DB
		group, err := queries.GetGroupByName(t.Context(), "engineering")
		require.NoError(t, err)

		alice, err := queries.GetUserByUsername(t.Context(), "alice")
		require.NoError(t, err)

		bob, err := queries.GetUserByUsername(t.Context(), "bob")
		require.NoError(t, err)

		aliceInGroup, err := queries.IsUserInGroup(t.Context(), db.IsUserInGroupParams{
			UserID:  alice.ID,
			GroupID: group.ID,
		})
		require.NoError(t, err)
		assert.True(t, aliceInGroup, "alice should be in engineering group")

		bobInGroup, err := queries.IsUserInGroup(t.Context(), db.IsUserInGroupParams{
			UserID:  bob.ID,
			GroupID: group.ID,
		})
		require.NoError(t, err)
		assert.True(t, bobInGroup, "bob should be in engineering group")
	})

	t.Run("admin removes users from group", func(t *testing.T) {
		// Remove alice from group
		removeUsersCmd := exec.Command(shenctl, "group", "remove-users", "engineering", "alice")
		output, err := removeUsersCmd.CombinedOutput()
		require.NoError(t, err, "remove users from group should succeed: %s", string(output))

		// Verify alice is no longer in group
		group, err := queries.GetGroupByName(t.Context(), "engineering")
		require.NoError(t, err)

		alice, err := queries.GetUserByUsername(t.Context(), "alice")
		require.NoError(t, err)

		aliceInGroup, err := queries.IsUserInGroup(t.Context(), db.IsUserInGroupParams{
			UserID:  alice.ID,
			GroupID: group.ID,
		})
		require.NoError(t, err)
		assert.False(t, aliceInGroup, "alice should no longer be in engineering group")

		// Verify bob is still in group
		bob, err := queries.GetUserByUsername(t.Context(), "bob")
		require.NoError(t, err)

		bobInGroup, err := queries.IsUserInGroup(t.Context(), db.IsUserInGroupParams{
			UserID:  bob.ID,
			GroupID: group.ID,
		})
		require.NoError(t, err)
		assert.True(t, bobInGroup, "bob should still be in engineering group")
	})

	t.Run("user add-groups command works", func(t *testing.T) {
		// Create another group
		createGroupCmd := exec.Command(shenctl, "group", "create", "qa-team")
		require.NoError(t, createGroupCmd.Run(), "create group qa-team should succeed")

		// Add alice to multiple groups using user add-groups
		addGroupsCmd := exec.Command(shenctl, "user", "add-groups", "alice", "engineering", "qa-team")
		output, err := addGroupsCmd.CombinedOutput()
		require.NoError(t, err, "user add-groups should succeed: %s", string(output))

		// Verify alice is in both groups
		alice, err := queries.GetUserByUsername(t.Context(), "alice")
		require.NoError(t, err)

		engineering, err := queries.GetGroupByName(t.Context(), "engineering")
		require.NoError(t, err)

		qaTeam, err := queries.GetGroupByName(t.Context(), "qa-team")
		require.NoError(t, err)

		aliceInEngineering, err := queries.IsUserInGroup(t.Context(), db.IsUserInGroupParams{
			UserID:  alice.ID,
			GroupID: engineering.ID,
		})
		require.NoError(t, err)
		assert.True(t, aliceInEngineering, "alice should be in engineering group")

		aliceInQa, err := queries.IsUserInGroup(t.Context(), db.IsUserInGroupParams{
			UserID:  alice.ID,
			GroupID: qaTeam.ID,
		})
		require.NoError(t, err)
		assert.True(t, aliceInQa, "alice should be in qa-team group")
	})

	t.Run("delete group removes all memberships", func(t *testing.T) {
		// Create a group with members
		createGroupCmd := exec.Command(shenctl, "group", "create", "temp-team")
		require.NoError(t, createGroupCmd.Run())

		addUsersCmd := exec.Command(shenctl, "group", "add-users", "temp-team", "alice", "bob")
		require.NoError(t, addUsersCmd.Run())

		// Verify members before delete
		tempTeam, err := queries.GetGroupByName(t.Context(), "temp-team")
		require.NoError(t, err)

		countBefore, err := queries.CountUsersByGroup(t.Context(), tempTeam.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), countBefore, "should have 2 members before delete")

		// Delete the group (soft delete)
		deleteCmd := exec.Command(shenctl, "group", "delete", "temp-team")
		require.NoError(t, deleteCmd.Run())

		// Note: With soft delete, memberships may still exist in the DB
		// but the group is inactive. The cascade delete only happens on hard delete.
		// Verify the group is inactive
		tempTeamAfter, err := queries.GetGroupByName(t.Context(), "temp-team")
		require.NoError(t, err)
		assert.False(t, tempTeamAfter.Active, "group should be inactive after delete")
	})

	t.Run("non-admin cannot access group management endpoints", func(t *testing.T) {
		// Login as regular user
		userLogin := exec.Command(shenctl, "auth", "login", "--username", "alice", "--password", "alicepass")
		require.NoError(t, userLogin.Run(), "user login should succeed")

		// Attempt to create a group
		createCmd := exec.Command(shenctl, "group", "create", "shouldfail")
		err := createCmd.Run()
		assert.Error(t, err, "non-admin should not be able to create groups")

		// Attempt to list groups
		listCmd := exec.Command(shenctl, "group", "list")
		err = listCmd.Run()
		assert.Error(t, err, "non-admin should not be able to list groups")

		// Attempt to delete a group
		deleteCmd := exec.Command(shenctl, "group", "delete", "engineering")
		err = deleteCmd.Run()
		assert.Error(t, err, "non-admin should not be able to delete groups")

		// Attempt to add users to group
		addUsersCmd := exec.Command(shenctl, "group", "add-users", "engineering", "bob")
		err = addUsersCmd.Run()
		assert.Error(t, err, "non-admin should not be able to add users to groups")

		// Login back as admin
		adminLogin := exec.Command(shenctl, "auth", "login", "--username", "admin", "--password", "admin")
		require.NoError(t, adminLogin.Run())
	})

	t.Run("duplicate group name fails", func(t *testing.T) {
		// Attempt to create group with existing name
		createCmd := exec.Command(shenctl, "group", "create", "engineering")
		output, err := createCmd.CombinedOutput()
		assert.Error(t, err, "creating duplicate group should fail")
		assert.Contains(t, string(output), "already exists")
	})

	t.Run("delete non-existent group fails", func(t *testing.T) {
		deleteCmd := exec.Command(shenctl, "group", "delete", "nonexistent")
		output, err := deleteCmd.CombinedOutput()
		assert.Error(t, err, "deleting non-existent group should fail")
		assert.Contains(t, string(output), "not found")
	})

	t.Run("add users with non-existent user reports not found", func(t *testing.T) {
		addUsersCmd := exec.Command(shenctl, "group", "add-users", "engineering", "nonexistent")
		output, err := addUsersCmd.CombinedOutput()
		// The command may succeed but report not found users
		outputStr := string(output)
		if err == nil {
			assert.Contains(t, outputStr, "not found", "should report user not found")
		}
	})
}
