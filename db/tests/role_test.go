//go:build integration

package db_tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedRoles = []struct {
	Name string
	ID   int32
}{
	{Name: "admin", ID: 1},
	{Name: "service", ID: 2},
	{Name: "user", ID: 3},
}

func TestGetRoleByID(t *testing.T) {
	tdb := SetupTestDB(t)

	tests := expectedRoles

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			r, err := tdb.Queries.GetRoleByID(tdb.Ctx, tt.ID)
			require.NoError(t, err, "Failed to fetch the role")
			assert.Equal(t, tt.ID, r.ID)
			assert.Equal(t, tt.Name, r.Name)
		})
	}

	t.Run("non-existent", func(t *testing.T) {
		_, err := tdb.Queries.GetRoleByID(tdb.Ctx, 999)
		assert.Error(t, err, "Should get error when fetching non-existent role")
	})
}

func TestGetRoleByName(t *testing.T) {
	tdb := SetupTestDB(t)
	tests := expectedRoles

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			r, err := tdb.Queries.GetRoleByName(tdb.Ctx, tt.Name)
			require.NoError(t, err, "Failed to fetch the role")
			assert.Equal(t, tt.ID, r.ID)
			assert.Equal(t, tt.Name, r.Name)
		})
	}

	t.Run("non-existent", func(t *testing.T) {
		_, err := tdb.Queries.GetRoleByName(tdb.Ctx, "non-existent")
		assert.Error(t, err, "Should get error when fetching non-existent role")
	})
}

func TestListRoles(t *testing.T) {
	tdb := SetupTestDB(t)

	actual, err := tdb.Queries.ListRoles(tdb.Ctx)
	require.NoError(t, err, "Failed to list roles")
	assert.Len(t, actual, len(expectedRoles))

	for i := range len(expectedRoles) {
		assert.Equal(t, expectedRoles[i].Name, actual[i].Name)
		assert.Equal(t, expectedRoles[i].ID, actual[i].ID)
	}
}
