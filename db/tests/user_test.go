//go:build integration

package db_tests

import (
	"slices"
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

func TestUpdateUserPassword(t *testing.T) {
	tdb := SetupTestDB(t)

	// Create user
	created, err := tdb.Queries.CreateUser(tdb.Ctx, db.CreateUserParams{
		Username:       "marla.singer",
		HashedPassword: pgtype.Text{String: "oldhash", Valid: true},
		Role:           2,
	})
	require.NoError(t, err, "Failed to create user")

	// Update password
	err = tdb.Queries.UpdateUserPassword(tdb.Ctx, db.UpdateUserPasswordParams{
		ID:             created.ID,
		HashedPassword: pgtype.Text{String: "newhash", Valid: true},
	})
	require.NoError(t, err, "Failed to update password")

	// Verify password updated
	updated, err := tdb.Queries.GetUserByID(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to get updated user")
	assert.Equal(t, "newhash", updated.HashedPassword.String)
	assert.Equal(t, created.Username, updated.Username)
}

func TestUpdateUserRole(t *testing.T) {
	tdb := SetupTestDB(t)

	// Create user
	created, err := tdb.Queries.CreateUser(tdb.Ctx, db.CreateUserParams{
		Username:       "robert.paulson",
		HashedPassword: pgtype.Text{String: "hash123", Valid: true},
		Role:           2,
	})
	require.NoError(t, err, "Failed to create user")

	// Update role
	err = tdb.Queries.UpdateUserRole(tdb.Ctx, db.UpdateUserRoleParams{
		ID:   created.ID,
		Role: 1,
	})
	require.NoError(t, err, "Failed to update role")

	// Verify role updated
	updated, err := tdb.Queries.GetUserByID(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to get updated user")
	assert.Equal(t, int32(1), updated.Role)
	assert.Equal(t, created.Username, updated.Username)
}

func TestActivateUser(t *testing.T) {
	tdb := SetupTestDB(t)

	// Create and deactivate user
	created, err := tdb.Queries.CreateUser(tdb.Ctx, db.CreateUserParams{
		Username:       "angel.face",
		HashedPassword: pgtype.Text{String: "hash123", Valid: true},
		Role:           2,
	})
	require.NoError(t, err, "Failed to create user")

	err = tdb.Queries.DeactivateUser(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to deactivate user")

	// Activate user
	err = tdb.Queries.ActivateUser(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to activate user")

	// Verify user is active
	activated, err := tdb.Queries.GetUserByID(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to get activated user")
	assert.True(t, activated.Active, "User should be active")
}

func TestDeleteUser(t *testing.T) {
	tdb := SetupTestDB(t)

	// Create user
	created, err := tdb.Queries.CreateUser(tdb.Ctx, db.CreateUserParams{
		Username:       "big.bob",
		HashedPassword: pgtype.Text{String: "hash123", Valid: true},
		Role:           2,
	})
	require.NoError(t, err, "Failed to create user")

	// Delete user
	err = tdb.Queries.DeleteUser(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to delete user")

	// Verify deletion - should get error when trying to fetch
	_, err = tdb.Queries.GetUserByID(tdb.Ctx, created.ID)
	assert.Error(t, err, "Should get error when fetching deleted user")
}

func TestListActiveUsers(t *testing.T) {
	tdb := SetupTestDB(t)

	usernames := []string{
		"tyler.durden",
		"robert.paulson",
		"marla.singer",
	}

	for _, username := range usernames {
		tdb.Queries.CreateUser(tdb.Ctx, db.CreateUserParams{
			Username:       username,
			HashedPassword: pgtype.Text{String: username + "-hash123", Valid: true},
			Role:           1,
		})
	}

	numberActiveUsers, err := tdb.Queries.CountActiveUsers(tdb.Ctx)
	require.NoError(t, err, "Failed to count users")
	assert.Equal(t, int64(len(usernames)), int64(numberActiveUsers))

	slices.Sort(usernames)

	// first page of users
	users, err := tdb.Queries.ListActiveUsers(tdb.Ctx, db.ListActiveUsersParams{
		Limit:  2,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to list users")
	assert.Equal(t, len(users), 2)
	assert.Equal(t, usernames[0], users[0].Username)
	assert.Equal(t, usernames[1], users[1].Username)

	// second page of users
	users, err = tdb.Queries.ListActiveUsers(tdb.Ctx, db.ListActiveUsersParams{
		Limit:  2,
		Offset: 2,
	})
	require.NoError(t, err, "Failed to list users")
	assert.Equal(t, len(users), 1)
	assert.Equal(t, usernames[2], users[0].Username)

	// deactivate middle user and verify filtering
	userToDeactivate, err := tdb.Queries.GetUserByUsername(tdb.Ctx, usernames[1])
	require.NoError(t, err, "Failed to get user to deactivate")
	err = tdb.Queries.DeactivateUser(tdb.Ctx, userToDeactivate.ID)
	require.NoError(t, err, "Failed to deactivate user")

	numberActiveUsers, err = tdb.Queries.CountActiveUsers(tdb.Ctx)
	require.NoError(t, err, "Failed to count active users after deactivation")
	assert.Equal(t, int64(2), numberActiveUsers)

	users, err = tdb.Queries.ListActiveUsers(tdb.Ctx, db.ListActiveUsersParams{
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to list active users after deactivation")
	assert.Equal(t, 2, len(users))
	assert.Equal(t, usernames[0], users[0].Username)
	assert.Equal(t, usernames[2], users[1].Username)
}

func TestListUsers(t *testing.T) {
	tdb := SetupTestDB(t)

	usernames := []string{
		"narrator",
		"chloe",
		"mechanic",
	}

	// create users
	for _, username := range usernames {
		_, err := tdb.Queries.CreateUser(tdb.Ctx, db.CreateUserParams{
			Username:       username,
			HashedPassword: pgtype.Text{String: username + "-hash123", Valid: true},
			Role:           2,
		})
		require.NoError(t, err, "Failed to create user")
	}

	// deactivate one user
	userToDeactivate, err := tdb.Queries.GetUserByUsername(tdb.Ctx, "chloe")
	require.NoError(t, err, "Failed to get user to deactivate")
	err = tdb.Queries.DeactivateUser(tdb.Ctx, userToDeactivate.ID)
	require.NoError(t, err, "Failed to deactivate user")

	slices.Sort(usernames)

	// ListUsers should return all users including deactivated
	allUsers, err := tdb.Queries.ListUsers(tdb.Ctx, db.ListUsersParams{
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to list all users")
	assert.Equal(t, 3, len(allUsers))
	assert.Equal(t, usernames[0], allUsers[0].Username)
	assert.Equal(t, usernames[1], allUsers[1].Username)
	assert.Equal(t, usernames[2], allUsers[2].Username)

	// verify count includes all users
	totalUsers, err := tdb.Queries.CountUsers(tdb.Ctx)
	require.NoError(t, err, "Failed to count all users")
	assert.Equal(t, int64(3), totalUsers)
}
