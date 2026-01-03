//go:build integration

package db_tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedApplicationRoles = []struct {
	Name     string
	ID       int32
	Priority int32
}{
	{Name: "authenticated", ID: 1, Priority: 100},
	{Name: "viewer", ID: 2, Priority: 200},
	{Name: "auditor", ID: 3, Priority: 300},
	{Name: "operator", ID: 4, Priority: 400},
	{Name: "admin", ID: 5, Priority: 500},
}

func TestGetApplicationRoleByID(t *testing.T) {
	tdb := SetupTestDB(t)

	tests := expectedApplicationRoles

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			r, err := tdb.Queries.GetApplicationRoleByID(tdb.Ctx, tt.ID)
			require.NoError(t, err, "Failed to fetch the application role")
			assert.Equal(t, tt.ID, r.ID)
			assert.Equal(t, tt.Name, r.Name)
			assert.Equal(t, tt.Priority, r.Priority)
		})
	}

	t.Run("non-existent", func(t *testing.T) {
		_, err := tdb.Queries.GetApplicationRoleByID(tdb.Ctx, 999)
		assert.Error(t, err, "Should get error when fetching non-existent application role")
	})
}

func TestGetApplicationRoleByName(t *testing.T) {
	tdb := SetupTestDB(t)
	tests := expectedApplicationRoles

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			r, err := tdb.Queries.GetApplicationRoleByName(tdb.Ctx, tt.Name)
			require.NoError(t, err, "Failed to fetch the application role")
			assert.Equal(t, tt.ID, r.ID)
			assert.Equal(t, tt.Name, r.Name)
			assert.Equal(t, tt.Priority, r.Priority)
		})
	}

	t.Run("non-existent", func(t *testing.T) {
		_, err := tdb.Queries.GetApplicationRoleByName(tdb.Ctx, "non-existent")
		assert.Error(t, err, "Should get error when fetching non-existent application role")
	})
}

func TestListApplicationRoles(t *testing.T) {
	tdb := SetupTestDB(t)

	actual, err := tdb.Queries.ListApplicationRoles(tdb.Ctx)
	require.NoError(t, err, "Failed to list application roles")
	assert.Len(t, actual, len(expectedApplicationRoles))

	for i := range len(expectedApplicationRoles) {
		assert.Equal(t, expectedApplicationRoles[i].Name, actual[i].Name)
		assert.Equal(t, expectedApplicationRoles[i].ID, actual[i].ID)
		assert.Equal(t, expectedApplicationRoles[i].Priority, actual[i].Priority)
	}
}
