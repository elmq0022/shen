//go:build integration

package db_tests

import (
	"testing"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateGroupManager(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	created, err := tdb.Queries.AddManagerToGroup(tdb.Ctx, db.AddManagerToGroupParams{
		UserID:  f.User1.ID,
		GroupID: f.Group1.ID,
	})
	require.NoError(t, err, "Failed to add manager to group")

	fetched, err := tdb.Queries.GetUserGroupManagerByID(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to retrieve group manager")
	assert.Equal(t, created, fetched)

	err = tdb.Queries.RemoveManagerFromGroup(tdb.Ctx, db.RemoveManagerFromGroupParams{
		UserID:  f.User1.ID,
		GroupID: f.Group1.ID,
	})
	require.NoError(t, err, "Failed to remove manager from group")

	_, err = tdb.Queries.GetUserGroupManagerByID(tdb.Ctx, created.ID)
	require.Error(t, err, "Group manager should not exist after removal")
}

func TestListManagersByGroup(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	group1Managers := []db.ShenUser{f.User1, f.User2, f.Admin}
	addManagersToGroup(t, tdb, group1Managers, f.Group1)
	sortUsersByUsername(group1Managers)

	group2Managers := []db.ShenUser{f.User1, f.User2}
	addManagersToGroup(t, tdb, group2Managers, f.Group2)
	sortUsersByUsername(group2Managers)

	fetchedGroup1Managers, err := tdb.Queries.ListManagersByGroup(tdb.Ctx, db.ListManagersByGroupParams{
		GroupID: f.Group1.ID,
		Limit:   10,
		Offset:  0,
	})
	require.NoError(t, err, "Failed to retrieve Group1 managers")
	assert.Equal(t, group1Managers, fetchedGroup1Managers)

	fetchedGroup2Managers, err := tdb.Queries.ListManagersByGroup(tdb.Ctx, db.ListManagersByGroupParams{
		GroupID: f.Group2.ID,
		Limit:   10,
		Offset:  0,
	})
	require.NoError(t, err, "Failed to retrieve Group2 managers")
	assert.Equal(t, group2Managers, fetchedGroup2Managers)
}

func TestListGroupsManagedByUser(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	user1Groups := []db.ShenGroup{f.Group1, f.Group2}
	addManagersToGroup(t, tdb, []db.ShenUser{f.User1}, f.Group1)
	addManagersToGroup(t, tdb, []db.ShenUser{f.User1}, f.Group2)
	sortGroupsByName(user1Groups)

	user2Groups := []db.ShenGroup{f.Group1}
	addManagersToGroup(t, tdb, []db.ShenUser{f.User2}, f.Group1)
	sortGroupsByName(user2Groups)

	fetchedUser1Groups, err := tdb.Queries.ListGroupsManagedByUser(tdb.Ctx, db.ListGroupsManagedByUserParams{
		UserID: f.User1.ID,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to retrieve User1 managed groups")
	assert.Equal(t, user1Groups, fetchedUser1Groups)

	fetchedUser2Groups, err := tdb.Queries.ListGroupsManagedByUser(tdb.Ctx, db.ListGroupsManagedByUserParams{
		UserID: f.User2.ID,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to retrieve User2 managed groups")
	assert.Equal(t, user2Groups, fetchedUser2Groups)
}

func TestIsUserManagerOfGroup(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	addManagersToGroup(t, tdb, []db.ShenUser{f.User1}, f.Group1)
	addManagersToGroup(t, tdb, []db.ShenUser{f.User2}, f.Group2)

	isUser1Manager, err := tdb.Queries.IsUserManagerOfGroup(tdb.Ctx, db.IsUserManagerOfGroupParams{
		UserID:  f.User1.ID,
		GroupID: f.Group1.ID,
	})
	require.NoError(t, err, "Failed to check if User1 is manager of Group1")
	assert.True(t, isUser1Manager, "User1 should be manager of Group1")

	isUser1ManagerOfGroup2, err := tdb.Queries.IsUserManagerOfGroup(tdb.Ctx, db.IsUserManagerOfGroupParams{
		UserID:  f.User1.ID,
		GroupID: f.Group2.ID,
	})
	require.NoError(t, err, "Failed to check if User1 is manager of Group2")
	assert.False(t, isUser1ManagerOfGroup2, "User1 should not be manager of Group2")
}

func TestListAllGroupManagers(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	addManagersToGroup(t, tdb, []db.ShenUser{f.User1, f.User2, f.Admin}, f.Group1)
	addManagersToGroup(t, tdb, []db.ShenUser{f.User1, f.User2}, f.Group2)

	allManagers, err := tdb.Queries.ListAllGroupManagers(tdb.Ctx, db.ListAllGroupManagersParams{
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to retrieve all group managers")

	expected := []struct {
		GroupName string
		Username  string
	}{
		{"test-group-1", "test.admin"},
		{"test-group-1", "test.user1"},
		{"test-group-1", "test.user2"},
		{"test-group-2", "test.user1"},
		{"test-group-2", "test.user2"},
	}

	require.Len(t, allManagers, len(expected), "Should have 5 total group manager assignments")

	for i, exp := range expected {
		assert.Equal(t, exp.GroupName, allManagers[i].GroupName)
		assert.Equal(t, exp.Username, allManagers[i].Username)
	}
}
