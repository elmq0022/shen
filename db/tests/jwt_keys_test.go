//go:build integration

package db_tests

import (
	"testing"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateJWTKey(t *testing.T) {
	tdb := SetupTestDB(t)

	created, err := tdb.Queries.CreateJWTKey(tdb.Ctx, db.CreateJWTKeyParams{
		Kid:                  "2025-01-04",
		EncryptedPrivateKey:  []byte("encrypted-private-key-data"),
		PublicKey:            "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...",
		ActiveForSigning:     true,
		ActiveForVerification: true,
	})
	require.NoError(t, err)

	assert.Equal(t, "2025-01-04", created.Kid)
	assert.Equal(t, []byte("encrypted-private-key-data"), created.EncryptedPrivateKey)
	assert.True(t, created.ActiveForSigning)
	assert.True(t, created.ActiveForVerification)
	assert.NotZero(t, created.ID)
	assert.NotZero(t, created.CreatedAt)
}

func TestGetJWTKeyByID(t *testing.T) {
	tdb := SetupTestDB(t)

	created := CreateTestJWTKey(t, tdb, "2025-01-04-get", []byte("encrypted-key"), "public-key", true, true)

	fetchedByID, err := tdb.Queries.GetJWTKeyByID(tdb.Ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created, fetchedByID)

	_, err = tdb.Queries.GetJWTKeyByID(tdb.Ctx, 999)
	assert.Error(t, err, "Should get error when fetching non-existent key")
}

func TestGetJWTKeyByKID(t *testing.T) {
	tdb := SetupTestDB(t)

	created := CreateTestJWTKey(t, tdb, "2025-01-04-kid", []byte("encrypted-key"), "public-key", true, true)

	fetchedByKID, err := tdb.Queries.GetJWTKeyByKID(tdb.Ctx, "2025-01-04-kid")
	require.NoError(t, err)
	assert.Equal(t, created, fetchedByKID)

	_, err = tdb.Queries.GetJWTKeyByKID(tdb.Ctx, "non-existent-kid")
	assert.Error(t, err, "Should get error when fetching non-existent kid")
}

func TestGetActiveSigningKey(t *testing.T) {
	tdb := SetupTestDB(t)

	CreateTestJWTKey(t, tdb, "2025-01-01-old", []byte("old-key"), "old-public", false, true)
	activeKey := CreateTestJWTKey(t, tdb, "2025-01-04-active", []byte("active-key"), "active-public", true, true)

	fetched, err := tdb.Queries.GetActiveSigningKey(tdb.Ctx)
	require.NoError(t, err)
	assert.Equal(t, activeKey.Kid, fetched.Kid)
	assert.True(t, fetched.ActiveForSigning)
}

func TestGetActiveSigningKeyWhenNoneActive(t *testing.T) {
	tdb := SetupTestDB(t)

	CreateTestJWTKey(t, tdb, "2025-01-01-inactive", []byte("inactive-key"), "inactive-public", false, true)

	_, err := tdb.Queries.GetActiveSigningKey(tdb.Ctx)
	assert.Error(t, err, "Should get error when no signing key is active")
}

func TestListJWTKeys(t *testing.T) {
	tdb := SetupTestDB(t)

	CreateTestJWTKeys(t, tdb, "key", 5, []byte("encrypted"), "public", true, true)

	// First page - no cursor
	page1, err := tdb.Queries.ListJWTKeys(tdb.Ctx, db.ListJWTKeysParams{
		CursorKid: "",
		Limit:     2,
	})
	require.NoError(t, err, "Failed to list first page of JWT keys")
	assert.Len(t, page1, 2)

	// Second page - use cursor from last item of page1
	page2, err := tdb.Queries.ListJWTKeys(tdb.Ctx, db.ListJWTKeysParams{
		CursorKid: page1[1].Kid,
		Limit:     2,
	})
	require.NoError(t, err, "Failed to list second page of JWT keys")
	assert.Len(t, page2, 2)

	// Get all keys - no cursor, high limit
	allKeys, err := tdb.Queries.ListJWTKeys(tdb.Ctx, db.ListJWTKeysParams{
		CursorKid: "",
		Limit:     10,
	})
	require.NoError(t, err, "Failed to list all JWT keys")
	assert.Len(t, allKeys, 5)
}

func TestListJWTKeysOrdering(t *testing.T) {
	tdb := SetupTestDB(t)

	key1 := CreateTestJWTKey(t, tdb, "2025-01-01", []byte("key1"), "public1", false, true)
	key2 := CreateTestJWTKey(t, tdb, "2025-01-02", []byte("key2"), "public2", false, true)
	key3 := CreateTestJWTKey(t, tdb, "2025-01-03", []byte("key3"), "public3", true, true)

	allKeys, err := tdb.Queries.ListJWTKeys(tdb.Ctx, db.ListJWTKeysParams{
		CursorKid: "",
		Limit:     10,
	})
	require.NoError(t, err)

	// Ordered by kid DESC (newest first since ULID is time-sortable)
	assert.Equal(t, key3.ID, allKeys[0].ID, "Newest key should be first")
	assert.Equal(t, key2.ID, allKeys[1].ID, "Second newest key should be second")
	assert.Equal(t, key1.ID, allKeys[2].ID, "Oldest key should be last")
}

func TestListActiveVerificationKeys(t *testing.T) {
	tdb := SetupTestDB(t)

	activeKey1 := CreateTestJWTKey(t, tdb, "2025-01-01-verify", []byte("key1"), "public1", false, true)
	activeKey2 := CreateTestJWTKey(t, tdb, "2025-01-04-verify", []byte("key2"), "public2", true, true)
	CreateTestJWTKey(t, tdb, "2025-01-02-inactive", []byte("key3"), "public3", false, false)

	allActiveKeys, err := tdb.Queries.ListActiveVerificationKeys(tdb.Ctx)
	require.NoError(t, err)
	assert.Len(t, allActiveKeys, 2, "Should only return keys active for verification")

	assert.Equal(t, activeKey2.ID, allActiveKeys[0].ID, "Newest key should be first")
	assert.Equal(t, activeKey1.ID, allActiveKeys[1].ID, "Older key should be second")
}

func TestDeactivateKeyForSigning(t *testing.T) {
	tdb := SetupTestDB(t)

	created := CreateTestJWTKey(t, tdb, "2025-01-04-deactivate", []byte("key"), "public", true, true)

	err := tdb.Queries.DeactivateKeyForSigning(tdb.Ctx, created.ID)
	require.NoError(t, err)

	deactivated, err := tdb.Queries.GetJWTKeyByID(tdb.Ctx, created.ID)
	require.NoError(t, err)
	assert.False(t, deactivated.ActiveForSigning)
	assert.True(t, deactivated.ActiveForVerification, "Verification flag should remain unchanged")
}

func TestDeactivateKeyForVerification(t *testing.T) {
	tdb := SetupTestDB(t)

	created := CreateTestJWTKey(t, tdb, "2025-01-04-deactivate-verify", []byte("key"), "public", false, true)

	err := tdb.Queries.DeactivateKeyForVerification(tdb.Ctx, created.ID)
	require.NoError(t, err)

	deactivated, err := tdb.Queries.GetJWTKeyByID(tdb.Ctx, created.ID)
	require.NoError(t, err)
	assert.False(t, deactivated.ActiveForVerification)
	assert.False(t, deactivated.ActiveForSigning)
}

func TestDeactivateAllKeysForSigning(t *testing.T) {
	tdb := SetupTestDB(t)

	key1 := CreateTestJWTKey(t, tdb, "2025-01-01", []byte("key1"), "public1", true, true)
	key2 := CreateTestJWTKey(t, tdb, "2025-01-02", []byte("key2"), "public2", true, true)
	key3 := CreateTestJWTKey(t, tdb, "2025-01-03", []byte("key3"), "public3", false, true)

	err := tdb.Queries.DeactivateAllKeysForSigning(tdb.Ctx)
	require.NoError(t, err)

	deactivated1, err := tdb.Queries.GetJWTKeyByID(tdb.Ctx, key1.ID)
	require.NoError(t, err)
	assert.False(t, deactivated1.ActiveForSigning)
	assert.True(t, deactivated1.ActiveForVerification, "Verification should remain active")

	deactivated2, err := tdb.Queries.GetJWTKeyByID(tdb.Ctx, key2.ID)
	require.NoError(t, err)
	assert.False(t, deactivated2.ActiveForSigning)

	unchanged3, err := tdb.Queries.GetJWTKeyByID(tdb.Ctx, key3.ID)
	require.NoError(t, err)
	assert.False(t, unchanged3.ActiveForSigning, "Already inactive key should remain unchanged")
}

func TestDeleteJWTKey(t *testing.T) {
	tdb := SetupTestDB(t)

	created := CreateTestJWTKey(t, tdb, "2025-01-04-delete", []byte("key"), "public", true, true)

	err := tdb.Queries.DeleteJWTKey(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to delete JWT key")

	_, err = tdb.Queries.GetJWTKeyByID(tdb.Ctx, created.ID)
	assert.Error(t, err, "Should get error when fetching deleted key")
}

func TestCountJWTKeys(t *testing.T) {
	tdb := SetupTestDB(t)

	CreateTestJWTKeys(t, tdb, "count-key", 3, []byte("encrypted"), "public", true, true)

	count, err := tdb.Queries.CountJWTKeys(tdb.Ctx)
	require.NoError(t, err, "Failed to count JWT keys")
	assert.Equal(t, int64(3), count)
}

func TestKeyRotationWorkflow(t *testing.T) {
	tdb := SetupTestDB(t)

	t.Run("initial key setup", func(t *testing.T) {
		initialKey := CreateTestJWTKey(t, tdb, "2025-01-01-initial", []byte("key1"), "public1", true, true)

		activeKey, err := tdb.Queries.GetActiveSigningKey(tdb.Ctx)
		require.NoError(t, err)
		assert.Equal(t, initialKey.Kid, activeKey.Kid)
	})

	t.Run("rotate to new key", func(t *testing.T) {
		err := tdb.Queries.DeactivateAllKeysForSigning(tdb.Ctx)
		require.NoError(t, err, "Failed to deactivate old signing keys")

		newKey := CreateTestJWTKey(t, tdb, "2025-01-04-rotated", []byte("key2"), "public2", true, true)

		activeKey, err := tdb.Queries.GetActiveSigningKey(tdb.Ctx)
		require.NoError(t, err)
		assert.Equal(t, newKey.Kid, activeKey.Kid)

		allVerificationKeys, err := tdb.Queries.ListActiveVerificationKeys(tdb.Ctx)
		require.NoError(t, err)
		assert.Len(t, allVerificationKeys, 2, "Both old and new keys should be active for verification")
	})

	t.Run("deactivate old verification key", func(t *testing.T) {
		oldKey, err := tdb.Queries.GetJWTKeyByKID(tdb.Ctx, "2025-01-01-initial")
		require.NoError(t, err)

		err = tdb.Queries.DeactivateKeyForVerification(tdb.Ctx, oldKey.ID)
		require.NoError(t, err)

		allVerificationKeys, err := tdb.Queries.ListActiveVerificationKeys(tdb.Ctx)
		require.NoError(t, err)
		assert.Len(t, allVerificationKeys, 1, "Only new key should be active for verification")
		assert.Equal(t, "2025-01-04-rotated", allVerificationKeys[0].Kid)
	})
}

func TestUniqueKIDConstraint(t *testing.T) {
	tdb := SetupTestDB(t)

	CreateTestJWTKey(t, tdb, "2025-01-04-unique", []byte("key1"), "public1", true, true)

	_, err := tdb.Queries.CreateJWTKey(tdb.Ctx, db.CreateJWTKeyParams{
		Kid:                   "2025-01-04-unique",
		EncryptedPrivateKey:   []byte("key2"),
		PublicKey:             "public2",
		ActiveForSigning:      true,
		ActiveForVerification: true,
	})
	assert.Error(t, err, "Should get error when creating key with duplicate kid")
}

func TestMultipleActiveVerificationKeys(t *testing.T) {
	tdb := SetupTestDB(t)

	CreateTestJWTKey(t, tdb, "2025-01-01", []byte("key1"), "public1", false, true)
	CreateTestJWTKey(t, tdb, "2025-01-02", []byte("key2"), "public2", false, true)
	CreateTestJWTKey(t, tdb, "2025-01-03", []byte("key3"), "public3", false, true)
	CreateTestJWTKey(t, tdb, "2025-01-04", []byte("key4"), "public4", true, true)

	allVerificationKeys, err := tdb.Queries.ListActiveVerificationKeys(tdb.Ctx)
	require.NoError(t, err)
	assert.Len(t, allVerificationKeys, 4, "Multiple keys can be active for verification simultaneously")
}
