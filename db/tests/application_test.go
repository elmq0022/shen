//go:build integration

package db_tests

import (
	"testing"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAndGetApplication(t *testing.T) {
	tdb := SetupTestDB(t)

	// Create Application
	created, err := tdb.Queries.CreateApplication(tdb.Ctx, "paper-street-soap")
	require.NoError(t, err, "Failed to create application")

	// Get by ID
	fetchByID, err := tdb.Queries.GetApplicationByID(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to fetch by application ID")
	assert.Equal(t, created, fetchByID)

	// Get by Name
	fetchedByName, err := tdb.Queries.GetApplicationByName(tdb.Ctx, "paper-street-soap")
	require.NoError(t, err, "Failed to get application by name")
	assert.Equal(t, created, fetchedByName)

	// Deactivate App
	err = tdb.Queries.DeactivateApplication(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to deactivate application")

	deactivated, err := tdb.Queries.GetApplicationByID(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to get deactivated application")
	assert.False(t, deactivated.Active, "App should be deactivated")
}

func TestUpdateApplication(t *testing.T) {
	tdb := SetupTestDB(t)

	// Create Application
	created, err := tdb.Queries.CreateApplication(tdb.Ctx, "paper-street-soap")
	require.NoError(t, err, "Failed to create application")

	// Update application name and active status
	err = tdb.Queries.UpdateApplication(tdb.Ctx, db.UpdateApplicationParams{
		ID:     created.ID,
		Name:   "fight-club",
		Active: false,
	})
	require.NoError(t, err, "Failed to update application")

	// Verify update
	updated, err := tdb.Queries.GetApplicationByID(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to get updated application")
	assert.Equal(t, "fight-club", updated.Name)
	assert.False(t, updated.Active)
	assert.Equal(t, created.ID, updated.ID)
}

func TestDeleteApplication(t *testing.T) {
	tdb := SetupTestDB(t)

	// Create Application
	created, err := tdb.Queries.CreateApplication(tdb.Ctx, "paper-street-soap")
	require.NoError(t, err, "Failed to create application")

	// Delete application
	err = tdb.Queries.DeleteApplication(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to delete application")

	// Verify deletion - should get error when trying to fetch
	_, err = tdb.Queries.GetApplicationByID(tdb.Ctx, created.ID)
	assert.Error(t, err, "Should get error when fetching deleted application")
}
