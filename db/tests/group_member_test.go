//go:build integration

package db_tests

import (
	"fmt"
	"slices"
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

func addUsersToGroup(t *testing.T, tdb *TestDB, users []db.ShenUser, group db.ShenGroup) {
	t.Helper()
	for _, user := range users {
		_, err := tdb.Queries.AddUserToGroup(tdb.Ctx, db.AddUserToGroupParams{
			UserID:  user.ID,
			GroupID: group.ID,
		})
		require.NoError(
			t, err,
			fmt.Sprintf("Failed to add user: %s to group: %s", user.Username, group.Name),
		)
	}
}

func sortUsersByUsername(users []db.ShenUser) {
	slices.SortFunc(users, func(a, b db.ShenUser) int {
		if a.Username < b.Username {
			return -1
		} else if a.Username == b.Username {
			return 0
		}
		return 1
	})
}

func TestListUsersByGroup(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	group1Members := []db.ShenUser{f.User1, f.User2, f.Admin}
	addUsersToGroup(t, tdb, group1Members, f.Group1)
	sortUsersByUsername(group1Members)

	group2Members := []db.ShenUser{f.User1, f.User2}
	addUsersToGroup(t, tdb, group2Members, f.Group2)
	sortUsersByUsername(group2Members)

	fetchedGroup1Members, err := tdb.Queries.ListUsersByGroup(tdb.Ctx, db.ListUsersByGroupParams{
		GroupID: f.Group1.ID,
		Limit:   10,
		Offset:  0,
	})
	require.NoError(t, err, "Failed to retrieve Group1 members")
	assert.Equal(t, group1Members, fetchedGroup1Members)

	fetchedGroup2Members, err := tdb.Queries.ListUsersByGroup(tdb.Ctx, db.ListUsersByGroupParams{
		GroupID: f.Group2.ID,
		Limit:   10,
		Offset:  0,
	})
	require.NoError(t, err, "Failed to retrieve Group2 members")
	assert.Equal(t, group2Members, fetchedGroup2Members)
}
