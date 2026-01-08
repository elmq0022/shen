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

func TestCreateUser(t *testing.T) {
	tdb := SetupTestDB(t)

	created := CreateTestUser(t, tdb, "alice", RoleUser)

	// Verify created user fields
	assert.Equal(t, "alice", created.Username)
	assert.True(t, created.HashedPassword.Valid)
	assert.Equal(t, "alice-hash123", created.HashedPassword.String)
	assert.Equal(t, RoleUser, created.Role)
	assert.True(t, created.Active, "User should be active by default")
	assert.NotZero(t, created.ID)
	assert.NotZero(t, created.CreatedAt)
}

func TestGetUser(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	// Get user by ID
	fetchedByID, err := tdb.Queries.GetUserByID(tdb.Ctx, f.User1.ID)
	require.NoError(t, err, "Failed to get user by ID")
	assert.Equal(t, f.User1, fetchedByID)

	// Get user by username
	fetchedByUsername, err := tdb.Queries.GetUserByUsername(tdb.Ctx, "test.user1")
	require.NoError(t, err, "Failed to get user by username")
	assert.Equal(t, f.User1, fetchedByUsername)
}

func TestDeactivateUser(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	err := tdb.Queries.DeactivateUser(tdb.Ctx, f.User1.ID)
	require.NoError(t, err, "Failed to deactivate user")

	deactivated, err := tdb.Queries.GetUserByID(tdb.Ctx, f.User1.ID)
	require.NoError(t, err, "Failed to get deactivated user")
	assert.False(t, deactivated.Active, "User should be deactivated")
}

func TestUpdateUserPassword(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	oldPassword := f.User1.HashedPassword.String

	// Update password
	newPassword := "newhash"
	err := tdb.Queries.UpdateUserPassword(tdb.Ctx, db.UpdateUserPasswordParams{
		ID:             f.User1.ID,
		HashedPassword: pgtype.Text{String: newPassword, Valid: true},
	})
	require.NoError(t, err, "Failed to update password")

	// Verify password updated
	updated, err := tdb.Queries.GetUserByID(tdb.Ctx, f.User1.ID)
	require.NoError(t, err, "Failed to get updated user")
	assert.Equal(t, newPassword, updated.HashedPassword.String)
	assert.NotEqual(t, oldPassword, updated.HashedPassword.String)
	assert.Equal(t, f.User1.Username, updated.Username)
}

func TestUpdateUserRole(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	assert.Equal(t, RoleUser, f.User1.Role)

	// Update role to admin
	err := tdb.Queries.UpdateUserRole(tdb.Ctx, db.UpdateUserRoleParams{
		ID:   f.User1.ID,
		Role: RoleAdmin,
	})
	require.NoError(t, err, "Failed to update role")

	// Verify role updated
	updated, err := tdb.Queries.GetUserByID(tdb.Ctx, f.User1.ID)
	require.NoError(t, err, "Failed to get updated user")
	assert.Equal(t, RoleAdmin, updated.Role)
	assert.Equal(t, f.User1.Username, updated.Username)
}

func TestActivateUser(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	// Deactivate user first
	err := tdb.Queries.DeactivateUser(tdb.Ctx, f.User1.ID)
	require.NoError(t, err, "Failed to deactivate user")

	// Activate user
	err = tdb.Queries.ActivateUser(tdb.Ctx, f.User1.ID)
	require.NoError(t, err, "Failed to activate user")

	// Verify user is active
	activated, err := tdb.Queries.GetUserByID(tdb.Ctx, f.User1.ID)
	require.NoError(t, err, "Failed to get activated user")
	assert.True(t, activated.Active, "User should be active")
}

func TestDeleteUser(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	// Delete user
	err := tdb.Queries.DeleteUser(tdb.Ctx, f.User1.ID)
	require.NoError(t, err, "Failed to delete user")

	// Verify deletion - should get error when trying to fetch
	_, err = tdb.Queries.GetUserByID(tdb.Ctx, f.User1.ID)
	assert.Error(t, err, "Should get error when fetching deleted user")
}

func TestListActiveUsers(t *testing.T) {
	tdb := SetupTestDB(t)

	// Create test users
	users := CreateTestUsers(t, tdb, "user", 3)

	usernames := []string{"user-1", "user-2", "user-3"}
	slices.Sort(usernames)

	numberActiveUsers, err := tdb.Queries.CountActiveUsers(tdb.Ctx)
	require.NoError(t, err, "Failed to count users")
	assert.Equal(t, int64(len(users)), numberActiveUsers)

	// first page of users (cursor-based)
	page1, err := tdb.Queries.ListActiveUsers(tdb.Ctx, db.ListActiveUsersParams{
		Column1: "",
		Limit:   2,
	})
	require.NoError(t, err, "Failed to list users")
	assert.Len(t, page1, 2)
	assert.Equal(t, usernames[0], page1[0].Username)
	assert.Equal(t, usernames[1], page1[1].Username)

	// second page of users (cursor-based)
	page2, err := tdb.Queries.ListActiveUsers(tdb.Ctx, db.ListActiveUsersParams{
		Column1: page1[len(page1)-1].Username,
		Limit:   2,
	})
	require.NoError(t, err, "Failed to list users")
	assert.Len(t, page2, 1)
	assert.Equal(t, usernames[2], page2[0].Username)

	// deactivate middle user and verify filtering
	err = tdb.Queries.DeactivateUser(tdb.Ctx, users[1].ID)
	require.NoError(t, err, "Failed to deactivate user")

	numberActiveUsers, err = tdb.Queries.CountActiveUsers(tdb.Ctx)
	require.NoError(t, err, "Failed to count active users after deactivation")
	assert.Equal(t, int64(2), numberActiveUsers)

	activeUsers, err := tdb.Queries.ListActiveUsers(tdb.Ctx, db.ListActiveUsersParams{
		Column1: "",
		Limit:   10,
	})
	require.NoError(t, err, "Failed to list active users after deactivation")
	assert.Len(t, activeUsers, 2)
	assert.Equal(t, usernames[0], activeUsers[0].Username)
	assert.Equal(t, usernames[2], activeUsers[1].Username)
}

func TestListUsers(t *testing.T) {
	tdb := SetupTestDB(t)

	// Create test users
	users := CreateTestUsers(t, tdb, "user", 3)

	// deactivate one user
	err := tdb.Queries.DeactivateUser(tdb.Ctx, users[1].ID)
	require.NoError(t, err, "Failed to deactivate user")

	usernames := []string{"user-1", "user-2", "user-3"}
	slices.Sort(usernames)

	// ListUsers should return all users including deactivated (cursor-based)
	allUsers, err := tdb.Queries.ListUsers(tdb.Ctx, db.ListUsersParams{
		Column1: "",
		Limit:   10,
	})
	require.NoError(t, err, "Failed to list all users")
	assert.Len(t, allUsers, 3)
	assert.Equal(t, usernames[0], allUsers[0].Username)
	assert.Equal(t, usernames[1], allUsers[1].Username)
	assert.Equal(t, usernames[2], allUsers[2].Username)

	// verify count includes all users
	totalUsers, err := tdb.Queries.CountUsers(tdb.Ctx)
	require.NoError(t, err, "Failed to count all users")
	assert.Equal(t, int64(3), totalUsers)
}

func TestCheckUsernameExists(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	// check existing username
	exists, err := tdb.Queries.CheckUsernameExists(tdb.Ctx, f.User1.Username)
	require.NoError(t, err, "Failed to check username exists")
	assert.True(t, exists)

	// check non-existent username
	exists, err = tdb.Queries.CheckUsernameExists(tdb.Ctx, "nonexistent.user")
	require.NoError(t, err, "Failed to check non-existent username")
	assert.False(t, exists)
}

func TestListUsersByRole(t *testing.T) {
	tdb := SetupTestDB(t)

	// create users with different roles
	admin1 := CreateTestUser(t, tdb, "admin1", RoleAdmin)
	admin2 := CreateTestUser(t, tdb, "admin2", RoleAdmin)
	user1 := CreateTestUser(t, tdb, "user1", RoleUser)
	user2 := CreateTestUser(t, tdb, "user2", RoleUser)

	// list admin users (cursor-based)
	admins, err := tdb.Queries.ListUsersByRole(tdb.Ctx, db.ListUsersByRoleParams{
		Role:    RoleAdmin,
		Column2: "",
		Limit:   10,
	})
	require.NoError(t, err, "Failed to list admin users")
	assert.Len(t, admins, 2)
	assert.Equal(t, admin1.Username, admins[0].Username)
	assert.Equal(t, admin2.Username, admins[1].Username)

	// list member users (cursor-based)
	members, err := tdb.Queries.ListUsersByRole(tdb.Ctx, db.ListUsersByRoleParams{
		Role:    RoleUser,
		Column2: "",
		Limit:   10,
	})
	require.NoError(t, err, "Failed to list member users")
	assert.Len(t, members, 2)
	assert.Equal(t, user1.Username, members[0].Username)
	assert.Equal(t, user2.Username, members[1].Username)
}
