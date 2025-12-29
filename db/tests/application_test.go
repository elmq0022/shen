//go:build integration

package db_tests

import (
	"testing"

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
