//go:build integration

package db_tests

import (
	"slices"
	"testing"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAndGetGroup(t *testing.T) {
	tdb := SetupTestDB(t)

	// Create Group
	created, err := tdb.Queries.CreateGroup(tdb.Ctx, "fight-club")
	require.NoError(t, err, "Failed to create group")

	// Verify created group fields
	assert.Equal(t, "fight-club", created.Name)
	assert.True(t, created.Active, "Group should be active by default")
	assert.NotZero(t, created.ID)
	assert.NotZero(t, created.CreatedAt)

	// Get by ID
	fetchByID, err := tdb.Queries.GetGroupByID(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to fetch by group ID")
	assert.Equal(t, created, fetchByID)

	// Get by Name
	fetchedByName, err := tdb.Queries.GetGroupByName(tdb.Ctx, "fight-club")
	require.NoError(t, err, "Failed to get group by name")
	assert.Equal(t, created, fetchedByName)

	// Deactivate Group
	err = tdb.Queries.DeactivateGroup(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to deactivate group")

	deactivated, err := tdb.Queries.GetGroupByID(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to get deactivated group")
	assert.False(t, deactivated.Active, "Group should be deactivated")
}

func TestUpdateGroup(t *testing.T) {
	tdb := SetupTestDB(t)

	// Create Group
	created, err := tdb.Queries.CreateGroup(tdb.Ctx, "mayhem-project")
	require.NoError(t, err, "Failed to create group")

	// Update group name and active status
	err = tdb.Queries.UpdateGroup(tdb.Ctx, db.UpdateGroupParams{
		ID:     created.ID,
		Name:   "project-mayhem",
		Active: false,
	})
	require.NoError(t, err, "Failed to update group")

	// Verify update
	updated, err := tdb.Queries.GetGroupByID(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to get updated group")
	assert.Equal(t, "project-mayhem", updated.Name)
	assert.False(t, updated.Active)
	assert.Equal(t, created.ID, updated.ID)
}

func TestDeleteGroup(t *testing.T) {
	tdb := SetupTestDB(t)

	// Create Group
	created, err := tdb.Queries.CreateGroup(tdb.Ctx, "space-monkeys")
	require.NoError(t, err, "Failed to create group")

	// Delete group
	err = tdb.Queries.DeleteGroup(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to delete group")

	// Verify deletion - should get error when trying to fetch
	_, err = tdb.Queries.GetGroupByID(tdb.Ctx, created.ID)
	assert.Error(t, err, "Should get error when fetching deleted group")
}

func TestListActiveGroups(t *testing.T) {
	tdb := SetupTestDB(t)

	groupNames := []string{
		"paper-street",
		"remaining-men",
		"fight-club",
	}

	for _, name := range groupNames {
		_, err := tdb.Queries.CreateGroup(tdb.Ctx, name)
		require.NoError(t, err, "Failed to create group")
	}

	slices.Sort(groupNames)

	// first page of groups
	groups, err := tdb.Queries.ListActiveGroups(tdb.Ctx, db.ListActiveGroupsParams{
		Limit:  2,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to list active groups")
	assert.Equal(t, 2, len(groups))
	assert.Equal(t, groupNames[0], groups[0].Name)
	assert.Equal(t, groupNames[1], groups[1].Name)

	// second page of groups
	groups, err = tdb.Queries.ListActiveGroups(tdb.Ctx, db.ListActiveGroupsParams{
		Limit:  2,
		Offset: 2,
	})
	require.NoError(t, err, "Failed to list active groups")
	assert.Equal(t, 1, len(groups))
	assert.Equal(t, groupNames[2], groups[0].Name)

	// deactivate middle group and verify filtering
	groupToDeactivate, err := tdb.Queries.GetGroupByName(tdb.Ctx, groupNames[1])
	require.NoError(t, err, "Failed to get group to deactivate")
	err = tdb.Queries.DeactivateGroup(tdb.Ctx, groupToDeactivate.ID)
	require.NoError(t, err, "Failed to deactivate group")

	groups, err = tdb.Queries.ListActiveGroups(tdb.Ctx, db.ListActiveGroupsParams{
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to list active groups after deactivation")
	assert.Equal(t, 2, len(groups))
	assert.Equal(t, groupNames[0], groups[0].Name)
	assert.Equal(t, groupNames[2], groups[1].Name)
}

func TestListGroups(t *testing.T) {
	tdb := SetupTestDB(t)

	groupNames := []string{
		"arson",
		"assault",
		"mischief",
	}

	for _, name := range groupNames {
		_, err := tdb.Queries.CreateGroup(tdb.Ctx, name)
		require.NoError(t, err, "Failed to create group")
	}

	slices.Sort(groupNames)

	// first page of groups
	groups, err := tdb.Queries.ListGroups(tdb.Ctx, db.ListGroupsParams{
		Limit:  2,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to list groups")
	assert.Equal(t, 2, len(groups))
	assert.Equal(t, groupNames[0], groups[0].Name)
	assert.Equal(t, groupNames[1], groups[1].Name)

	// second page of groups
	groups, err = tdb.Queries.ListGroups(tdb.Ctx, db.ListGroupsParams{
		Limit:  2,
		Offset: 2,
	})
	require.NoError(t, err, "Failed to list groups")
	assert.Equal(t, 1, len(groups))
	assert.Equal(t, groupNames[2], groups[0].Name)

	// deactivate one group
	groupToDeactivate, err := tdb.Queries.GetGroupByName(tdb.Ctx, "assault")
	require.NoError(t, err, "Failed to get group to deactivate")
	err = tdb.Queries.DeactivateGroup(tdb.Ctx, groupToDeactivate.ID)
	require.NoError(t, err, "Failed to deactivate group")

	// ListGroups should return all groups including deactivated
	allGroups, err := tdb.Queries.ListGroups(tdb.Ctx, db.ListGroupsParams{
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to list all groups")
	assert.Equal(t, 3, len(allGroups))
	assert.Equal(t, groupNames[0], allGroups[0].Name)
	assert.Equal(t, groupNames[1], allGroups[1].Name)
	assert.Equal(t, groupNames[2], allGroups[2].Name)
}
