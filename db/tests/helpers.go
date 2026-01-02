//go:build integration

package db_tests

import (
	"fmt"
	"slices"
	"testing"
	"time"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

// addUsersToGroup adds multiple users as members to a group.
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

// addManagersToGroup adds multiple users as managers to a group.
func addManagersToGroup(t *testing.T, tdb *TestDB, users []db.ShenUser, group db.ShenGroup) {
	t.Helper()
	for _, user := range users {
		_, err := tdb.Queries.AddManagerToGroup(tdb.Ctx, db.AddManagerToGroupParams{
			UserID:  user.ID,
			GroupID: group.ID,
		})
		require.NoError(
			t, err,
			fmt.Sprintf("Failed to add manager: %s to group: %s", user.Username, group.Name),
		)
	}
}

// sortUsersByUsername sorts a slice of users by username in ascending order.
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

// sortGroupsByName sorts a slice of groups by name in ascending order.
func sortGroupsByName(groups []db.ShenGroup) {
	slices.SortFunc(groups, func(a, b db.ShenGroup) int {
		if a.Name < b.Name {
			return -1
		} else if a.Name == b.Name {
			return 0
		}
		return 1
	})
}

// setGroupApplicationPermissions sets permissions for multiple group-application pairs.
// Takes a slice of SetGroupApplicationPermissionParams.
func setGroupApplicationPermissions(t *testing.T, tdb *TestDB, permissions []db.SetGroupApplicationPermissionParams) {
	t.Helper()
	for _, perm := range permissions {
		_, err := tdb.Queries.SetGroupApplicationPermission(tdb.Ctx, perm)
		require.NoError(t, err, "Failed to set group application permission")
	}
}

// CreateTestToken creates a single token with the given parameters.
func CreateTestToken(t *testing.T, tdb *TestDB, name, hashedToken string, userID, appID int32, expiresAt pgtype.Timestamp) db.ShenToken {
	t.Helper()
	created, err := tdb.Queries.CreateToken(tdb.Ctx, db.CreateTokenParams{
		Name:          name,
		HashedToken:   hashedToken,
		UserID:        userID,
		ApplicationID: appID,
		ExpiresAt:     expiresAt,
	})
	require.NoError(t, err, fmt.Sprintf("Failed to create token: %s", name))
	return created
}

// CreateTestTokens creates multiple tokens with standardized naming.
// Creates tokens named "<prefix>-1", "<prefix>-2", etc.
func CreateTestTokens(t *testing.T, tdb *TestDB, prefix string, count int, userID, appID int32, expiresAt pgtype.Timestamp) []db.ShenToken {
	t.Helper()
	tokens := make([]db.ShenToken, count)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s-%d", prefix, i+1)
		hashedToken := fmt.Sprintf("%s-hash-%d", prefix, i+1)
		tokens[i] = CreateTestToken(t, tdb, name, hashedToken, userID, appID, expiresAt)
	}
	return tokens
}

// GetActiveExpiresAt returns a pgtype.Timestamp set 24 hours in the future.
func GetActiveExpiresAt() pgtype.Timestamp {
	return pgtype.Timestamp{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}
}

// GetExpiredExpiresAt returns a pgtype.Timestamp set 1 hour in the past.
func GetExpiredExpiresAt() pgtype.Timestamp {
	return pgtype.Timestamp{
		Time:  time.Now().Add(-1 * time.Hour),
		Valid: true,
	}
}

// CreateTestSession creates a single session with the given parameters.
func CreateTestSession(t *testing.T, tdb *TestDB, hashedToken string, userID int32, expiresAt pgtype.Timestamp) db.ShenSession {
	t.Helper()
	created, err := tdb.Queries.CreateSession(tdb.Ctx, db.CreateSessionParams{
		HashedToken: hashedToken,
		UserID:      userID,
		ExpiresAt:   expiresAt,
	})
	require.NoError(t, err, fmt.Sprintf("Failed to create session: %s", hashedToken))
	return created
}

// CreateTestSessions creates multiple sessions with standardized naming.
// Creates sessions with hashed tokens named "<prefix>-1", "<prefix>-2", etc.
func CreateTestSessions(t *testing.T, tdb *TestDB, prefix string, count int, userID int32, expiresAt pgtype.Timestamp) []db.ShenSession {
	t.Helper()
	sessions := make([]db.ShenSession, count)
	for i := 0; i < count; i++ {
		hashedToken := fmt.Sprintf("%s-%d", prefix, i+1)
		sessions[i] = CreateTestSession(t, tdb, hashedToken, userID, expiresAt)
	}
	return sessions
}
