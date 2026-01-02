//go:build integration

package db_tests

import (
	"testing"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateToken(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	created := CreateTestToken(t, tdb, "api-token", "hash-token-123", f.User1.ID, f.App1.ID, GetActiveExpiresAt())

	assert.Equal(t, "api-token", created.Name)
	assert.Equal(t, "hash-token-123", created.HashedToken)
	assert.Equal(t, f.User1.ID, created.UserID)
	assert.Equal(t, f.App1.ID, created.ApplicationID)
	assert.False(t, created.Revoked)
	assert.False(t, created.RevokedAt.Valid)
	assert.NotZero(t, created.ID)
	assert.NotZero(t, created.CreatedAt)
	assert.True(t, created.ExpiresAt.Valid)
}

func TestGetToken(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	created := CreateTestToken(t, tdb, "test-token", "hash-token-456", f.User1.ID, f.App1.ID, GetActiveExpiresAt())

	fetchedByID, err := tdb.Queries.GetTokenByID(tdb.Ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created, fetchedByID)

	fetchedByHashedToken, err := tdb.Queries.GetTokenByHashedToken(tdb.Ctx, "hash-token-456")
	require.NoError(t, err)
	assert.Equal(t, created, fetchedByHashedToken)

	fetchedByUserApplicationName, err := tdb.Queries.GetTokenByUserApplicationName(tdb.Ctx, db.GetTokenByUserApplicationNameParams{
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		Name:          "test-token",
	})
	require.NoError(t, err)
	assert.Equal(t, created, fetchedByUserApplicationName)

	_, err = tdb.Queries.GetTokenByID(tdb.Ctx, 999)
	assert.Error(t, err)
}

func TestRevokeToken(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	created := CreateTestToken(t, tdb, "revoke-test", "hash-token-789", f.User1.ID, f.App1.ID, GetActiveExpiresAt())

	err := tdb.Queries.RevokeToken(tdb.Ctx, created.ID)
	require.NoError(t, err)

	revoked, err := tdb.Queries.GetTokenByID(tdb.Ctx, created.ID)
	require.NoError(t, err)
	assert.True(t, revoked.Revoked)
	assert.True(t, revoked.RevokedAt.Valid)
}

func TestRevokeTokenByHashedToken(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	created := CreateTestToken(t, tdb, "revoke-by-hash", "hash-token-abc", f.User1.ID, f.App1.ID, GetActiveExpiresAt())

	err := tdb.Queries.RevokeTokenByHashedToken(tdb.Ctx, "hash-token-abc")
	require.NoError(t, err)

	revoked, err := tdb.Queries.GetTokenByID(tdb.Ctx, created.ID)
	require.NoError(t, err)
	assert.True(t, revoked.Revoked)
}

func TestRevokeAllUserTokens(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := GetActiveExpiresAt()

	token1 := CreateTestToken(t, tdb, "user1-app1-token1", "user1-token-1", f.User1.ID, f.App1.ID, expiresAt)
	token2 := CreateTestToken(t, tdb, "user1-app2-token1", "user1-token-2", f.User1.ID, f.App2.ID, expiresAt)
	token3 := CreateTestToken(t, tdb, "user2-app1-token1", "user2-token-1", f.User2.ID, f.App1.ID, expiresAt)

	err := tdb.Queries.RevokeAllUserTokens(tdb.Ctx, f.User1.ID)
	require.NoError(t, err, "Failed to revoke all user tokens")

	revokedToken1, err := tdb.Queries.GetTokenByID(tdb.Ctx, token1.ID)
	require.NoError(t, err)
	assert.True(t, revokedToken1.Revoked)

	revokedToken2, err := tdb.Queries.GetTokenByID(tdb.Ctx, token2.ID)
	require.NoError(t, err)
	assert.True(t, revokedToken2.Revoked)

	activeToken3, err := tdb.Queries.GetTokenByID(tdb.Ctx, token3.ID)
	require.NoError(t, err)
	assert.False(t, activeToken3.Revoked, "User2 token should not be revoked")
}

func TestRevokeAllUserApplicationTokens(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := GetActiveExpiresAt()

	token1 := CreateTestToken(t, tdb, "user1-app1-token-a", "user1-app1-token-1", f.User1.ID, f.App1.ID, expiresAt)
	token2 := CreateTestToken(t, tdb, "user1-app1-token-b", "user1-app1-token-2", f.User1.ID, f.App1.ID, expiresAt)
	token3 := CreateTestToken(t, tdb, "user1-app2-token-a", "user1-app2-token-1", f.User1.ID, f.App2.ID, expiresAt)

	err := tdb.Queries.RevokeAllUserApplicationTokens(tdb.Ctx, db.RevokeAllUserApplicationTokensParams{
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
	})
	require.NoError(t, err, "Failed to revoke user application tokens")

	revokedToken1, err := tdb.Queries.GetTokenByID(tdb.Ctx, token1.ID)
	require.NoError(t, err)
	assert.True(t, revokedToken1.Revoked)

	revokedToken2, err := tdb.Queries.GetTokenByID(tdb.Ctx, token2.ID)
	require.NoError(t, err)
	assert.True(t, revokedToken2.Revoked)

	activeToken3, err := tdb.Queries.GetTokenByID(tdb.Ctx, token3.ID)
	require.NoError(t, err)
	assert.False(t, activeToken3.Revoked, "User1 App2 token should not be revoked")
}

func TestDeleteToken(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	created := CreateTestToken(t, tdb, "delete-test", "hash-delete-123", f.User1.ID, f.App1.ID, GetActiveExpiresAt())

	err := tdb.Queries.DeleteToken(tdb.Ctx, created.ID)
	require.NoError(t, err)

	_, err = tdb.Queries.GetTokenByID(tdb.Ctx, created.ID)
	assert.Error(t, err, "Should get error when fetching deleted token")
}

func TestDeleteExpiredTokens(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiredToken := CreateTestToken(t, tdb, "expired-token", "expired-hash", f.User1.ID, f.App1.ID, GetExpiredExpiresAt())
	activeToken := CreateTestToken(t, tdb, "active-token", "active-hash", f.User1.ID, f.App1.ID, GetActiveExpiresAt())

	err := tdb.Queries.DeleteExpiredTokens(tdb.Ctx)
	require.NoError(t, err, "Failed to delete expired tokens")

	_, err = tdb.Queries.GetTokenByID(tdb.Ctx, expiredToken.ID)
	assert.Error(t, err, "Expired token should be deleted")

	_, err = tdb.Queries.GetTokenByID(tdb.Ctx, activeToken.ID)
	require.NoError(t, err, "Active token should still exist")
}

func TestDeleteRevokedTokens(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	revokedToken := CreateTestToken(t, tdb, "revoked-token", "revoked-hash", f.User1.ID, f.App1.ID, GetActiveExpiresAt())
	err := tdb.Queries.RevokeToken(tdb.Ctx, revokedToken.ID)
	require.NoError(t, err, "Failed to revoke token")

	activeToken := CreateTestToken(t, tdb, "active-token-2", "active-hash-2", f.User1.ID, f.App1.ID, GetActiveExpiresAt())

	err = tdb.Queries.DeleteRevokedTokens(tdb.Ctx)
	require.NoError(t, err, "Failed to delete revoked tokens")

	_, err = tdb.Queries.GetTokenByID(tdb.Ctx, revokedToken.ID)
	assert.Error(t, err, "Revoked token should be deleted")

	_, err = tdb.Queries.GetTokenByID(tdb.Ctx, activeToken.ID)
	require.NoError(t, err, "Active token should still exist")
}

func TestIsTokenValid(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	CreateTestToken(t, tdb, "valid-token", "valid-hash", f.User1.ID, f.App1.ID, GetActiveExpiresAt())

	valid, err := tdb.Queries.IsTokenValid(tdb.Ctx, "valid-hash")
	require.NoError(t, err)
	assert.True(t, valid)

	CreateTestToken(t, tdb, "expired-token", "expired-hash", f.User1.ID, f.App1.ID, GetExpiredExpiresAt())

	valid, err = tdb.Queries.IsTokenValid(tdb.Ctx, "expired-hash")
	require.NoError(t, err)
	assert.False(t, valid)

	revokedToken := CreateTestToken(t, tdb, "revoked-token", "revoked-hash", f.User1.ID, f.App1.ID, GetActiveExpiresAt())
	err = tdb.Queries.RevokeToken(tdb.Ctx, revokedToken.ID)
	require.NoError(t, err, "Failed to revoke token")

	valid, err = tdb.Queries.IsTokenValid(tdb.Ctx, "revoked-hash")
	require.NoError(t, err)
	assert.False(t, valid)

	valid, err = tdb.Queries.IsTokenValid(tdb.Ctx, "nonexistent-hash")
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestListActiveTokensByUser(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := GetActiveExpiresAt()

	CreateTestTokens(t, tdb, "user1-token", 3, f.User1.ID, f.App1.ID, expiresAt)
	CreateTestTokens(t, tdb, "user2-token", 2, f.User2.ID, f.App1.ID, expiresAt)

	user1Tokens, err := tdb.Queries.ListActiveTokensByUser(tdb.Ctx, db.ListActiveTokensByUserParams{
		UserID: f.User1.ID,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to list User1 tokens")
	assert.Len(t, user1Tokens, 3)

	user2Tokens, err := tdb.Queries.ListActiveTokensByUser(tdb.Ctx, db.ListActiveTokensByUserParams{
		UserID: f.User2.ID,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to list User2 tokens")
	assert.Len(t, user2Tokens, 2)
}

func TestListActiveTokensByUserApplication(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := GetActiveExpiresAt()

	CreateTestTokens(t, tdb, "user1-app1", 3, f.User1.ID, f.App1.ID, expiresAt)
	CreateTestTokens(t, tdb, "user1-app2", 2, f.User1.ID, f.App2.ID, expiresAt)

	user1App1Tokens, err := tdb.Queries.ListActiveTokensByUserApplication(tdb.Ctx, db.ListActiveTokensByUserApplicationParams{
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		Limit:         10,
		Offset:        0,
	})
	require.NoError(t, err, "Failed to list User1 App1 tokens")
	assert.Len(t, user1App1Tokens, 3)

	user1App2Tokens, err := tdb.Queries.ListActiveTokensByUserApplication(tdb.Ctx, db.ListActiveTokensByUserApplicationParams{
		UserID:        f.User1.ID,
		ApplicationID: f.App2.ID,
		Limit:         10,
		Offset:        0,
	})
	require.NoError(t, err, "Failed to list User1 App2 tokens")
	assert.Len(t, user1App2Tokens, 2)
}

func TestListTokensByUser(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	CreateTestToken(t, tdb, "user1-active", "u1-active-hash", f.User1.ID, f.App1.ID, GetActiveExpiresAt())

	revokedToken := CreateTestToken(t, tdb, "user1-revoked", "u1-revoked-hash", f.User1.ID, f.App1.ID, GetActiveExpiresAt())
	err := tdb.Queries.RevokeToken(tdb.Ctx, revokedToken.ID)
	require.NoError(t, err, "Failed to revoke token")

	CreateTestToken(t, tdb, "user1-expired", "u1-expired-hash", f.User1.ID, f.App1.ID, GetExpiredExpiresAt())

	allTokens, err := tdb.Queries.ListTokensByUser(tdb.Ctx, db.ListTokensByUserParams{
		UserID: f.User1.ID,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to list all User1 tokens")
	assert.Len(t, allTokens, 3, "Should return all tokens including revoked and expired")
}

func TestListTokensByUserApplication(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := GetActiveExpiresAt()

	CreateTestTokens(t, tdb, "user1-app1-all", 3, f.User1.ID, f.App1.ID, expiresAt)
	CreateTestTokens(t, tdb, "user1-app2-all", 2, f.User1.ID, f.App2.ID, expiresAt)

	user1App1Tokens, err := tdb.Queries.ListTokensByUserApplication(tdb.Ctx, db.ListTokensByUserApplicationParams{
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		Limit:         10,
		Offset:        0,
	})
	require.NoError(t, err, "Failed to list User1 App1 tokens")
	assert.Len(t, user1App1Tokens, 3)

	user1App2Tokens, err := tdb.Queries.ListTokensByUserApplication(tdb.Ctx, db.ListTokensByUserApplicationParams{
		UserID:        f.User1.ID,
		ApplicationID: f.App2.ID,
		Limit:         10,
		Offset:        0,
	})
	require.NoError(t, err, "Failed to list User1 App2 tokens")
	assert.Len(t, user1App2Tokens, 2)
}

func TestListTokensByApplication(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := GetActiveExpiresAt()

	CreateTestTokens(t, tdb, "app1-token", 3, f.User1.ID, f.App1.ID, expiresAt)
	CreateTestTokens(t, tdb, "app2-token", 2, f.User1.ID, f.App2.ID, expiresAt)

	app1Tokens, err := tdb.Queries.ListTokensByApplication(tdb.Ctx, db.ListTokensByApplicationParams{
		ApplicationID: f.App1.ID,
		Limit:         10,
		Offset:        0,
	})
	require.NoError(t, err, "Failed to list App1 tokens")
	assert.Len(t, app1Tokens, 3)

	app2Tokens, err := tdb.Queries.ListTokensByApplication(tdb.Ctx, db.ListTokensByApplicationParams{
		ApplicationID: f.App2.ID,
		Limit:         10,
		Offset:        0,
	})
	require.NoError(t, err, "Failed to list App2 tokens")
	assert.Len(t, app2Tokens, 2)
}

func TestCountActiveTokensByUser(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := GetActiveExpiresAt()

	CreateTestTokens(t, tdb, "count-user1", 3, f.User1.ID, f.App1.ID, expiresAt)

	revokedToken := CreateTestToken(t, tdb, "count-user1-revoked", "count-u1-revoked", f.User1.ID, f.App1.ID, expiresAt)
	err := tdb.Queries.RevokeToken(tdb.Ctx, revokedToken.ID)
	require.NoError(t, err, "Failed to revoke token")

	count, err := tdb.Queries.CountActiveTokensByUser(tdb.Ctx, f.User1.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count, "Should only count active tokens")
}

func TestCountTokensByUser(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	CreateTestTokens(t, tdb, "count-all-user1", 2, f.User1.ID, f.App1.ID, GetActiveExpiresAt())

	revokedToken := CreateTestToken(t, tdb, "count-all-revoked", "count-all-revoked", f.User1.ID, f.App1.ID, GetActiveExpiresAt())
	err := tdb.Queries.RevokeToken(tdb.Ctx, revokedToken.ID)
	require.NoError(t, err, "Failed to revoke token")

	CreateTestToken(t, tdb, "count-all-expired", "count-all-expired", f.User1.ID, f.App1.ID, GetExpiredExpiresAt())

	count, err := tdb.Queries.CountTokensByUser(tdb.Ctx, f.User1.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(4), count, "Should count all tokens including revoked and expired")
}

func TestCountTokensByUserApplication(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := GetActiveExpiresAt()

	CreateTestTokens(t, tdb, "count-u1a1", 3, f.User1.ID, f.App1.ID, expiresAt)
	CreateTestTokens(t, tdb, "count-u1a2", 2, f.User1.ID, f.App2.ID, expiresAt)

	count, err := tdb.Queries.CountTokensByUserApplication(tdb.Ctx, db.CountTokensByUserApplicationParams{
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	count, err = tdb.Queries.CountTokensByUserApplication(tdb.Ctx, db.CountTokensByUserApplicationParams{
		UserID:        f.User1.ID,
		ApplicationID: f.App2.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestCountTokensByApplication(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := GetActiveExpiresAt()

	CreateTestTokens(t, tdb, "count-app1", 4, f.User1.ID, f.App1.ID, expiresAt)
	CreateTestTokens(t, tdb, "count-app2", 2, f.User1.ID, f.App2.ID, expiresAt)

	count, err := tdb.Queries.CountTokensByApplication(tdb.Ctx, f.App1.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(4), count)

	count, err = tdb.Queries.CountTokensByApplication(tdb.Ctx, f.App2.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}
