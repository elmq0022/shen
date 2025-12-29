//go:build integration

package db_tests

import (
	"testing"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAndGetUser(t *testing.T) {
	tdb := SetupTestDB(t)

	// Create user
	created, err := tdb.Queries.CreateUser(tdb.Ctx, db.CreateUserParams{
		Username:       "tyler.durden",
		HashedPassword: pgtype.Text{String: "hash123", Valid: true},
		Role:           2,
	})
	require.NoError(t, err, "Failed to create user")

	// Verify created user fields
	assert.Equal(t, "tyler.durden", created.Username)
	assert.True(t, created.HashedPassword.Valid)
	assert.Equal(t, "hash123", created.HashedPassword.String)
	assert.Equal(t, int32(2), created.Role)
	assert.True(t, created.Active, "User should be active by default")
	assert.NotZero(t, created.ID)
	assert.NotZero(t, created.CreatedAt)

	// Get user by ID
	fetched, err := tdb.Queries.GetUserByID(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to get user by ID")
	assert.Equal(t, created, fetched)

	// Get user by username
	fetchedByUsername, err := tdb.Queries.GetUserByUsername(tdb.Ctx, "tyler.durden")
	require.NoError(t, err, "Failed to get user by username")
	assert.Equal(t, created, fetchedByUsername)

	// Deactivate User
	err = tdb.Queries.DeactivateUser(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to deactivate user")

	// Verify user is deactivated
	deactivated, err := tdb.Queries.GetUserByID(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to get deactivated user")
	assert.False(t, deactivated.Active, "User should be deactivated")
}
