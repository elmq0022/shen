//go:build integration

package db_tests

import (
	"testing"

	"github.com/elmq0022/shen/internal/bootstrap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAdmin(t *testing.T) {
	mockHash := func(password string) (string, error) {
		return "hashed_" + password, nil
	}

	t.Run("creates admin user when no users exist", func(t *testing.T) {
		tdb := SetupTestDB(t)

		bootstrap.CreateAdmin(tdb.Ctx, tdb.Queries, mockHash)

		count, err := tdb.Queries.CountUsers(tdb.Ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		admin, err := tdb.Queries.GetUserByUsername(tdb.Ctx, bootstrap.DefaultAdminUser)
		require.NoError(t, err)
		assert.Equal(t, bootstrap.DefaultAdminUser, admin.Username)
		assert.True(t, admin.HashedPassword.Valid)
		assert.Equal(t, "hashed_"+bootstrap.DefaultAdminPassword, admin.HashedPassword.String)
		assert.True(t, admin.Active)

		role, err := tdb.Queries.GetRoleByName(tdb.Ctx, "admin")
		require.NoError(t, err)
		assert.Equal(t, role.ID, admin.Role)
	})

	t.Run("creates admin with custom username from env var", func(t *testing.T) {
		customUser := "superadmin"
		t.Setenv(bootstrap.AdminUserEnv, customUser)
		tdb := SetupTestDB(t)

		bootstrap.CreateAdmin(tdb.Ctx, tdb.Queries, mockHash)

		admin, err := tdb.Queries.GetUserByUsername(tdb.Ctx, customUser)
		require.NoError(t, err)
		assert.Equal(t, customUser, admin.Username)
	})

	t.Run("creates admin with custom password from env var", func(t *testing.T) {
		customPassword := "super_secret_password"
		t.Setenv(bootstrap.AdminPasswordEnv, customPassword)
		tdb := SetupTestDB(t)

		bootstrap.CreateAdmin(tdb.Ctx, tdb.Queries, mockHash)

		admin, err := tdb.Queries.GetUserByUsername(tdb.Ctx, bootstrap.DefaultAdminUser)
		require.NoError(t, err)
		assert.Equal(t, "hashed_"+customPassword, admin.HashedPassword.String)
	})

	t.Run("does not create admin when users already exist", func(t *testing.T) {
		tdb := SetupTestDB(t)

		existingUser := CreateTestUser(t, tdb, "existing_user", RoleUser)

		bootstrap.CreateAdmin(tdb.Ctx, tdb.Queries, mockHash)

		count, err := tdb.Queries.CountUsers(tdb.Ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		user, err := tdb.Queries.GetUserByID(tdb.Ctx, existingUser.ID)
		require.NoError(t, err)
		assert.Equal(t, existingUser.Username, user.Username)

		_, err = tdb.Queries.GetUserByUsername(tdb.Ctx, bootstrap.DefaultAdminUser)
		assert.Error(t, err)
	})

	t.Run("does not create admin when admin already exists", func(t *testing.T) {
		tdb := SetupTestDB(t)

		bootstrap.CreateAdmin(tdb.Ctx, tdb.Queries, mockHash)

		firstAdmin, err := tdb.Queries.GetUserByUsername(tdb.Ctx, bootstrap.DefaultAdminUser)
		require.NoError(t, err)

		bootstrap.CreateAdmin(tdb.Ctx, tdb.Queries, mockHash)

		count, err := tdb.Queries.CountUsers(tdb.Ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		secondAdmin, err := tdb.Queries.GetUserByUsername(tdb.Ctx, bootstrap.DefaultAdminUser)
		require.NoError(t, err)
		assert.Equal(t, firstAdmin.ID, secondAdmin.ID)
		assert.Equal(t, firstAdmin.CreatedAt, secondAdmin.CreatedAt)
	})
}
