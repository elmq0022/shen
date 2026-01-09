//go:build integration

package db_tests

import (
	"testing"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateGroupMember(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	created, err := tdb.Queries.AddUserToGroup(tdb.Ctx, db.AddUserToGroupParams{
		UserID:  f.User1.ID,
		GroupID: f.Group1.ID,
	})
	require.NoError(t, err, "Failed to add member to group")

	fetched, err := tdb.Queries.GetUserGroupMemberByID(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to retrieve group member")
	assert.Equal(t, created, fetched)

	err = tdb.Queries.RemoveUserFromGroup(tdb.Ctx, db.RemoveUserFromGroupParams{
		UserID:  f.User1.ID,
		GroupID: f.Group1.ID,
	})
	require.NoError(t, err, "Failed to remove user from group")

	_, err = tdb.Queries.GetUserGroupMemberByID(tdb.Ctx, created.ID)
	require.Error(t, err, "Group member should not exist after removal")
}

func TestListUsersByGroup(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	users := CreateTestUsers(t, tdb, "member", 3)
	allMembers := []db.ShenUser{f.User1, f.User2, f.Admin, users[0], users[1], users[2]}
	addUsersToGroup(t, tdb, allMembers, f.Group1)
	sortUsersByUsername(allMembers)

	page1, err := tdb.Queries.ListUsersByGroup(tdb.Ctx, db.ListUsersByGroupParams{
		GroupID: f.Group1.ID,
		Limit:   2,
		Column2: "",
	})
	require.NoError(t, err, "Failed to list group members")
	assert.Len(t, page1, 2)
	assert.Equal(t, allMembers[0].Username, page1[0].Username)
	assert.Equal(t, allMembers[1].Username, page1[1].Username)

	page2, err := tdb.Queries.ListUsersByGroup(tdb.Ctx, db.ListUsersByGroupParams{
		GroupID: f.Group1.ID,
		Limit:   2,
		Column2: page1[len(page1)-1].Username,
	})
	require.NoError(t, err, "Failed to list group members")
	assert.Len(t, page2, 2)
	assert.Equal(t, allMembers[2].Username, page2[0].Username)
	assert.Equal(t, allMembers[3].Username, page2[1].Username)

	page3, err := tdb.Queries.ListUsersByGroup(tdb.Ctx, db.ListUsersByGroupParams{
		GroupID: f.Group1.ID,
		Limit:   2,
		Column2: page2[len(page2)-1].Username,
	})
	require.NoError(t, err, "Failed to list group members")
	assert.Len(t, page3, 2)
	assert.Equal(t, allMembers[4].Username, page3[0].Username)
	assert.Equal(t, allMembers[5].Username, page3[1].Username)

	allFetched, err := tdb.Queries.ListUsersByGroup(tdb.Ctx, db.ListUsersByGroupParams{
		GroupID: f.Group1.ID,
		Limit:   10,
		Column2: "",
	})
	require.NoError(t, err, "Failed to retrieve all Group1 members")
	assert.Equal(t, allMembers, allFetched)
}

func TestListGroupsByUser(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	groups := CreateTestGroups(t, tdb, "team", 3)
	allGroups := []db.ShenGroup{f.Group1, f.Group2, groups[0], groups[1], groups[2]}

	for _, group := range allGroups {
		addUsersToGroup(t, tdb, []db.ShenUser{f.User1}, group)
	}
	sortGroupsByName(allGroups)

	page1, err := tdb.Queries.ListGroupsByUser(tdb.Ctx, db.ListGroupsByUserParams{
		UserID:  f.User1.ID,
		Limit:   2,
		Column2: "",
	})
	require.NoError(t, err, "Failed to list user groups")
	assert.Len(t, page1, 2)
	assert.Equal(t, allGroups[0].Name, page1[0].Name)
	assert.Equal(t, allGroups[1].Name, page1[1].Name)

	page2, err := tdb.Queries.ListGroupsByUser(tdb.Ctx, db.ListGroupsByUserParams{
		UserID:  f.User1.ID,
		Limit:   2,
		Column2: page1[len(page1)-1].Name,
	})
	require.NoError(t, err, "Failed to list user groups")
	assert.Len(t, page2, 2)
	assert.Equal(t, allGroups[2].Name, page2[0].Name)
	assert.Equal(t, allGroups[3].Name, page2[1].Name)

	page3, err := tdb.Queries.ListGroupsByUser(tdb.Ctx, db.ListGroupsByUserParams{
		UserID:  f.User1.ID,
		Limit:   2,
		Column2: page2[len(page2)-1].Name,
	})
	require.NoError(t, err, "Failed to list user groups")
	assert.Len(t, page3, 1)
	assert.Equal(t, allGroups[4].Name, page3[0].Name)

	allFetched, err := tdb.Queries.ListGroupsByUser(tdb.Ctx, db.ListGroupsByUserParams{
		UserID:  f.User1.ID,
		Limit:   10,
		Column2: "",
	})
	require.NoError(t, err, "Failed to retrieve all User1 groups")
	assert.Equal(t, allGroups, allFetched)
}

func TestIsUserInGroup(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	addUsersToGroup(t, tdb, []db.ShenUser{f.User1}, f.Group1)
	addUsersToGroup(t, tdb, []db.ShenUser{f.User2}, f.Group2)

	isUser1InGroup1, err := tdb.Queries.IsUserInGroup(tdb.Ctx, db.IsUserInGroupParams{
		UserID:  f.User1.ID,
		GroupID: f.Group1.ID,
	})
	require.NoError(t, err, "Failed to check if User1 is in Group1")
	assert.True(t, isUser1InGroup1, "User1 should be in Group1")

	isUser1InGroup2, err := tdb.Queries.IsUserInGroup(tdb.Ctx, db.IsUserInGroupParams{
		UserID:  f.User1.ID,
		GroupID: f.Group2.ID,
	})
	require.NoError(t, err, "Failed to check if User1 is in Group2")
	assert.False(t, isUser1InGroup2, "User1 should not be in Group2")
}

func TestCountUsersByGroup(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	users := CreateTestUsers(t, tdb, "member", 3)
	allMembers := []db.ShenUser{f.User1, f.User2, f.Admin, users[0], users[1], users[2]}
	addUsersToGroup(t, tdb, allMembers, f.Group1)

	count, err := tdb.Queries.CountUsersByGroup(tdb.Ctx, f.Group1.ID)
	require.NoError(t, err, "Failed to count users by group")
	assert.Equal(t, int64(6), count)

	addUsersToGroup(t, tdb, []db.ShenUser{f.User1, f.User2}, f.Group2)

	count, err = tdb.Queries.CountUsersByGroup(tdb.Ctx, f.Group2.ID)
	require.NoError(t, err, "Failed to count users in Group2")
	assert.Equal(t, int64(2), count)
}

func TestCountGroupsByUser(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	groups := CreateTestGroups(t, tdb, "team", 3)
	allGroups := []db.ShenGroup{f.Group1, f.Group2, groups[0], groups[1], groups[2]}

	for _, group := range allGroups {
		addUsersToGroup(t, tdb, []db.ShenUser{f.User1}, group)
	}

	count, err := tdb.Queries.CountGroupsByUser(tdb.Ctx, f.User1.ID)
	require.NoError(t, err, "Failed to count groups by user")
	assert.Equal(t, int64(5), count)

	addUsersToGroup(t, tdb, []db.ShenUser{f.User2}, f.Group1)

	count, err = tdb.Queries.CountGroupsByUser(tdb.Ctx, f.User2.ID)
	require.NoError(t, err, "Failed to count groups for User2")
	assert.Equal(t, int64(1), count)
}

func TestCountAllGroupMembers(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	addUsersToGroup(t, tdb, []db.ShenUser{f.User1, f.User2, f.Admin}, f.Group1)
	addUsersToGroup(t, tdb, []db.ShenUser{f.User1, f.User2}, f.Group2)

	count, err := tdb.Queries.CountAllGroupMembers(tdb.Ctx)
	require.NoError(t, err, "Failed to count all group members")
	assert.Equal(t, int64(5), count)
}

func TestListAllGroupMembers(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	addUsersToGroup(t, tdb, []db.ShenUser{f.User1, f.User2, f.Admin}, f.Group1)
	addUsersToGroup(t, tdb, []db.ShenUser{f.User1, f.User2}, f.Group2)

	allMembers, err := tdb.Queries.ListAllGroupMembers(tdb.Ctx, db.ListAllGroupMembersParams{
		Limit:    10,
		Column1:  "",
		Username: "",
	})
	require.NoError(t, err, "Failed to retrieve all group members")

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

	require.Len(t, allMembers, len(expected), "Should have 5 total group memberships")

	for i, exp := range expected {
		assert.Equal(t, exp.GroupName, allMembers[i].GroupName)
		assert.Equal(t, exp.Username, allMembers[i].Username)
	}
}
