//go:build integration

package db_tests

import (
	"slices"
	"testing"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateApplication(t *testing.T) {
	tdb := SetupTestDB(t)

	created := CreateTestApplication(t, tdb, "my-app")

	assert.Equal(t, "my-app", created.Name)
	assert.True(t, created.Active, "Application should be active by default")
	assert.NotZero(t, created.ID)
	assert.NotZero(t, created.CreatedAt)
}

func TestGetApplication(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	// Get by ID
	fetchedByID, err := tdb.Queries.GetApplicationByID(tdb.Ctx, f.App1.ID)
	require.NoError(t, err, "Failed to fetch by application ID")
	assert.Equal(t, f.App1, fetchedByID)

	// Get by Name
	fetchedByName, err := tdb.Queries.GetApplicationByName(tdb.Ctx, f.App1.Name)
	require.NoError(t, err, "Failed to get application by name")
	assert.Equal(t, f.App1, fetchedByName)
}

func TestDeactivateApplication(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	err := tdb.Queries.DeactivateApplication(tdb.Ctx, f.App1.ID)
	require.NoError(t, err, "Failed to deactivate application")

	deactivated, err := tdb.Queries.GetApplicationByID(tdb.Ctx, f.App1.ID)
	require.NoError(t, err, "Failed to get deactivated application")
	assert.False(t, deactivated.Active, "Application should be deactivated")
}

func TestUpdateApplication(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	// Update application name and active status
	err := tdb.Queries.UpdateApplication(tdb.Ctx, db.UpdateApplicationParams{
		ID:     f.App1.ID,
		Name:   "updated-app",
		Active: false,
	})
	require.NoError(t, err, "Failed to update application")

	// Verify update
	updated, err := tdb.Queries.GetApplicationByID(tdb.Ctx, f.App1.ID)
	require.NoError(t, err, "Failed to get updated application")
	assert.Equal(t, "updated-app", updated.Name)
	assert.False(t, updated.Active)
	assert.Equal(t, f.App1.ID, updated.ID)
}

func TestDeleteApplication(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	// Delete application
	err := tdb.Queries.DeleteApplication(tdb.Ctx, f.App1.ID)
	require.NoError(t, err, "Failed to delete application")

	// Verify deletion - should get error when trying to fetch
	_, err = tdb.Queries.GetApplicationByID(tdb.Ctx, f.App1.ID)
	assert.Error(t, err, "Should get error when fetching deleted application")
}

func TestListActiveApplications(t *testing.T) {
	tdb := SetupTestDB(t)

	// Create test applications
	apps := CreateTestApplications(t, tdb, "app", 3)

	appNames := []string{"app-1", "app-2", "app-3"}
	slices.Sort(appNames)

	// first page of apps (cursor-based)
	page1, err := tdb.Queries.ListActiveApplications(tdb.Ctx, db.ListActiveApplicationsParams{
		Column1: "",
		Limit:   2,
	})
	require.NoError(t, err, "Failed to list active applications")
	assert.Len(t, page1, 2)
	assert.Equal(t, appNames[0], page1[0].Name)
	assert.Equal(t, appNames[1], page1[1].Name)

	// second page of apps (cursor-based)
	page2, err := tdb.Queries.ListActiveApplications(tdb.Ctx, db.ListActiveApplicationsParams{
		Column1: page1[len(page1)-1].Name,
		Limit:   2,
	})
	require.NoError(t, err, "Failed to list active applications")
	assert.Len(t, page2, 1)
	assert.Equal(t, appNames[2], page2[0].Name)

	// deactivate middle app and verify filtering
	err = tdb.Queries.DeactivateApplication(tdb.Ctx, apps[1].ID)
	require.NoError(t, err, "Failed to deactivate application")

	activeApps, err := tdb.Queries.ListActiveApplications(tdb.Ctx, db.ListActiveApplicationsParams{
		Column1: "",
		Limit:   10,
	})
	require.NoError(t, err, "Failed to list active apps after deactivation")
	assert.Len(t, activeApps, 2)
	assert.Equal(t, appNames[0], activeApps[0].Name)
	assert.Equal(t, appNames[2], activeApps[1].Name)
}

func TestListApplications(t *testing.T) {
	tdb := SetupTestDB(t)

	// Create test applications
	apps := CreateTestApplications(t, tdb, "app", 3)

	appNames := []string{"app-1", "app-2", "app-3"}
	slices.Sort(appNames)

	// first page of apps (cursor-based)
	page1, err := tdb.Queries.ListApplications(tdb.Ctx, db.ListApplicationsParams{
		Column1: "",
		Limit:   2,
	})
	require.NoError(t, err, "Failed to list applications")
	assert.Len(t, page1, 2)
	assert.Equal(t, appNames[0], page1[0].Name)
	assert.Equal(t, appNames[1], page1[1].Name)

	// second page of apps (cursor-based)
	page2, err := tdb.Queries.ListApplications(tdb.Ctx, db.ListApplicationsParams{
		Column1: page1[len(page1)-1].Name,
		Limit:   2,
	})
	require.NoError(t, err, "Failed to list applications")
	assert.Len(t, page2, 1)
	assert.Equal(t, appNames[2], page2[0].Name)

	// deactivate one app
	err = tdb.Queries.DeactivateApplication(tdb.Ctx, apps[1].ID)
	require.NoError(t, err, "Failed to deactivate application")

	// ListApplications should return all apps including deactivated (cursor-based)
	allApps, err := tdb.Queries.ListApplications(tdb.Ctx, db.ListApplicationsParams{
		Column1: "",
		Limit:   10,
	})
	require.NoError(t, err, "Failed to list all applications")
	assert.Len(t, allApps, 3)
	assert.Equal(t, appNames[0], allApps[0].Name)
	assert.Equal(t, appNames[1], allApps[1].Name)
	assert.Equal(t, appNames[2], allApps[2].Name)
}
