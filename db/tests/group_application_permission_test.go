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
