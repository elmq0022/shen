//go:build integration

package db_tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedPermissions = []struct {
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

func TestGetPermissionByID(t *testing.T) {
	tdb := SetupTestDB(t)

	tests := expectedPermissions

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			p, err := tdb.Queries.GetPermissionByID(tdb.Ctx, tt.ID)
			require.NoError(t, err, "Failed to fetch the permission")
			assert.Equal(t, tt.ID, p.ID)
			assert.Equal(t, tt.Name, p.Name)
			assert.Equal(t, tt.Priority, p.Priority)
		})
	}

	t.Run("non-existent", func(t *testing.T) {
		_, err := tdb.Queries.GetPermissionByID(tdb.Ctx, 999)
		assert.Error(t, err, "Should get error when fetching non-existent permission")
	})
}

func TestGetPermissionByName(t *testing.T) {
	tdb := SetupTestDB(t)
	tests := expectedPermissions

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			p, err := tdb.Queries.GetPermissionByName(tdb.Ctx, tt.Name)
			require.NoError(t, err, "Failed to fetch the permission")
			assert.Equal(t, tt.ID, p.ID)
			assert.Equal(t, tt.Name, p.Name)
			assert.Equal(t, tt.Priority, p.Priority)
		})
	}

	t.Run("non-existent", func(t *testing.T) {
		_, err := tdb.Queries.GetPermissionByName(tdb.Ctx, "non-existent")
		assert.Error(t, err, "Should get error when fetching non-existent permission")
	})
}

func TestListPermissions(t *testing.T) {
	tdb := SetupTestDB(t)

	actual, err := tdb.Queries.ListPermissions(tdb.Ctx)
	require.NoError(t, err, "Failed to list permissions")
	assert.Len(t, actual, len(expectedPermissions))

	for i := range len(expectedPermissions) {
		assert.Equal(t, expectedPermissions[i].Name, actual[i].Name)
		assert.Equal(t, expectedPermissions[i].ID, actual[i].ID)
		assert.Equal(t, expectedPermissions[i].Priority, actual[i].Priority)
	}
}
