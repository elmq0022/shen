//go:build integration

package db_tests

import (
	"encoding/base64"
	"testing"

	"github.com/elmq0022/shen/internal/bootstrap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateInitialKeys(t *testing.T) {
	// Generate a test KEK (32 bytes base64 encoded)
	testKEK := base64.StdEncoding.EncodeToString([]byte("12345678901234567890123456789012"))

	t.Run("generates initial key when no keys exist", func(t *testing.T) {
		t.Setenv("SHEN_KEY_ENCRYPTION_KEY", testKEK)
		tdb := SetupTestDB(t)

		bootstrap.GenerateInitialKeys(tdb.Ctx, tdb.Queries)

		count, err := tdb.Queries.CountJWTKeys(tdb.Ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		activeKey, err := tdb.Queries.GetActiveSigningKey(tdb.Ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, activeKey.Kid, "KID should be generated")
		assert.NotEmpty(t, activeKey.EncryptedPrivateKey, "Encrypted private key should be generated")
		assert.NotEmpty(t, activeKey.PublicKey, "Public key should be generated")
		assert.True(t, activeKey.ActiveForSigning, "Key should be active for signing")
		assert.True(t, activeKey.ActiveForVerification, "Key should be active for verification")
		assert.NotZero(t, activeKey.CreatedAt, "Created timestamp should be set")
	})

	t.Run("does not generate key when keys already exist", func(t *testing.T) {
		t.Setenv("SHEN_KEY_ENCRYPTION_KEY", testKEK)
		tdb := SetupTestDB(t)

		existingKey := CreateTestJWTKey(t, tdb, "existing-key", []byte("encrypted-key"), "public-key", true, true)

		bootstrap.GenerateInitialKeys(tdb.Ctx, tdb.Queries)

		count, err := tdb.Queries.CountJWTKeys(tdb.Ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "Should not create additional keys")

		fetchedKey, err := tdb.Queries.GetJWTKeyByID(tdb.Ctx, existingKey.ID)
		require.NoError(t, err)
		assert.Equal(t, existingKey.Kid, fetchedKey.Kid, "Existing key should remain unchanged")
	})

	t.Run("does not generate additional keys on subsequent calls", func(t *testing.T) {
		t.Setenv("SHEN_KEY_ENCRYPTION_KEY", testKEK)
		tdb := SetupTestDB(t)

		bootstrap.GenerateInitialKeys(tdb.Ctx, tdb.Queries)

		firstKey, err := tdb.Queries.GetActiveSigningKey(tdb.Ctx)
		require.NoError(t, err)

		bootstrap.GenerateInitialKeys(tdb.Ctx, tdb.Queries)

		count, err := tdb.Queries.CountJWTKeys(tdb.Ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "Should not create duplicate keys")

		secondKey, err := tdb.Queries.GetActiveSigningKey(tdb.Ctx)
		require.NoError(t, err)
		assert.Equal(t, firstKey.ID, secondKey.ID, "Same key should be returned")
		assert.Equal(t, firstKey.CreatedAt, secondKey.CreatedAt, "Created timestamp should be unchanged")
	})

	t.Run("generates valid RSA key pair", func(t *testing.T) {
		t.Setenv("SHEN_KEY_ENCRYPTION_KEY", testKEK)
		tdb := SetupTestDB(t)

		bootstrap.GenerateInitialKeys(tdb.Ctx, tdb.Queries)

		activeKey, err := tdb.Queries.GetActiveSigningKey(tdb.Ctx)
		require.NoError(t, err)

		// Verify public key format
		assert.Contains(t, activeKey.PublicKey, "BEGIN PUBLIC KEY", "Public key should be in PEM format")
		assert.Contains(t, activeKey.PublicKey, "END PUBLIC KEY", "Public key should be in PEM format")

		// Verify encrypted private key is not empty and looks encrypted
		assert.Greater(t, len(activeKey.EncryptedPrivateKey), 100, "Encrypted private key should have substantial length")

		// Verify KID format (ULID - 26 characters)
		assert.Len(t, activeKey.Kid, 26, "KID should be in ULID format (26 characters)")
	})

	t.Run("panics when KEK is not set", func(t *testing.T) {
		t.Setenv("SHEN_KEY_ENCRYPTION_KEY", "")
		tdb := SetupTestDB(t)

		assert.Panics(t, func() {
			bootstrap.GenerateInitialKeys(tdb.Ctx, tdb.Queries)
		}, "Should panic when KEK is not set")
	})

	t.Run("panics when KEK is invalid", func(t *testing.T) {
		t.Setenv("SHEN_KEY_ENCRYPTION_KEY", "invalid-base64!!!")
		tdb := SetupTestDB(t)

		assert.Panics(t, func() {
			bootstrap.GenerateInitialKeys(tdb.Ctx, tdb.Queries)
		}, "Should panic when KEK is not valid base64")
	})

	t.Run("panics when KEK is wrong length", func(t *testing.T) {
		// KEK must be exactly 32 bytes
		shortKEK := base64.StdEncoding.EncodeToString([]byte("short"))
		t.Setenv("SHEN_KEY_ENCRYPTION_KEY", shortKEK)
		tdb := SetupTestDB(t)

		assert.Panics(t, func() {
			bootstrap.GenerateInitialKeys(tdb.Ctx, tdb.Queries)
		}, "Should panic when KEK is not 32 bytes")
	})
}
