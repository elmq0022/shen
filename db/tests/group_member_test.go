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

	td := CreateStandardFixtures(t, tdb)

	created, err := tdb.Queries.AddUserToGroup(tdb.Ctx, db.AddUserToGroupParams{
		UserID:  td.User1.ID,
		GroupID: td.Group1.ID,
	})
	require.NoError(t, err, "Failed to add member to group")

	fetched, err := tdb.Queries.GetUserGroupMemberByID(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to retrieve group member")
	assert.Equal(t, created, fetched)

	err = tdb.Queries.RemoveUserFromGroup(tdb.Ctx, db.RemoveUserFromGroupParams{
		UserID:  td.User1.ID,
		GroupID: td.Group1.ID,
	})
	require.NoError(t, err, "Failed to remove user from group")

	_, err = tdb.Queries.GetUserGroupMemberByID(tdb.Ctx, created.ID)
	require.Error(t, err, "Group member should not exist after removal")
}
