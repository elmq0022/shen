//go:build integration

package db_tests

import (
	"fmt"
	"slices"
	"testing"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/stretchr/testify/require"
)

// addUsersToGroup adds multiple users as members to a group.
func addUsersToGroup(t *testing.T, tdb *TestDB, users []db.ShenUser, group db.ShenGroup) {
	t.Helper()
	for _, user := range users {
		_, err := tdb.Queries.AddUserToGroup(tdb.Ctx, db.AddUserToGroupParams{
			UserID:  user.ID,
			GroupID: group.ID,
		})
		require.NoError(
			t, err,
			fmt.Sprintf("Failed to add user: %s to group: %s", user.Username, group.Name),
		)
	}
}

// addManagersToGroup adds multiple users as managers to a group.
func addManagersToGroup(t *testing.T, tdb *TestDB, users []db.ShenUser, group db.ShenGroup) {
	t.Helper()
	for _, user := range users {
		_, err := tdb.Queries.AddManagerToGroup(tdb.Ctx, db.AddManagerToGroupParams{
			UserID:  user.ID,
			GroupID: group.ID,
		})
		require.NoError(
			t, err,
			fmt.Sprintf("Failed to add manager: %s to group: %s", user.Username, group.Name),
		)
	}
}

// sortUsersByUsername sorts a slice of users by username in ascending order.
func sortUsersByUsername(users []db.ShenUser) {
	slices.SortFunc(users, func(a, b db.ShenUser) int {
		if a.Username < b.Username {
			return -1
		} else if a.Username == b.Username {
			return 0
		}
		return 1
	})
}

// sortGroupsByName sorts a slice of groups by name in ascending order.
func sortGroupsByName(groups []db.ShenGroup) {
	slices.SortFunc(groups, func(a, b db.ShenGroup) int {
		if a.Name < b.Name {
			return -1
		} else if a.Name == b.Name {
			return 0
		}
		return 1
	})
}

// setGroupApplicationPermissions sets permissions for multiple group-application pairs.
// Takes a slice of structs containing GroupID, ApplicationID, and PermissionID.
func setGroupApplicationPermissions(t *testing.T, tdb *TestDB, permissions []struct {
	GroupID       int32
	ApplicationID int32
	PermissionID  int32
}) {
	t.Helper()
	for _, perm := range permissions {
		_, err := tdb.Queries.SetGroupApplicationPermission(tdb.Ctx, db.SetGroupApplicationPermissionParams{
			GroupID:       perm.GroupID,
			ApplicationID: perm.ApplicationID,
			PermissionID:  perm.PermissionID,
		})
		require.NoError(t, err, "Failed to set group application permission")
	}
}
