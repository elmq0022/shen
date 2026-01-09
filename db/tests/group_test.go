//go:build integration

package db_tests

import (
	"slices"
	"testing"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateGroup(t *testing.T) {
	tdb := SetupTestDB(t)

	created := CreateTestGroup(t, tdb, "my-group")

	assert.Equal(t, "my-group", created.Name)
	assert.True(t, created.Active, "Group should be active by default")
	assert.NotZero(t, created.ID)
	assert.NotZero(t, created.CreatedAt)
}

func TestGetGroup(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	// Get by ID
	fetchedByID, err := tdb.Queries.GetGroupByID(tdb.Ctx, f.Group1.ID)
	require.NoError(t, err, "Failed to fetch by group ID")
	assert.Equal(t, f.Group1, fetchedByID)

	// Get by Name
	fetchedByName, err := tdb.Queries.GetGroupByName(tdb.Ctx, f.Group1.Name)
	require.NoError(t, err, "Failed to get group by name")
	assert.Equal(t, f.Group1, fetchedByName)
}

func TestDeactivateGroup(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	err := tdb.Queries.DeactivateGroup(tdb.Ctx, f.Group1.ID)
	require.NoError(t, err, "Failed to deactivate group")

	deactivated, err := tdb.Queries.GetGroupByID(tdb.Ctx, f.Group1.ID)
	require.NoError(t, err, "Failed to get deactivated group")
	assert.False(t, deactivated.Active, "Group should be deactivated")
}

func TestUpdateGroup(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	// Update group name and active status
	err := tdb.Queries.UpdateGroup(tdb.Ctx, db.UpdateGroupParams{
		ID:     f.Group1.ID,
		Name:   "updated-group",
		Active: false,
	})
	require.NoError(t, err, "Failed to update group")

	// Verify update
	updated, err := tdb.Queries.GetGroupByID(tdb.Ctx, f.Group1.ID)
	require.NoError(t, err, "Failed to get updated group")
	assert.Equal(t, "updated-group", updated.Name)
	assert.False(t, updated.Active)
	assert.Equal(t, f.Group1.ID, updated.ID)
}

func TestDeleteGroup(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	// Delete group
	err := tdb.Queries.DeleteGroup(tdb.Ctx, f.Group1.ID)
	require.NoError(t, err, "Failed to delete group")

	// Verify deletion - should get error when trying to fetch
	_, err = tdb.Queries.GetGroupByID(tdb.Ctx, f.Group1.ID)
	assert.Error(t, err, "Should get error when fetching deleted group")
}

func TestListActiveGroups(t *testing.T) {
	tdb := SetupTestDB(t)

	// Create test groups
	groups := CreateTestGroups(t, tdb, "group", 3)

	groupNames := []string{"group-1", "group-2", "group-3"}
	slices.Sort(groupNames)

	// first page of groups (cursor-based)
	page1, err := tdb.Queries.ListActiveGroups(tdb.Ctx, db.ListActiveGroupsParams{
		Column1: "",
		Limit:   2,
	})
	require.NoError(t, err, "Failed to list active groups")
	assert.Len(t, page1, 2)
	assert.Equal(t, groupNames[0], page1[0].Name)
	assert.Equal(t, groupNames[1], page1[1].Name)

	// second page of groups (cursor-based)
	page2, err := tdb.Queries.ListActiveGroups(tdb.Ctx, db.ListActiveGroupsParams{
		Column1: page1[len(page1)-1].Name,
		Limit:   2,
	})
	require.NoError(t, err, "Failed to list active groups")
	assert.Len(t, page2, 1)
	assert.Equal(t, groupNames[2], page2[0].Name)

	// deactivate middle group and verify filtering
	err = tdb.Queries.DeactivateGroup(tdb.Ctx, groups[1].ID)
	require.NoError(t, err, "Failed to deactivate group")

	activeGroups, err := tdb.Queries.ListActiveGroups(tdb.Ctx, db.ListActiveGroupsParams{
		Column1: "",
		Limit:   10,
	})
	require.NoError(t, err, "Failed to list active groups after deactivation")
	assert.Len(t, activeGroups, 2)
	assert.Equal(t, groupNames[0], activeGroups[0].Name)
	assert.Equal(t, groupNames[2], activeGroups[1].Name)
}

func TestListGroups(t *testing.T) {
	tdb := SetupTestDB(t)

	// Create test groups
	groups := CreateTestGroups(t, tdb, "group", 3)

	groupNames := []string{"group-1", "group-2", "group-3"}
	slices.Sort(groupNames)

	// first page of groups (cursor-based)
	page1, err := tdb.Queries.ListGroups(tdb.Ctx, db.ListGroupsParams{
		Column1: "",
		Limit:   2,
	})
	require.NoError(t, err, "Failed to list groups")
	assert.Len(t, page1, 2)
	assert.Equal(t, groupNames[0], page1[0].Name)
	assert.Equal(t, groupNames[1], page1[1].Name)

	// second page of groups (cursor-based)
	page2, err := tdb.Queries.ListGroups(tdb.Ctx, db.ListGroupsParams{
		Column1: page1[len(page1)-1].Name,
		Limit:   2,
	})
	require.NoError(t, err, "Failed to list groups")
	assert.Len(t, page2, 1)
	assert.Equal(t, groupNames[2], page2[0].Name)

	// deactivate one group
	err = tdb.Queries.DeactivateGroup(tdb.Ctx, groups[1].ID)
	require.NoError(t, err, "Failed to deactivate group")

	// ListGroups should return all groups including deactivated (cursor-based)
	allGroups, err := tdb.Queries.ListGroups(tdb.Ctx, db.ListGroupsParams{
		Column1: "",
		Limit:   10,
	})
	require.NoError(t, err, "Failed to list all groups")
	assert.Len(t, allGroups, 3)
	assert.Equal(t, groupNames[0], allGroups[0].Name)
	assert.Equal(t, groupNames[1], allGroups[1].Name)
	assert.Equal(t, groupNames[2], allGroups[2].Name)
}
