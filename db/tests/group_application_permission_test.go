//go:build integration

package db_tests

import (
	"testing"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetGroupApplicationPermission(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	_, err := tdb.Queries.GetGroupApplicationPermission(tdb.Ctx, db.GetGroupApplicationPermissionParams{
		GroupID:       f.Group1.ID,
		ApplicationID: f.App1.ID,
	})
	require.Error(t, err, "Should not exist before creation")

	created, err := tdb.Queries.SetGroupApplicationPermission(tdb.Ctx, db.SetGroupApplicationPermissionParams{
		GroupID:       f.Group1.ID,
		ApplicationID: f.App1.ID,
		PermissionID:  PermissionViewer,
	})
	require.NoError(t, err, "Failed to create group application permission")

	fetched, err := tdb.Queries.GetGroupApplicationPermission(tdb.Ctx, db.GetGroupApplicationPermissionParams{
		GroupID:       f.Group1.ID,
		ApplicationID: f.App1.ID,
	})
	require.NoError(t, err, "Failed to get group application permission")
	assert.Equal(t, created, fetched)
	assert.Equal(t, PermissionViewer, fetched.PermissionID)

	updated, err := tdb.Queries.SetGroupApplicationPermission(tdb.Ctx, db.SetGroupApplicationPermissionParams{
		GroupID:       f.Group1.ID,
		ApplicationID: f.App1.ID,
		PermissionID:  PermissionAdmin,
	})
	require.NoError(t, err, "Failed to update group application permission")

	fetchedAfterUpdate, err := tdb.Queries.GetGroupApplicationPermission(tdb.Ctx, db.GetGroupApplicationPermissionParams{
		GroupID:       f.Group1.ID,
		ApplicationID: f.App1.ID,
	})
	require.NoError(t, err, "Failed to get updated group application permission")
	assert.Equal(t, updated, fetchedAfterUpdate)
	assert.Equal(t, created.ID, updated.ID, "ID should not change on upsert")
	assert.Equal(t, PermissionAdmin, fetchedAfterUpdate.PermissionID)
}

func TestGetGroupApplicationPermissionByID(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	created, err := tdb.Queries.SetGroupApplicationPermission(tdb.Ctx, db.SetGroupApplicationPermissionParams{
		GroupID:       f.Group1.ID,
		ApplicationID: f.App1.ID,
		PermissionID:  PermissionViewer,
	})
	require.NoError(t, err, "Failed to create group application permission")

	fetched, err := tdb.Queries.GetGroupApplicationPermissionByID(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to get group application permission by ID")
	assert.Equal(t, created, fetched)

	_, err = tdb.Queries.GetGroupApplicationPermissionByID(tdb.Ctx, 999)
	require.Error(t, err, "Should get error for non-existent ID")
}

func TestDeleteGroupApplicationPermission(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	created, err := tdb.Queries.SetGroupApplicationPermission(tdb.Ctx, db.SetGroupApplicationPermissionParams{
		GroupID:       f.Group1.ID,
		ApplicationID: f.App1.ID,
		PermissionID:  PermissionOperator,
	})
	require.NoError(t, err, "Failed to create group application permission")

	_, err = tdb.Queries.GetGroupApplicationPermissionByID(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to verify permission exists before deletion")

	err = tdb.Queries.DeleteGroupApplicationPermission(tdb.Ctx, db.DeleteGroupApplicationPermissionParams{
		GroupID:       created.GroupID,
		ApplicationID: created.ApplicationID,
	})
	require.NoError(t, err, "Failed to delete group application permission")

	_, err = tdb.Queries.GetGroupApplicationPermissionByID(tdb.Ctx, created.ID)
	require.Error(t, err, "Should get error when fetching deleted permission")
}

func TestGetUserApplicationPermission(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	_, err := tdb.Queries.SetGroupApplicationPermission(tdb.Ctx, db.SetGroupApplicationPermissionParams{
		GroupID:       f.Group1.ID,
		ApplicationID: f.App1.ID,
		PermissionID:  PermissionViewer,
	})
	require.NoError(t, err, "Failed to set Group1 application permission")

	_, err = tdb.Queries.SetGroupApplicationPermission(tdb.Ctx, db.SetGroupApplicationPermissionParams{
		GroupID:       f.Group2.ID,
		ApplicationID: f.App1.ID,
		PermissionID:  PermissionAdmin,
	})
	require.NoError(t, err, "Failed to set Group2 application permission")

	_, err = tdb.Queries.AddUserToGroup(tdb.Ctx, db.AddUserToGroupParams{
		UserID:  f.User1.ID,
		GroupID: f.Group1.ID,
	})
	require.NoError(t, err, "Failed to add User1 to Group1")

	_, err = tdb.Queries.AddUserToGroup(tdb.Ctx, db.AddUserToGroupParams{
		UserID:  f.User1.ID,
		GroupID: f.Group2.ID,
	})
	require.NoError(t, err, "Failed to add User1 to Group2")

	expected, err := tdb.Queries.GetPermissionByID(tdb.Ctx, PermissionAdmin)
	require.NoError(t, err, "Failed to get expected permission")

	actual, err := tdb.Queries.GetUserApplicationPermission(tdb.Ctx, db.GetUserApplicationPermissionParams{
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
	})
	require.NoError(t, err, "Failed to get user application permission")

	assert.Equal(t, expected, actual)
}

func TestGetUserApplicationPermissionEdgeCases(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	t.Run("user not in any groups", func(t *testing.T) {
		_, err := tdb.Queries.SetGroupApplicationPermission(tdb.Ctx, db.SetGroupApplicationPermissionParams{
			GroupID:       f.Group1.ID,
			ApplicationID: f.App1.ID,
			PermissionID:  PermissionViewer,
		})
		require.NoError(t, err, "Failed to set group application permission")

		_, err = tdb.Queries.GetUserApplicationPermission(tdb.Ctx, db.GetUserApplicationPermissionParams{
			UserID:        f.User1.ID,
			ApplicationID: f.App1.ID,
		})
		require.Error(t, err, "Should get error when user is not in any groups with permission")
	})

	t.Run("user in group but no permission for application", func(t *testing.T) {
		_, err := tdb.Queries.AddUserToGroup(tdb.Ctx, db.AddUserToGroupParams{
			UserID:  f.User1.ID,
			GroupID: f.Group1.ID,
		})
		require.NoError(t, err, "Failed to add user to group")

		_, err = tdb.Queries.GetUserApplicationPermission(tdb.Ctx, db.GetUserApplicationPermissionParams{
			UserID:        f.User1.ID,
			ApplicationID: f.App2.ID,
		})
		require.Error(t, err, "Should get error when user's groups have no permission for application")
	})

	t.Run("user in single group with permission", func(t *testing.T) {
		_, err := tdb.Queries.SetGroupApplicationPermission(tdb.Ctx, db.SetGroupApplicationPermissionParams{
			GroupID:       f.Group1.ID,
			ApplicationID: f.App2.ID,
			PermissionID:  PermissionOperator,
		})
		require.NoError(t, err, "Failed to set group application permission")

		expected, err := tdb.Queries.GetPermissionByID(tdb.Ctx, PermissionOperator)
		require.NoError(t, err, "Failed to get expected permission")

		actual, err := tdb.Queries.GetUserApplicationPermission(tdb.Ctx, db.GetUserApplicationPermissionParams{
			UserID:        f.User1.ID,
			ApplicationID: f.App2.ID,
		})
		require.NoError(t, err, "Failed to get user application permission")

		assert.Equal(t, expected, actual)
	})

	t.Run("user in multiple groups with same permission level", func(t *testing.T) {
		_, err := tdb.Queries.AddUserToGroup(tdb.Ctx, db.AddUserToGroupParams{
			UserID:  f.User2.ID,
			GroupID: f.Group1.ID,
		})
		require.NoError(t, err, "Failed to add User2 to Group1")

		_, err = tdb.Queries.AddUserToGroup(tdb.Ctx, db.AddUserToGroupParams{
			UserID:  f.User2.ID,
			GroupID: f.Group2.ID,
		})
		require.NoError(t, err, "Failed to add User2 to Group2")

		_, err = tdb.Queries.SetGroupApplicationPermission(tdb.Ctx, db.SetGroupApplicationPermissionParams{
			GroupID:       f.Group1.ID,
			ApplicationID: f.App1.ID,
			PermissionID:  PermissionAuditor,
		})
		require.NoError(t, err, "Failed to set Group1 application permission")

		_, err = tdb.Queries.SetGroupApplicationPermission(tdb.Ctx, db.SetGroupApplicationPermissionParams{
			GroupID:       f.Group2.ID,
			ApplicationID: f.App1.ID,
			PermissionID:  PermissionAuditor,
		})
		require.NoError(t, err, "Failed to set Group2 application permission")

		expected, err := tdb.Queries.GetPermissionByID(tdb.Ctx, PermissionAuditor)
		require.NoError(t, err, "Failed to get expected permission")

		actual, err := tdb.Queries.GetUserApplicationPermission(tdb.Ctx, db.GetUserApplicationPermissionParams{
			UserID:        f.User2.ID,
			ApplicationID: f.App1.ID,
		})
		require.NoError(t, err, "Failed to get user application permission")

		assert.Equal(t, expected, actual)
	})
}
