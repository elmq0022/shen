//go:build integration

package db_tests

import (
	"testing"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddGroupApplicationRole(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	// Add first role
	created1, err := tdb.Queries.AddGroupApplicationRole(tdb.Ctx, db.AddGroupApplicationRoleParams{
		GroupID:       f.Group1.ID,
		ApplicationID: f.App1.ID,
		RoleID:        ApplicationRoleViewer,
	})
	require.NoError(t, err, "Failed to add group application role")
	assert.Equal(t, ApplicationRoleViewer, created1.RoleID)

	// Add second role for same group-app (multi-role)
	created2, err := tdb.Queries.AddGroupApplicationRole(tdb.Ctx, db.AddGroupApplicationRoleParams{
		GroupID:       f.Group1.ID,
		ApplicationID: f.App1.ID,
		RoleID:        ApplicationRoleAuditor,
	})
	require.NoError(t, err, "Failed to add second role for same group-app")
	assert.Equal(t, ApplicationRoleAuditor, created2.RoleID)

	// Try to add duplicate - should fail
	_, err = tdb.Queries.AddGroupApplicationRole(tdb.Ctx, db.AddGroupApplicationRoleParams{
		GroupID:       f.Group1.ID,
		ApplicationID: f.App1.ID,
		RoleID:        ApplicationRoleViewer,
	})
	require.Error(t, err, "Should fail when adding duplicate role")
}

func TestGetGroupApplicationRole(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	created, err := tdb.Queries.AddGroupApplicationRole(tdb.Ctx, db.AddGroupApplicationRoleParams{
		GroupID:       f.Group1.ID,
		ApplicationID: f.App1.ID,
		RoleID:        ApplicationRoleOperator,
	})
	require.NoError(t, err, "Failed to create group application role")

	fetched, err := tdb.Queries.GetGroupApplicationRole(tdb.Ctx, db.GetGroupApplicationRoleParams{
		GroupID:       f.Group1.ID,
		ApplicationID: f.App1.ID,
		RoleID:        ApplicationRoleOperator,
	})
	require.NoError(t, err, "Failed to get group application role")
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, created.RoleID, fetched.RoleID)
}

func TestGetGroupApplicationRoleByID(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	created, err := tdb.Queries.AddGroupApplicationRole(tdb.Ctx, db.AddGroupApplicationRoleParams{
		GroupID:       f.Group1.ID,
		ApplicationID: f.App1.ID,
		RoleID:        ApplicationRoleViewer,
	})
	require.NoError(t, err, "Failed to create group application role")

	fetched, err := tdb.Queries.GetGroupApplicationRoleByID(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to get group application role by ID")
	assert.Equal(t, created, fetched)

	_, err = tdb.Queries.GetGroupApplicationRoleByID(tdb.Ctx, 999)
	require.Error(t, err, "Should get error for non-existent ID")
}

func TestDeleteGroupApplicationRole(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	// Add two roles
	created1, err := tdb.Queries.AddGroupApplicationRole(tdb.Ctx, db.AddGroupApplicationRoleParams{
		GroupID:       f.Group1.ID,
		ApplicationID: f.App1.ID,
		RoleID:        ApplicationRoleViewer,
	})
	require.NoError(t, err, "Failed to create first role")

	created2, err := tdb.Queries.AddGroupApplicationRole(tdb.Ctx, db.AddGroupApplicationRoleParams{
		GroupID:       f.Group1.ID,
		ApplicationID: f.App1.ID,
		RoleID:        ApplicationRoleOperator,
	})
	require.NoError(t, err, "Failed to create second role")

	// Delete one role
	err = tdb.Queries.DeleteGroupApplicationRole(tdb.Ctx, db.DeleteGroupApplicationRoleParams{
		GroupID:       created1.GroupID,
		ApplicationID: created1.ApplicationID,
		RoleID:        created1.RoleID,
	})
	require.NoError(t, err, "Failed to delete group application role")

	// Verify first role is gone
	_, err = tdb.Queries.GetGroupApplicationRoleByID(tdb.Ctx, created1.ID)
	require.Error(t, err, "Should get error when fetching deleted role")

	// Verify second role still exists
	_, err = tdb.Queries.GetGroupApplicationRoleByID(tdb.Ctx, created2.ID)
	require.NoError(t, err, "Second role should still exist")
}

func TestDeleteAllGroupApplicationRoles(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	// Add multiple roles for same group-app
	addGroupApplicationRoles(t, tdb, []db.AddGroupApplicationRoleParams{
		{GroupID: f.Group1.ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleViewer},
		{GroupID: f.Group1.ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleOperator},
		{GroupID: f.Group1.ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleAuditor},
	})

	// Count before deletion
	count, err := tdb.Queries.CountGroupApplicationRolesByGroup(tdb.Ctx, f.Group1.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// Delete all roles for this group-app
	err = tdb.Queries.DeleteAllGroupApplicationRoles(tdb.Ctx, db.DeleteAllGroupApplicationRolesParams{
		GroupID:       f.Group1.ID,
		ApplicationID: f.App1.ID,
	})
	require.NoError(t, err, "Failed to delete all group application roles")

	// Count after deletion
	count, err = tdb.Queries.CountGroupApplicationRolesByGroup(tdb.Ctx, f.Group1.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestGetUserApplicationRoles(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	// Group1 grants viewer and operator roles
	// Group2 grants auditor role
	addGroupApplicationRoles(t, tdb, []db.AddGroupApplicationRoleParams{
		{GroupID: f.Group1.ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleViewer},
		{GroupID: f.Group1.ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleOperator},
		{GroupID: f.Group2.ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleAuditor},
	})

	// Add user to both groups
	_, err := tdb.Queries.AddUserToGroup(tdb.Ctx, db.AddUserToGroupParams{
		UserID:  f.User1.ID,
		GroupID: f.Group1.ID,
	})
	require.NoError(t, err, "Failed to add User1 to Group1")

	_, err = tdb.Queries.AddUserToGroup(tdb.Ctx, db.AddUserToGroupParams{
		UserID:  f.User1.ID,
		GroupID: f.Group2.ID,
	})
	require.NoError(t, err, "Failed to add User1 to Group2")

	// Get all roles for user
	roles, err := tdb.Queries.GetUserApplicationRoles(tdb.Ctx, db.GetUserApplicationRolesParams{
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
	})
	require.NoError(t, err, "Failed to get user application roles")
	assert.Len(t, roles, 3, "User should have 3 roles from 2 groups")

	// Verify roles are sorted by priority (desc)
	assert.Equal(t, "operator", roles[0].Name, "Highest priority role should be first")
	assert.Equal(t, "auditor", roles[1].Name)
	assert.Equal(t, "viewer", roles[2].Name)
}

func TestGetUserGroups(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	// Add user to groups
	_, err := tdb.Queries.AddUserToGroup(tdb.Ctx, db.AddUserToGroupParams{
		UserID:  f.User1.ID,
		GroupID: f.Group1.ID,
	})
	require.NoError(t, err)

	_, err = tdb.Queries.AddUserToGroup(tdb.Ctx, db.AddUserToGroupParams{
		UserID:  f.User1.ID,
		GroupID: f.Group2.ID,
	})
	require.NoError(t, err)

	// Get user's groups
	groups, err := tdb.Queries.GetUserGroups(tdb.Ctx, f.User1.ID)
	require.NoError(t, err, "Failed to get user groups")
	assert.Len(t, groups, 2, "User should be in 2 groups")

	// Verify groups are sorted by name
	assert.Equal(t, f.Group1.Name, groups[0].Name)
	assert.Equal(t, f.Group2.Name, groups[1].Name)
}

func TestListGroupApplicationRolesByGroup(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	apps := CreateTestApplications(t, tdb, "zapp", 2)

	addGroupApplicationRoles(t, tdb, []db.AddGroupApplicationRoleParams{
		{GroupID: f.Group1.ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleViewer},
		{GroupID: f.Group1.ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleOperator},
		{GroupID: f.Group1.ID, ApplicationID: f.App2.ID, RoleID: ApplicationRoleAdmin},
		{GroupID: f.Group1.ID, ApplicationID: apps[0].ID, RoleID: ApplicationRoleOperator},
		{GroupID: f.Group1.ID, ApplicationID: apps[1].ID, RoleID: ApplicationRoleAuditor},
	})

	// First page - no cursor
	page1, err := tdb.Queries.ListGroupApplicationRolesByGroup(tdb.Ctx, db.ListGroupApplicationRolesByGroupParams{
		GroupID:               f.Group1.ID,
		CursorApplicationName: "",
		CursorRoleName:        "",
		Limit:                 2,
	})
	require.NoError(t, err, "Failed to list first page of roles")
	assert.Len(t, page1, 2)

	// Second page - use cursor from last item of page1
	page2, err := tdb.Queries.ListGroupApplicationRolesByGroup(tdb.Ctx, db.ListGroupApplicationRolesByGroupParams{
		GroupID:               f.Group1.ID,
		CursorApplicationName: page1[1].ApplicationName,
		CursorRoleName:        page1[1].RoleName,
		Limit:                 3,
	})
	require.NoError(t, err, "Failed to list second page of roles")
	assert.Len(t, page2, 3)

	// Get all roles - no cursor, high limit
	all, err := tdb.Queries.ListGroupApplicationRolesByGroup(tdb.Ctx, db.ListGroupApplicationRolesByGroupParams{
		GroupID:               f.Group1.ID,
		CursorApplicationName: "",
		CursorRoleName:        "",
		Limit:                 10,
	})
	require.NoError(t, err, "Failed to list all roles for group")
	assert.Len(t, all, 5)
}

func TestListGroupApplicationRolesByApplication(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	groups := CreateTestGroups(t, tdb, "team", 2)

	addGroupApplicationRoles(t, tdb, []db.AddGroupApplicationRoleParams{
		{GroupID: f.Group1.ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleViewer},
		{GroupID: f.Group1.ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleOperator},
		{GroupID: f.Group2.ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleAdmin},
		{GroupID: groups[0].ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleOperator},
		{GroupID: groups[1].ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleAuditor},
	})

	// First page - no cursor
	page1, err := tdb.Queries.ListGroupApplicationRolesByApplication(tdb.Ctx, db.ListGroupApplicationRolesByApplicationParams{
		ApplicationID:   f.App1.ID,
		CursorGroupName: "",
		CursorRoleName:  "",
		Limit:           2,
	})
	require.NoError(t, err, "Failed to list first page of roles")
	assert.Len(t, page1, 2)

	// Second page - use cursor from last item of page1
	page2, err := tdb.Queries.ListGroupApplicationRolesByApplication(tdb.Ctx, db.ListGroupApplicationRolesByApplicationParams{
		ApplicationID:   f.App1.ID,
		CursorGroupName: page1[1].GroupName,
		CursorRoleName:  page1[1].RoleName,
		Limit:           2,
	})
	require.NoError(t, err, "Failed to list second page of roles")
	assert.Len(t, page2, 2)

	// Get all roles - no cursor, high limit
	all, err := tdb.Queries.ListGroupApplicationRolesByApplication(tdb.Ctx, db.ListGroupApplicationRolesByApplicationParams{
		ApplicationID:   f.App1.ID,
		CursorGroupName: "",
		CursorRoleName:  "",
		Limit:           10,
	})
	require.NoError(t, err, "Failed to list all roles for application")
	assert.Len(t, all, 5)
}

func TestCountGroupApplicationRolesByGroup(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	apps := CreateTestApplications(t, tdb, "app", 3)

	addGroupApplicationRoles(t, tdb, []db.AddGroupApplicationRoleParams{
		{GroupID: f.Group1.ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleViewer},
		{GroupID: f.Group1.ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleOperator},
		{GroupID: f.Group1.ID, ApplicationID: apps[0].ID, RoleID: ApplicationRoleAdmin},
		{GroupID: f.Group1.ID, ApplicationID: apps[1].ID, RoleID: ApplicationRoleOperator},
	})

	count, err := tdb.Queries.CountGroupApplicationRolesByGroup(tdb.Ctx, f.Group1.ID)
	require.NoError(t, err, "Failed to count roles for Group1")
	assert.Equal(t, int64(4), count)
}

func TestCountGroupApplicationRolesByApplication(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	groups := CreateTestGroups(t, tdb, "team", 3)

	addGroupApplicationRoles(t, tdb, []db.AddGroupApplicationRoleParams{
		{GroupID: f.Group1.ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleViewer},
		{GroupID: f.Group1.ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleOperator},
		{GroupID: groups[0].ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleAdmin},
		{GroupID: groups[1].ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleOperator},
	})

	count, err := tdb.Queries.CountGroupApplicationRolesByApplication(tdb.Ctx, f.App1.ID)
	require.NoError(t, err, "Failed to count roles for App1")
	assert.Equal(t, int64(4), count)
}

func TestCountAllGroupApplicationRoles(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	addGroupApplicationRoles(t, tdb, []db.AddGroupApplicationRoleParams{
		{GroupID: f.Group1.ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleViewer},
		{GroupID: f.Group1.ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleOperator},
		{GroupID: f.Group1.ID, ApplicationID: f.App2.ID, RoleID: ApplicationRoleAdmin},
		{GroupID: f.Group2.ID, ApplicationID: f.App1.ID, RoleID: ApplicationRoleOperator},
	})

	count, err := tdb.Queries.CountAllGroupApplicationRoles(tdb.Ctx)
	require.NoError(t, err, "Failed to count all roles")
	assert.Equal(t, int64(4), count)
}
