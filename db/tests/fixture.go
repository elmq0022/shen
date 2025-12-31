//go:build integration

package db_tests

import (
	"fmt"
	"testing"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

const (
	RoleService int32 = 1
	RoleUser    int32 = 2
	RoleAdmin   int32 = 3
)

func CreateTestUser(t *testing.T, tdb *TestDB, username string, role int32) db.ShenUser {
	t.Helper()

	user, err := tdb.Queries.CreateUser(tdb.Ctx, db.CreateUserParams{
		Username:       username,
		HashedPassword: pgtype.Text{String: username + "-hash123", Valid: true},
		Role:           role,
	})
	require.NoError(t, err, "Failed to create test user: %s", username)

	return user
}

// CreateTestGroup creates a group with the given name.
func CreateTestGroup(t *testing.T, tdb *TestDB, name string) db.ShenGroup {
	t.Helper()

	group, err := tdb.Queries.CreateGroup(tdb.Ctx, name)
	require.NoError(t, err, "Failed to create test group: %s", name)

	return group
}

// CreateTestApplication creates an application with the given name.
func CreateTestApplication(t *testing.T, tdb *TestDB, name string) db.ShenApplication {
	t.Helper()

	app, err := tdb.Queries.CreateApplication(tdb.Ctx, name)
	require.NoError(t, err, "Failed to create test application: %s", name)

	return app
}

// CreateTestUsers creates multiple users with sequential naming.
// Example: CreateTestUsers(t, tdb, "user", 3) creates user-1, user-2, user-3
func CreateTestUsers(t *testing.T, tdb *TestDB, prefix string, count int) []db.ShenUser {
	t.Helper()

	users := make([]db.ShenUser, count)
	for i := 0; i < count; i++ {
		username := fmt.Sprintf("%s-%d", prefix, i+1)
		users[i] = CreateTestUser(t, tdb, username, RoleUser)
	}

	return users
}

// CreateTestGroups creates multiple groups with sequential naming.
// Example: CreateTestGroups(t, tdb, "group", 3) creates group-1, group-2, group-3
func CreateTestGroups(t *testing.T, tdb *TestDB, prefix string, count int) []db.ShenGroup {
	t.Helper()

	groups := make([]db.ShenGroup, count)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s-%d", prefix, i+1)
		groups[i] = CreateTestGroup(t, tdb, name)
	}

	return groups
}

// CreateTestApplications creates multiple applications with sequential naming.
// Example: CreateTestApplications(t, tdb, "app", 3) creates app-1, app-2, app-3
func CreateTestApplications(t *testing.T, tdb *TestDB, prefix string, count int) []db.ShenApplication {
	t.Helper()

	apps := make([]db.ShenApplication, count)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s-%d", prefix, i+1)
		apps[i] = CreateTestApplication(t, tdb, name)
	}

	return apps
}

// StandardFixtures contains commonly used test data
type StandardFixtures struct {
	User1  db.ShenUser
	User2  db.ShenUser
	Admin  db.ShenUser
	Group1 db.ShenGroup
	Group2 db.ShenGroup
	App1   db.ShenApplication
	App2   db.ShenApplication
}

// CreateStandardFixtures creates a standard set of test data with predictable names.
// Returns fixtures with 2 regular users, 1 admin user, 2 groups, and 2 applications.
func CreateStandardFixtures(t *testing.T, tdb *TestDB) *StandardFixtures {
	t.Helper()

	return &StandardFixtures{
		User1:  CreateTestUser(t, tdb, "test.user1", RoleUser),
		User2:  CreateTestUser(t, tdb, "test.user2", RoleUser),
		Admin:  CreateTestUser(t, tdb, "test.admin", RoleAdmin),
		Group1: CreateTestGroup(t, tdb, "test-group-1"),
		Group2: CreateTestGroup(t, tdb, "test-group-2"),
		App1:   CreateTestApplication(t, tdb, "test-app-1"),
		App2:   CreateTestApplication(t, tdb, "test-app-2"),
	}
}
