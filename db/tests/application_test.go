//go:build integration

package db_tests

import (
	"slices"
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

func TestListActiveApplications(t *testing.T) {
	tdb := SetupTestDB(t)

	appNames := []string{
		"microsoft",
		"planet-starbucks",
		"fedex",
	}

	for _, name := range appNames {
		_, err := tdb.Queries.CreateApplication(tdb.Ctx, name)
		require.NoError(t, err, "Failed to create application")
	}

	slices.Sort(appNames)

	// list all active apps
	apps, err := tdb.Queries.ListActiveApplications(tdb.Ctx, db.ListActiveApplicationsParams{
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to list active applications")
	assert.Equal(t, 3, len(apps))
	assert.Equal(t, appNames[0], apps[0].Name)
	assert.Equal(t, appNames[1], apps[1].Name)
	assert.Equal(t, appNames[2], apps[2].Name)

	// deactivate middle app and verify filtering
	appToDeactivate, err := tdb.Queries.GetApplicationByName(tdb.Ctx, appNames[1])
	require.NoError(t, err, "Failed to get app to deactivate")
	err = tdb.Queries.DeactivateApplication(tdb.Ctx, appToDeactivate.ID)
	require.NoError(t, err, "Failed to deactivate application")

	apps, err = tdb.Queries.ListActiveApplications(tdb.Ctx, db.ListActiveApplicationsParams{
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to list active apps after deactivation")
	assert.Equal(t, 2, len(apps))
	assert.Equal(t, appNames[0], apps[0].Name)
	assert.Equal(t, appNames[2], apps[1].Name)
}

func TestListApplications(t *testing.T) {
	tdb := SetupTestDB(t)

	appNames := []string{
		"blockbuster-video",
		"compuserve",
		"burger-king",
	}

	for _, name := range appNames {
		_, err := tdb.Queries.CreateApplication(tdb.Ctx, name)
		require.NoError(t, err, "Failed to create application")
	}

	// deactivate one app
	appToDeactivate, err := tdb.Queries.GetApplicationByName(tdb.Ctx, "compuserve")
	require.NoError(t, err, "Failed to get app to deactivate")
	err = tdb.Queries.DeactivateApplication(tdb.Ctx, appToDeactivate.ID)
	require.NoError(t, err, "Failed to deactivate application")

	slices.Sort(appNames)

	// ListApplications should return all apps including deactivated
	allApps, err := tdb.Queries.ListApplications(tdb.Ctx, db.ListApplicationsParams{
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to list all applications")
	assert.Equal(t, 3, len(allApps))
	assert.Equal(t, appNames[0], allApps[0].Name)
	assert.Equal(t, appNames[1], allApps[1].Name)
	assert.Equal(t, appNames[2], allApps[2].Name)
}
