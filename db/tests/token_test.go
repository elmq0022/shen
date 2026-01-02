//go:build integration

package db_tests

import (
	"testing"
	"time"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateToken(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	created, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "api-token",
		HashedToken:   "hash-token-123",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     expiresAt,
	})
	require.NoError(t, err)

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

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	created, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "test-token",
		HashedToken:   "hash-token-456",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     expiresAt,
	})
	require.NoError(t, err)

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

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	created, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "revoke-test",
		HashedToken:   "hash-token-789",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     expiresAt,
	})
	require.NoError(t, err)

	err = tdb.Queries.RevokeToken(tdb.Ctx, created.ID)
	require.NoError(t, err)

	revoked, err := tdb.Queries.GetTokenByID(tdb.Ctx, created.ID)
	require.NoError(t, err)
	assert.True(t, revoked.Revoked)
	assert.True(t, revoked.RevokedAt.Valid)
}

func TestRevokeTokenByHashedToken(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	created, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "revoke-by-hash",
		HashedToken:   "hash-token-abc",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     expiresAt,
	})
	require.NoError(t, err)

	err = tdb.Queries.RevokeTokenByHashedToken(tdb.Ctx, "hash-token-abc")
	require.NoError(t, err)

	revoked, err := tdb.Queries.GetTokenByID(tdb.Ctx, created.ID)
	require.NoError(t, err)
	assert.True(t, revoked.Revoked)
}

func TestRevokeAllUserTokens(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	token1, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "user1-app1-token1",
		HashedToken:   "user1-token-1",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     expiresAt,
	})
	require.NoError(t, err, "Failed to create token 1")

	token2, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "user1-app2-token1",
		HashedToken:   "user1-token-2",
		UserID:        f.User1.ID,
		ApplicationID: f.App2.ID,
		ExpiresAt:     expiresAt,
	})
	require.NoError(t, err, "Failed to create token 2")

	token3, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "user2-app1-token1",
		HashedToken:   "user2-token-1",
		UserID:        f.User2.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     expiresAt,
	})
	require.NoError(t, err, "Failed to create token 3")

	err = tdb.Queries.RevokeAllUserTokens(tdb.Ctx, f.User1.ID)
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

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	token1, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "user1-app1-token-a",
		HashedToken:   "user1-app1-token-1",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     expiresAt,
	})
	require.NoError(t, err, "Failed to create token 1")

	token2, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "user1-app1-token-b",
		HashedToken:   "user1-app1-token-2",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     expiresAt,
	})
	require.NoError(t, err, "Failed to create token 2")

	token3, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "user1-app2-token-a",
		HashedToken:   "user1-app2-token-1",
		UserID:        f.User1.ID,
		ApplicationID: f.App2.ID,
		ExpiresAt:     expiresAt,
	})
	require.NoError(t, err, "Failed to create token 3")

	err = tdb.Queries.RevokeAllUserApplicationTokens(tdb.Ctx, db.RevokeAllUserApplicationTokensParams{
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

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	created, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "delete-test",
		HashedToken:   "hash-delete-123",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     expiresAt,
	})
	require.NoError(t, err)

	err = tdb.Queries.DeleteToken(tdb.Ctx, created.ID)
	require.NoError(t, err)

	_, err = tdb.Queries.GetTokenByID(tdb.Ctx, created.ID)
	assert.Error(t, err, "Should get error when fetching deleted token")
}

func TestDeleteExpiredTokens(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiredTime := pgtype.Timestamp{
		Time:  time.Now().Add(-1 * time.Hour),
		Valid: true,
	}
	expiredToken, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "expired-token",
		HashedToken:   "expired-hash",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     expiredTime,
	})
	require.NoError(t, err, "Failed to create expired token")

	activeTime := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}
	activeToken, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "active-token",
		HashedToken:   "active-hash",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     activeTime,
	})
	require.NoError(t, err, "Failed to create active token")

	err = tdb.Queries.DeleteExpiredTokens(tdb.Ctx)
	require.NoError(t, err, "Failed to delete expired tokens")

	_, err = tdb.Queries.GetTokenByID(tdb.Ctx, expiredToken.ID)
	assert.Error(t, err, "Expired token should be deleted")

	_, err = tdb.Queries.GetTokenByID(tdb.Ctx, activeToken.ID)
	require.NoError(t, err, "Active token should still exist")
}

func TestDeleteRevokedTokens(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	revokedToken, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "revoked-token",
		HashedToken:   "revoked-hash",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     expiresAt,
	})
	require.NoError(t, err, "Failed to create token")

	err = tdb.Queries.RevokeToken(tdb.Ctx, revokedToken.ID)
	require.NoError(t, err, "Failed to revoke token")

	activeToken, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "active-token-2",
		HashedToken:   "active-hash-2",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     expiresAt,
	})
	require.NoError(t, err, "Failed to create active token")

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

	activeTime := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}
	_, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "valid-token",
		HashedToken:   "valid-hash",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     activeTime,
	})
	require.NoError(t, err, "Failed to create valid token")

	valid, err := tdb.Queries.IsTokenValid(tdb.Ctx, "valid-hash")
	require.NoError(t, err)
	assert.True(t, valid)

	expiredTime := pgtype.Timestamp{
		Time:  time.Now().Add(-1 * time.Hour),
		Valid: true,
	}
	_, err = tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "expired-token",
		HashedToken:   "expired-hash",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     expiredTime,
	})
	require.NoError(t, err, "Failed to create expired token")

	valid, err = tdb.Queries.IsTokenValid(tdb.Ctx, "expired-hash")
	require.NoError(t, err)
	assert.False(t, valid)

	revokedToken, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "revoked-token",
		HashedToken:   "revoked-hash",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     activeTime,
	})
	require.NoError(t, err, "Failed to create token for revocation")

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

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	for i := 1; i <= 3; i++ {
		_, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
			Name:          "user1-token-" + string(rune('a'+i)),
			HashedToken:   "user1-hash-" + string(rune('a'+i)),
			UserID:        f.User1.ID,
			ApplicationID: f.App1.ID,
			ExpiresAt:     expiresAt,
		})
		require.NoError(t, err, "Failed to create User1 token")
	}

	for i := 1; i <= 2; i++ {
		_, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
			Name:          "user2-token-" + string(rune('a'+i)),
			HashedToken:   "user2-hash-" + string(rune('a'+i)),
			UserID:        f.User2.ID,
			ApplicationID: f.App1.ID,
			ExpiresAt:     expiresAt,
		})
		require.NoError(t, err, "Failed to create User2 token")
	}

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

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	for i := 1; i <= 3; i++ {
		_, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
			Name:          "user1-app1-" + string(rune('a'+i)),
			HashedToken:   "u1a1-hash-" + string(rune('a'+i)),
			UserID:        f.User1.ID,
			ApplicationID: f.App1.ID,
			ExpiresAt:     expiresAt,
		})
		require.NoError(t, err, "Failed to create User1 App1 token")
	}

	for i := 1; i <= 2; i++ {
		_, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
			Name:          "user1-app2-" + string(rune('a'+i)),
			HashedToken:   "u1a2-hash-" + string(rune('a'+i)),
			UserID:        f.User1.ID,
			ApplicationID: f.App2.ID,
			ExpiresAt:     expiresAt,
		})
		require.NoError(t, err, "Failed to create User1 App2 token")
	}

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

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	_, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "user1-active",
		HashedToken:   "u1-active-hash",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     expiresAt,
	})
	require.NoError(t, err, "Failed to create active token")

	revokedToken, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "user1-revoked",
		HashedToken:   "u1-revoked-hash",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     expiresAt,
	})
	require.NoError(t, err, "Failed to create token")
	err = tdb.Queries.RevokeToken(tdb.Ctx, revokedToken.ID)
	require.NoError(t, err, "Failed to revoke token")

	expiredTime := pgtype.Timestamp{
		Time:  time.Now().Add(-1 * time.Hour),
		Valid: true,
	}
	_, err = tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "user1-expired",
		HashedToken:   "u1-expired-hash",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     expiredTime,
	})
	require.NoError(t, err, "Failed to create expired token")

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

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	for i := 1; i <= 3; i++ {
		_, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
			Name:          "user1-app1-all-" + string(rune('a'+i)),
			HashedToken:   "u1a1-all-" + string(rune('a'+i)),
			UserID:        f.User1.ID,
			ApplicationID: f.App1.ID,
			ExpiresAt:     expiresAt,
		})
		require.NoError(t, err, "Failed to create User1 App1 token")
	}

	for i := 1; i <= 2; i++ {
		_, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
			Name:          "user1-app2-all-" + string(rune('a'+i)),
			HashedToken:   "u1a2-all-" + string(rune('a'+i)),
			UserID:        f.User1.ID,
			ApplicationID: f.App2.ID,
			ExpiresAt:     expiresAt,
		})
		require.NoError(t, err, "Failed to create User1 App2 token")
	}

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

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	for i := 1; i <= 3; i++ {
		_, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
			Name:          "app1-token-" + string(rune('a'+i)),
			HashedToken:   "app1-hash-" + string(rune('a'+i)),
			UserID:        f.User1.ID,
			ApplicationID: f.App1.ID,
			ExpiresAt:     expiresAt,
		})
		require.NoError(t, err, "Failed to create App1 token")
	}

	for i := 1; i <= 2; i++ {
		_, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
			Name:          "app2-token-" + string(rune('a'+i)),
			HashedToken:   "app2-hash-" + string(rune('a'+i)),
			UserID:        f.User1.ID,
			ApplicationID: f.App2.ID,
			ExpiresAt:     expiresAt,
		})
		require.NoError(t, err, "Failed to create App2 token")
	}

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

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	for i := 1; i <= 3; i++ {
		_, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
			Name:          "count-user1-" + string(rune('a'+i)),
			HashedToken:   "count-u1-" + string(rune('a'+i)),
			UserID:        f.User1.ID,
			ApplicationID: f.App1.ID,
			ExpiresAt:     expiresAt,
		})
		require.NoError(t, err, "Failed to create token")
	}

	revokedToken, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "count-user1-revoked",
		HashedToken:   "count-u1-revoked",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     expiresAt,
	})
	require.NoError(t, err, "Failed to create token")
	err = tdb.Queries.RevokeToken(tdb.Ctx, revokedToken.ID)
	require.NoError(t, err, "Failed to revoke token")

	count, err := tdb.Queries.CountActiveTokensByUser(tdb.Ctx, f.User1.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count, "Should only count active tokens")
}

func TestCountTokensByUser(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	for i := 1; i <= 2; i++ {
		_, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
			Name:          "count-all-user1-" + string(rune('a'+i)),
			HashedToken:   "count-all-u1-" + string(rune('a'+i)),
			UserID:        f.User1.ID,
			ApplicationID: f.App1.ID,
			ExpiresAt:     expiresAt,
		})
		require.NoError(t, err, "Failed to create token")
	}

	revokedToken, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "count-all-revoked",
		HashedToken:   "count-all-revoked",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     expiresAt,
	})
	require.NoError(t, err, "Failed to create token")
	err = tdb.Queries.RevokeToken(tdb.Ctx, revokedToken.ID)
	require.NoError(t, err, "Failed to revoke token")

	expiredTime := pgtype.Timestamp{
		Time:  time.Now().Add(-1 * time.Hour),
		Valid: true,
	}
	_, err = tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          "count-all-expired",
		HashedToken:   "count-all-expired",
		UserID:        f.User1.ID,
		ApplicationID: f.App1.ID,
		ExpiresAt:     expiredTime,
	})
	require.NoError(t, err, "Failed to create expired token")

	count, err := tdb.Queries.CountTokensByUser(tdb.Ctx, f.User1.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(4), count, "Should count all tokens including revoked and expired")
}

func TestCountTokensByUserApplication(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	for i := 1; i <= 3; i++ {
		_, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
			Name:          "count-u1a1-" + string(rune('a'+i)),
			HashedToken:   "count-u1a1-hash-" + string(rune('a'+i)),
			UserID:        f.User1.ID,
			ApplicationID: f.App1.ID,
			ExpiresAt:     expiresAt,
		})
		require.NoError(t, err, "Failed to create token")
	}

	for i := 1; i <= 2; i++ {
		_, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
			Name:          "count-u1a2-" + string(rune('a'+i)),
			HashedToken:   "count-u1a2-hash-" + string(rune('a'+i)),
			UserID:        f.User1.ID,
			ApplicationID: f.App2.ID,
			ExpiresAt:     expiresAt,
		})
		require.NoError(t, err, "Failed to create token")
	}

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

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	for i := 1; i <= 4; i++ {
		_, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
			Name:          "count-app1-" + string(rune('a'+i)),
			HashedToken:   "count-app1-hash-" + string(rune('a'+i)),
			UserID:        f.User1.ID,
			ApplicationID: f.App1.ID,
			ExpiresAt:     expiresAt,
		})
		require.NoError(t, err, "Failed to create token")
	}

	for i := 1; i <= 2; i++ {
		_, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
			Name:          "count-app2-" + string(rune('a'+i)),
			HashedToken:   "count-app2-hash-" + string(rune('a'+i)),
			UserID:        f.User1.ID,
			ApplicationID: f.App2.ID,
			ExpiresAt:     expiresAt,
		})
		require.NoError(t, err, "Failed to create token")
	}

	count, err := tdb.Queries.CountTokensByApplication(tdb.Ctx, f.App1.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(4), count)

	count, err = tdb.Queries.CountTokensByApplication(tdb.Ctx, f.App2.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}
