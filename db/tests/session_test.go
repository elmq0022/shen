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

func TestCreateSession(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	created, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "hash-token-123",
		UserID:      f.User1.ID,
		ExpiresAt:   expiresAt,
	})
	require.NoError(t, err)

	assert.Equal(t, "hash-token-123", created.HashedToken)
	assert.Equal(t, f.User1.ID, created.UserID)
	assert.False(t, created.Revoked)
	assert.False(t, created.RevokedAt.Valid)
	assert.NotZero(t, created.ID)
	assert.NotZero(t, created.CreatedAt)
	assert.True(t, created.ExpiresAt.Valid)
}

func TestGetSession(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	created, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "hash-token-456",
		UserID:      f.User1.ID,
		ExpiresAt:   expiresAt,
	})
	require.NoError(t, err)

	fetchedByID, err := tdb.Queries.GetSessionByID(tdb.Ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created, fetchedByID)

	fetchedByHashedToken, err := tdb.Queries.GetSessionByHashedToken(tdb.Ctx, "hash-token-456")
	require.NoError(t, err)
	assert.Equal(t, created, fetchedByHashedToken)

	_, err = tdb.Queries.GetSessionByID(tdb.Ctx, 999)
	assert.Error(t, err)
}

func TestRevokeSession(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	created, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "hash-token-789",
		UserID:      f.User1.ID,
		ExpiresAt:   expiresAt,
	})
	require.NoError(t, err)

	err = tdb.Queries.RevokeSession(tdb.Ctx, created.ID)
	require.NoError(t, err)

	revoked, err := tdb.Queries.GetSessionByID(tdb.Ctx, created.ID)
	require.NoError(t, err)
	assert.True(t, revoked.Revoked)
	assert.True(t, revoked.RevokedAt.Valid)
}

func TestRevokeSessionByHashedToken(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	created, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "hash-token-abc",
		UserID:      f.User1.ID,
		ExpiresAt:   expiresAt,
	})
	require.NoError(t, err)

	err = tdb.Queries.RevokeSessionByHashedToken(tdb.Ctx, "hash-token-abc")
	require.NoError(t, err)

	revoked, err := tdb.Queries.GetSessionByID(tdb.Ctx, created.ID)
	require.NoError(t, err)
	assert.True(t, revoked.Revoked)
}

func TestRevokeAllUserSessions(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	session1, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "user1-session-1",
		UserID:      f.User1.ID,
		ExpiresAt:   expiresAt,
	})
	require.NoError(t, err, "Failed to create session 1")

	session2, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "user1-session-2",
		UserID:      f.User1.ID,
		ExpiresAt:   expiresAt,
	})
	require.NoError(t, err, "Failed to create session 2")

	session3, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "user2-session-1",
		UserID:      f.User2.ID,
		ExpiresAt:   expiresAt,
	})
	require.NoError(t, err, "Failed to create session 3")

	err = tdb.Queries.RevokeAllUserSessions(tdb.Ctx, f.User1.ID)
	require.NoError(t, err, "Failed to revoke all user sessions")

	revokedSession1, err := tdb.Queries.GetSessionByID(tdb.Ctx, session1.ID)
	require.NoError(t, err)
	assert.True(t, revokedSession1.Revoked)

	revokedSession2, err := tdb.Queries.GetSessionByID(tdb.Ctx, session2.ID)
	require.NoError(t, err)
	assert.True(t, revokedSession2.Revoked)

	activeSession3, err := tdb.Queries.GetSessionByID(tdb.Ctx, session3.ID)
	require.NoError(t, err)
	assert.False(t, activeSession3.Revoked, "User2 session should not be revoked")
}

func TestDeleteSession(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	created, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "hash-token-delete",
		UserID:      f.User1.ID,
		ExpiresAt:   expiresAt,
	})
	require.NoError(t, err, "Failed to create session")

	err = tdb.Queries.DeleteSession(tdb.Ctx, created.ID)
	require.NoError(t, err, "Failed to delete session")

	_, err = tdb.Queries.GetSessionByID(tdb.Ctx, created.ID)
	assert.Error(t, err, "Should get error when fetching deleted session")
}

func TestDeleteExpiredSessions(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiredTime := pgtype.Timestamp{
		Time:  time.Now().Add(-1 * time.Hour),
		Valid: true,
	}
	expiredSession, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "expired-session",
		UserID:      f.User1.ID,
		ExpiresAt:   expiredTime,
	})
	require.NoError(t, err, "Failed to create expired session")

	activeTime := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}
	activeSession, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "active-session",
		UserID:      f.User1.ID,
		ExpiresAt:   activeTime,
	})
	require.NoError(t, err, "Failed to create active session")

	err = tdb.Queries.DeleteExpiredSessions(tdb.Ctx)
	require.NoError(t, err, "Failed to delete expired sessions")

	_, err = tdb.Queries.GetSessionByID(tdb.Ctx, expiredSession.ID)
	assert.Error(t, err, "Expired session should be deleted")

	_, err = tdb.Queries.GetSessionByID(tdb.Ctx, activeSession.ID)
	require.NoError(t, err, "Active session should still exist")
}

func TestDeleteRevokedSessions(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	revokedSession, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "revoked-session",
		UserID:      f.User1.ID,
		ExpiresAt:   expiresAt,
	})
	require.NoError(t, err, "Failed to create session")

	err = tdb.Queries.RevokeSession(tdb.Ctx, revokedSession.ID)
	require.NoError(t, err, "Failed to revoke session")

	activeSession, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "active-session-2",
		UserID:      f.User1.ID,
		ExpiresAt:   expiresAt,
	})
	require.NoError(t, err, "Failed to create active session")

	err = tdb.Queries.DeleteRevokedSessions(tdb.Ctx)
	require.NoError(t, err, "Failed to delete revoked sessions")

	_, err = tdb.Queries.GetSessionByID(tdb.Ctx, revokedSession.ID)
	assert.Error(t, err, "Revoked session should be deleted")

	_, err = tdb.Queries.GetSessionByID(tdb.Ctx, activeSession.ID)
	require.NoError(t, err, "Active session should still exist")
}

func TestIsSessionValid(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	t.Run("valid active session", func(t *testing.T) {
		expiresAt := pgtype.Timestamp{
			Time:  time.Now().Add(24 * time.Hour),
			Valid: true,
		}

		_, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
			HashedToken: "valid-session",
			UserID:      f.User1.ID,
			ExpiresAt:   expiresAt,
		})
		require.NoError(t, err, "Failed to create session")

		isValid, err := tdb.Queries.IsSessionValid(tdb.Ctx, "valid-session")
		require.NoError(t, err, "Failed to check session validity")
		assert.True(t, isValid, "Session should be valid")
	})

	t.Run("revoked session", func(t *testing.T) {
		expiresAt := pgtype.Timestamp{
			Time:  time.Now().Add(24 * time.Hour),
			Valid: true,
		}

		created, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
			HashedToken: "revoked-check",
			UserID:      f.User1.ID,
			ExpiresAt:   expiresAt,
		})
		require.NoError(t, err, "Failed to create session")

		err = tdb.Queries.RevokeSession(tdb.Ctx, created.ID)
		require.NoError(t, err, "Failed to revoke session")

		isValid, err := tdb.Queries.IsSessionValid(tdb.Ctx, "revoked-check")
		require.NoError(t, err, "Failed to check session validity")
		assert.False(t, isValid, "Revoked session should not be valid")
	})

	t.Run("expired session", func(t *testing.T) {
		expiredTime := pgtype.Timestamp{
			Time:  time.Now().Add(-1 * time.Hour),
			Valid: true,
		}

		_, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
			HashedToken: "expired-check",
			UserID:      f.User1.ID,
			ExpiresAt:   expiredTime,
		})
		require.NoError(t, err, "Failed to create session")

		isValid, err := tdb.Queries.IsSessionValid(tdb.Ctx, "expired-check")
		require.NoError(t, err, "Failed to check session validity")
		assert.False(t, isValid, "Expired session should not be valid")
	})

	t.Run("non-existent session", func(t *testing.T) {
		isValid, err := tdb.Queries.IsSessionValid(tdb.Ctx, "non-existent")
		require.NoError(t, err, "Failed to check session validity")
		assert.False(t, isValid, "Non-existent session should not be valid")
	})
}

func TestListActiveSessions(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	for i := 1; i <= 5; i++ {
		_, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
			HashedToken: "active-session-" + string(rune('a'+i)),
			UserID:      f.User1.ID,
			ExpiresAt:   expiresAt,
		})
		require.NoError(t, err, "Failed to create session")
	}

	revokedSession, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "revoked-list",
		UserID:      f.User1.ID,
		ExpiresAt:   expiresAt,
	})
	require.NoError(t, err, "Failed to create session")
	err = tdb.Queries.RevokeSession(tdb.Ctx, revokedSession.ID)
	require.NoError(t, err, "Failed to revoke session")

	expiredTime := pgtype.Timestamp{
		Time:  time.Now().Add(-1 * time.Hour),
		Valid: true,
	}
	_, err = tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "expired-list",
		UserID:      f.User1.ID,
		ExpiresAt:   expiredTime,
	})
	require.NoError(t, err, "Failed to create expired session")

	page1, err := tdb.Queries.ListActiveSessions(tdb.Ctx, db.ListActiveSessionsParams{
		Limit:  2,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to list first page of active sessions")
	assert.Len(t, page1, 2)

	page2, err := tdb.Queries.ListActiveSessions(tdb.Ctx, db.ListActiveSessionsParams{
		Limit:  2,
		Offset: 2,
	})
	require.NoError(t, err, "Failed to list second page of active sessions")
	assert.Len(t, page2, 2)

	allActiveSessions, err := tdb.Queries.ListActiveSessions(tdb.Ctx, db.ListActiveSessionsParams{
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to list all active sessions")
	assert.Len(t, allActiveSessions, 5, "Should only have 5 active sessions")
}

func TestListActiveSessionsByUser(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	for i := 1; i <= 3; i++ {
		_, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
			HashedToken: "user1-session-" + string(rune('a'+i)),
			UserID:      f.User1.ID,
			ExpiresAt:   expiresAt,
		})
		require.NoError(t, err, "Failed to create User1 session")
	}

	for i := 1; i <= 2; i++ {
		_, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
			HashedToken: "user2-session-" + string(rune('a'+i)),
			UserID:      f.User2.ID,
			ExpiresAt:   expiresAt,
		})
		require.NoError(t, err, "Failed to create User2 session")
	}

	user1Sessions, err := tdb.Queries.ListActiveSessionsByUser(tdb.Ctx, db.ListActiveSessionsByUserParams{
		UserID: f.User1.ID,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to list User1 sessions")
	assert.Len(t, user1Sessions, 3)

	user2Sessions, err := tdb.Queries.ListActiveSessionsByUser(tdb.Ctx, db.ListActiveSessionsByUserParams{
		UserID: f.User2.ID,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to list User2 sessions")
	assert.Len(t, user2Sessions, 2)
}

func TestListSessionsByUser(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	_, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "user1-active",
		UserID:      f.User1.ID,
		ExpiresAt:   expiresAt,
	})
	require.NoError(t, err, "Failed to create active session")

	revokedSession, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "user1-revoked",
		UserID:      f.User1.ID,
		ExpiresAt:   expiresAt,
	})
	require.NoError(t, err, "Failed to create session")
	err = tdb.Queries.RevokeSession(tdb.Ctx, revokedSession.ID)
	require.NoError(t, err, "Failed to revoke session")

	expiredTime := pgtype.Timestamp{
		Time:  time.Now().Add(-1 * time.Hour),
		Valid: true,
	}
	_, err = tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "user1-expired",
		UserID:      f.User1.ID,
		ExpiresAt:   expiredTime,
	})
	require.NoError(t, err, "Failed to create expired session")

	allSessions, err := tdb.Queries.ListSessionsByUser(tdb.Ctx, db.ListSessionsByUserParams{
		UserID: f.User1.ID,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to list all User1 sessions")
	assert.Len(t, allSessions, 3, "Should return all sessions including revoked and expired")
}

func TestCountActiveSessionsByUser(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	for i := 1; i <= 3; i++ {
		_, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
			HashedToken: "count-user1-" + string(rune('a'+i)),
			UserID:      f.User1.ID,
			ExpiresAt:   expiresAt,
		})
		require.NoError(t, err, "Failed to create session")
	}

	revokedSession, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "count-user1-revoked",
		UserID:      f.User1.ID,
		ExpiresAt:   expiresAt,
	})
	require.NoError(t, err, "Failed to create session")
	err = tdb.Queries.RevokeSession(tdb.Ctx, revokedSession.ID)
	require.NoError(t, err, "Failed to revoke session")

	count, err := tdb.Queries.CountActiveSessionsByUser(tdb.Ctx, f.User1.ID)
	require.NoError(t, err, "Failed to count active sessions")
	assert.Equal(t, int64(3), count, "Should only count active sessions")
}

func TestCountSessionsByUser(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	for i := 1; i <= 2; i++ {
		_, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
			HashedToken: "count-all-user1-" + string(rune('a'+i)),
			UserID:      f.User1.ID,
			ExpiresAt:   expiresAt,
		})
		require.NoError(t, err, "Failed to create session")
	}

	revokedSession, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "count-all-revoked",
		UserID:      f.User1.ID,
		ExpiresAt:   expiresAt,
	})
	require.NoError(t, err, "Failed to create session")
	err = tdb.Queries.RevokeSession(tdb.Ctx, revokedSession.ID)
	require.NoError(t, err, "Failed to revoke session")

	expiredTime := pgtype.Timestamp{
		Time:  time.Now().Add(-1 * time.Hour),
		Valid: true,
	}
	_, err = tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "count-all-expired",
		UserID:      f.User1.ID,
		ExpiresAt:   expiredTime,
	})
	require.NoError(t, err, "Failed to create expired session")

	count, err := tdb.Queries.CountSessionsByUser(tdb.Ctx, f.User1.ID)
	require.NoError(t, err, "Failed to count all sessions")
	assert.Equal(t, int64(4), count, "Should count all sessions including revoked and expired")
}

func TestCountActiveSessions(t *testing.T) {
	tdb := SetupTestDB(t)
	f := CreateStandardFixtures(t, tdb)

	expiresAt := pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	for i := 1; i <= 3; i++ {
		_, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
			HashedToken: "global-active-" + string(rune('a'+i)),
			UserID:      f.User1.ID,
			ExpiresAt:   expiresAt,
		})
		require.NoError(t, err, "Failed to create session")
	}

	revokedSession, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "global-revoked",
		UserID:      f.User1.ID,
		ExpiresAt:   expiresAt,
	})
	require.NoError(t, err, "Failed to create session")
	err = tdb.Queries.RevokeSession(tdb.Ctx, revokedSession.ID)
	require.NoError(t, err, "Failed to revoke session")

	expiredTime := pgtype.Timestamp{
		Time:  time.Now().Add(-1 * time.Hour),
		Valid: true,
	}
	_, err = tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: "global-expired",
		UserID:      f.User1.ID,
		ExpiresAt:   expiredTime,
	})
	require.NoError(t, err, "Failed to create expired session")

	count, err := tdb.Queries.CountActiveSessions(tdb.Ctx)
	require.NoError(t, err, "Failed to count active sessions")
	assert.Equal(t, int64(3), count, "Should only count active, non-expired sessions")
}
