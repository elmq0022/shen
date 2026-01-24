package middleware_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/crypto"
	"github.com/elmq0022/shen/internal/middleware"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockQueries implements a mock for db.Queries
type mockQueries struct {
	getSessionByHashedTokenFunc func(ctx context.Context, hashedToken string) (db.ShenSession, error)
	getUserByIDFunc             func(ctx context.Context, id int32) (db.ShenUser, error)
	listRolesFunc               func(ctx context.Context) ([]db.ShenUserRole, error)
	getTokenByHashedTokenFunc   func(ctx context.Context, hashedToken string) (db.ShenToken, error)
}

func (m *mockQueries) GetSessionByHashedToken(ctx context.Context, hashedToken string) (db.ShenSession, error) {
	if m.getSessionByHashedTokenFunc != nil {
		return m.getSessionByHashedTokenFunc(ctx, hashedToken)
	}
	return db.ShenSession{}, errors.New("not implemented")
}

func (m *mockQueries) GetUserByID(ctx context.Context, id int32) (db.ShenUser, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, id)
	}
	return db.ShenUser{}, errors.New("not implemented")
}

func (m *mockQueries) ListRoles(ctx context.Context) ([]db.ShenUserRole, error) {
	if m.listRolesFunc != nil {
		return m.listRolesFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *mockQueries) GetTokenByHashedToken(ctx context.Context, hashedToken string) (db.ShenToken, error) {
	if m.getTokenByHashedTokenFunc != nil {
		return m.getTokenByHashedTokenFunc(ctx, hashedToken)
	}
	return db.ShenToken{}, errors.New("not implemented")
}

func createMockMiddleware(t *testing.T) (middleware.Middleware, *mockQueries) {
	mock := &mockQueries{
		listRolesFunc: func(ctx context.Context) ([]db.ShenUserRole, error) {
			return []db.ShenUserRole{
				{ID: 1, Name: "admin"},
				{ID: 2, Name: "user"},
				{ID: 3, Name: "service"},
			}, nil
		},
	}

	m, err := middleware.NewMiddleware(context.Background(), mock)
	require.NoError(t, err)
	return m, mock
}

func setupEcho() (*echo.Echo, *httptest.ResponseRecorder) {
	e := echo.New()
	rec := httptest.NewRecorder()
	return e, rec
}

func createRequest(method, path, authHeader string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	return req
}

func TestNewMiddleware_Success(t *testing.T) {
	mock := &mockQueries{
		listRolesFunc: func(ctx context.Context) ([]db.ShenUserRole, error) {
			return []db.ShenUserRole{
				{ID: 1, Name: "admin"},
				{ID: 2, Name: "user"},
			}, nil
		},
	}

	m, err := middleware.NewMiddleware(context.Background(), mock)
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestNewMiddleware_FailsOnListRolesError(t *testing.T) {
	mock := &mockQueries{
		listRolesFunc: func(ctx context.Context) ([]db.ShenUserRole, error) {
			return nil, errors.New("database error")
		},
	}

	_, err := middleware.NewMiddleware(context.Background(), mock)
	require.Error(t, err)
}

func TestIsAdmin_MissingAuthorizationHeader(t *testing.T) {
	e, rec := setupEcho()
	req := createRequest(http.MethodGet, "/admin", "")
	c := e.NewContext(req, rec)

	m, _ := createMockMiddleware(t)

	nextCalled := false
	handler := m.IsAdmin(func(c echo.Context) error {
		nextCalled = true
		return nil
	})

	err := handler(c)
	require.Error(t, err)
	assert.False(t, nextCalled)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	assert.Equal(t, "Authorization header required", httpErr.Message)
}

func TestIsAdmin_InvalidAuthorizationFormat(t *testing.T) {
	testCases := []struct {
		name   string
		header string
	}{
		{
			name:   "no Bearer prefix",
			header: "token123",
		},
		{
			name:   "wrong prefix",
			header: "Basic token123",
		},
		{
			name:   "lowercase bearer",
			header: "bearer token123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e, rec := setupEcho()
			req := createRequest(http.MethodGet, "/admin", tc.header)
			c := e.NewContext(req, rec)

			m, _ := createMockMiddleware(t)

			nextCalled := false
			handler := m.IsAdmin(func(c echo.Context) error {
				nextCalled = true
				return nil
			})

			err := handler(c)
			require.Error(t, err)
			assert.False(t, nextCalled)

			httpErr, ok := err.(*echo.HTTPError)
			require.True(t, ok)
			assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
			assert.Equal(t, "Invalid authorization header format", httpErr.Message)
		})
	}
}

func TestIsAdmin_SessionNotFound(t *testing.T) {
	e, rec := setupEcho()
	req := createRequest(http.MethodGet, "/admin", "Bearer validtoken123")
	c := e.NewContext(req, rec)

	m, mock := createMockMiddleware(t)
	mock.getSessionByHashedTokenFunc = func(ctx context.Context, hashedToken string) (db.ShenSession, error) {
		return db.ShenSession{}, errors.New("session not found")
	}

	nextCalled := false
	handler := m.IsAdmin(func(c echo.Context) error {
		nextCalled = true
		return nil
	})

	err := handler(c)
	require.Error(t, err)
	assert.False(t, nextCalled)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	assert.Equal(t, "Invalid or expired session", httpErr.Message)
}

func TestIsAdmin_RevokedSession(t *testing.T) {
	e, rec := setupEcho()
	token := "validtoken123"
	hashedToken := crypto.HashToken(token)
	req := createRequest(http.MethodGet, "/admin", fmt.Sprintf("Bearer %s", token))
	c := e.NewContext(req, rec)

	m, mock := createMockMiddleware(t)
	mock.getSessionByHashedTokenFunc = func(ctx context.Context, ht string) (db.ShenSession, error) {
		assert.Equal(t, hashedToken, ht)
		return db.ShenSession{
			ID:          1,
			HashedToken: hashedToken,
			UserID:      100,
			Revoked:     true,
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(24 * time.Hour),
				Valid: true,
			},
		}, nil
	}

	nextCalled := false
	handler := m.IsAdmin(func(c echo.Context) error {
		nextCalled = true
		return nil
	})

	err := handler(c)
	require.Error(t, err)
	assert.False(t, nextCalled)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	assert.Equal(t, "Session has been revoked", httpErr.Message)
}

func TestIsAdmin_ExpiredSession(t *testing.T) {
	e, rec := setupEcho()
	token := "validtoken123"
	hashedToken := crypto.HashToken(token)
	req := createRequest(http.MethodGet, "/admin", fmt.Sprintf("Bearer %s", token))
	c := e.NewContext(req, rec)

	m, mock := createMockMiddleware(t)
	mock.getSessionByHashedTokenFunc = func(ctx context.Context, ht string) (db.ShenSession, error) {
		assert.Equal(t, hashedToken, ht)
		return db.ShenSession{
			ID:          1,
			HashedToken: hashedToken,
			UserID:      100,
			Revoked:     false,
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
				Valid: true,
			},
		}, nil
	}

	nextCalled := false
	handler := m.IsAdmin(func(c echo.Context) error {
		nextCalled = true
		return nil
	})

	err := handler(c)
	require.Error(t, err)
	assert.False(t, nextCalled)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	assert.Equal(t, "Session has expired", httpErr.Message)
}

func TestIsAdmin_UserNotFound(t *testing.T) {
	e, rec := setupEcho()
	token := "validtoken123"
	hashedToken := crypto.HashToken(token)
	req := createRequest(http.MethodGet, "/admin", fmt.Sprintf("Bearer %s", token))
	c := e.NewContext(req, rec)

	m, mock := createMockMiddleware(t)
	mock.getSessionByHashedTokenFunc = func(ctx context.Context, ht string) (db.ShenSession, error) {
		assert.Equal(t, hashedToken, ht)
		return db.ShenSession{
			ID:          1,
			HashedToken: hashedToken,
			UserID:      100,
			Revoked:     false,
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(24 * time.Hour),
				Valid: true,
			},
		}, nil
	}
	mock.getUserByIDFunc = func(ctx context.Context, id int32) (db.ShenUser, error) {
		assert.Equal(t, int32(100), id)
		return db.ShenUser{}, errors.New("user not found")
	}

	nextCalled := false
	handler := m.IsAdmin(func(c echo.Context) error {
		nextCalled = true
		return nil
	})

	err := handler(c)
	require.Error(t, err)
	assert.False(t, nextCalled)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	assert.Equal(t, "User not found", httpErr.Message)
}

func TestIsAdmin_InactiveUser(t *testing.T) {
	e, rec := setupEcho()
	token := "validtoken123"
	hashedToken := crypto.HashToken(token)
	req := createRequest(http.MethodGet, "/admin", fmt.Sprintf("Bearer %s", token))
	c := e.NewContext(req, rec)

	m, mock := createMockMiddleware(t)
	mock.getSessionByHashedTokenFunc = func(ctx context.Context, ht string) (db.ShenSession, error) {
		assert.Equal(t, hashedToken, ht)
		return db.ShenSession{
			ID:          1,
			HashedToken: hashedToken,
			UserID:      100,
			Revoked:     false,
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(24 * time.Hour),
				Valid: true,
			},
		}, nil
	}
	mock.getUserByIDFunc = func(ctx context.Context, id int32) (db.ShenUser, error) {
		assert.Equal(t, int32(100), id)
		return db.ShenUser{
			ID:       100,
			Username: "inactiveuser",
			Active:   false, // Inactive user
			Role:     1,
		}, nil
	}

	nextCalled := false
	handler := m.IsAdmin(func(c echo.Context) error {
		nextCalled = true
		return nil
	})

	err := handler(c)
	require.Error(t, err)
	assert.False(t, nextCalled)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusForbidden, httpErr.Code)
	assert.Equal(t, "Account is not active", httpErr.Message)
}

func TestIsAdmin_NonAdminUser(t *testing.T) {
	e, rec := setupEcho()
	token := "validtoken123"
	hashedToken := crypto.HashToken(token)
	req := createRequest(http.MethodGet, "/admin", fmt.Sprintf("Bearer %s", token))
	c := e.NewContext(req, rec)

	m, mock := createMockMiddleware(t)
	mock.getSessionByHashedTokenFunc = func(ctx context.Context, ht string) (db.ShenSession, error) {
		assert.Equal(t, hashedToken, ht)
		return db.ShenSession{
			ID:          1,
			HashedToken: hashedToken,
			UserID:      100,
			Revoked:     false,
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(24 * time.Hour),
				Valid: true,
			},
		}, nil
	}
	mock.getUserByIDFunc = func(ctx context.Context, id int32) (db.ShenUser, error) {
		assert.Equal(t, int32(100), id)
		return db.ShenUser{
			ID:       100,
			Username: "regularuser",
			Active:   true,
			Role:     2, // Non-admin role (user role)
		}, nil
	}

	nextCalled := false
	handler := m.IsAdmin(func(c echo.Context) error {
		nextCalled = true
		return nil
	})

	err := handler(c)
	require.Error(t, err)
	assert.False(t, nextCalled)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusForbidden, httpErr.Code)
	assert.Equal(t, "Insufficient permissions", httpErr.Message)
}

func TestIsAdmin_Success(t *testing.T) {
	e, rec := setupEcho()
	token := "validtoken123"
	hashedToken := crypto.HashToken(token)
	req := createRequest(http.MethodGet, "/admin", fmt.Sprintf("Bearer %s", token))
	c := e.NewContext(req, rec)

	expectedUser := db.ShenUser{
		ID:       100,
		Username: "adminuser",
		Active:   true,
		Role:     1,
	}

	m, mock := createMockMiddleware(t)
	mock.getSessionByHashedTokenFunc = func(ctx context.Context, ht string) (db.ShenSession, error) {
		assert.Equal(t, hashedToken, ht)
		return db.ShenSession{
			ID:          1,
			HashedToken: hashedToken,
			UserID:      100,
			Revoked:     false,
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(24 * time.Hour),
				Valid: true,
			},
		}, nil
	}
	mock.getUserByIDFunc = func(ctx context.Context, id int32) (db.ShenUser, error) {
		assert.Equal(t, int32(100), id)
		return expectedUser, nil
	}

	nextCalled := false
	handler := m.IsAdmin(func(c echo.Context) error {
		nextCalled = true
		// Verify user is set in context
		user := c.Get("user")
		require.NotNil(t, user)
		userFromContext, ok := user.(db.ShenUser)
		require.True(t, ok)
		assert.Equal(t, expectedUser, userFromContext)
		return nil
	})

	err := handler(c)
	require.NoError(t, err)
	assert.True(t, nextCalled)
}

func TestIsAdmin_TokenExtractionCorrect(t *testing.T) {
	// This test verifies the bug fix for token extraction
	e, rec := setupEcho()
	token := "mytoken123"
	expectedHash := crypto.HashToken(token)
	req := createRequest(http.MethodGet, "/admin", fmt.Sprintf("Bearer %s", token))
	c := e.NewContext(req, rec)

	hashReceived := ""
	m, mock := createMockMiddleware(t)
	mock.getSessionByHashedTokenFunc = func(ctx context.Context, ht string) (db.ShenSession, error) {
		hashReceived = ht
		return db.ShenSession{
			ID:          1,
			HashedToken: ht,
			UserID:      100,
			Revoked:     false,
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(24 * time.Hour),
				Valid: true,
			},
		}, nil
	}
	mock.getUserByIDFunc = func(ctx context.Context, id int32) (db.ShenUser, error) {
		return db.ShenUser{
			ID:       100,
			Username: "testuser",
			Active:   true,
			Role:     1,
		}, nil
	}

	handler := m.IsAdmin(func(c echo.Context) error {
		return nil
	})

	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, expectedHash, hashReceived, "Token should be correctly extracted and hashed")
}

func TestIsAdmin_SessionWithoutExpiresAt(t *testing.T) {
	// Test session without expiration time (Valid = false)
	e, rec := setupEcho()
	token := "validtoken123"
	hashedToken := crypto.HashToken(token)
	req := createRequest(http.MethodGet, "/admin", fmt.Sprintf("Bearer %s", token))
	c := e.NewContext(req, rec)

	m, mock := createMockMiddleware(t)
	mock.getSessionByHashedTokenFunc = func(ctx context.Context, ht string) (db.ShenSession, error) {
		return db.ShenSession{
			ID:          1,
			HashedToken: hashedToken,
			UserID:      100,
			Revoked:     false,
			ExpiresAt: pgtype.Timestamptz{
				Valid: false, // No expiration time set
			},
		}, nil
	}
	mock.getUserByIDFunc = func(ctx context.Context, id int32) (db.ShenUser, error) {
		return db.ShenUser{
			ID:       100,
			Username: "adminuser",
			Active:   true,
			Role:     1,
		}, nil
	}

	nextCalled := false
	handler := m.IsAdmin(func(c echo.Context) error {
		nextCalled = true
		return nil
	})

	err := handler(c)
	require.NoError(t, err)
	assert.True(t, nextCalled, "Session without expiration should be allowed")
}

func TestRequireRole_SingleRole(t *testing.T) {
	e, rec := setupEcho()
	token := "validtoken123"
	hashedToken := crypto.HashToken(token)
	req := createRequest(http.MethodGet, "/api", fmt.Sprintf("Bearer %s", token))
	c := e.NewContext(req, rec)

	m, mock := createMockMiddleware(t)
	mock.getSessionByHashedTokenFunc = func(ctx context.Context, ht string) (db.ShenSession, error) {
		return db.ShenSession{
			ID:          1,
			HashedToken: hashedToken,
			UserID:      100,
			Revoked:     false,
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(24 * time.Hour),
				Valid: true,
			},
		}, nil
	}
	mock.getUserByIDFunc = func(ctx context.Context, id int32) (db.ShenUser, error) {
		return db.ShenUser{
			ID:       100,
			Username: "regularuser",
			Active:   true,
			Role:     2, // user role
		}, nil
	}

	nextCalled := false
	handler := m.RequireRole("user")(func(c echo.Context) error {
		nextCalled = true
		return nil
	})

	err := handler(c)
	require.NoError(t, err)
	assert.True(t, nextCalled)
}

func TestRequireRole_MultipleRoles(t *testing.T) {
	e, rec := setupEcho()
	token := "validtoken123"
	hashedToken := crypto.HashToken(token)
	req := createRequest(http.MethodGet, "/api", fmt.Sprintf("Bearer %s", token))
	c := e.NewContext(req, rec)

	m, mock := createMockMiddleware(t)
	mock.getSessionByHashedTokenFunc = func(ctx context.Context, ht string) (db.ShenSession, error) {
		return db.ShenSession{
			ID:          1,
			HashedToken: hashedToken,
			UserID:      100,
			Revoked:     false,
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(24 * time.Hour),
				Valid: true,
			},
		}, nil
	}
	mock.getUserByIDFunc = func(ctx context.Context, id int32) (db.ShenUser, error) {
		return db.ShenUser{
			ID:       100,
			Username: "serviceuser",
			Active:   true,
			Role:     3, // service role
		}, nil
	}

	nextCalled := false
	handler := m.RequireRole("admin", "service")(func(c echo.Context) error {
		nextCalled = true
		return nil
	})

	err := handler(c)
	require.NoError(t, err)
	assert.True(t, nextCalled)
}

func TestRequireRole_NoMatchingRole(t *testing.T) {
	e, rec := setupEcho()
	token := "validtoken123"
	hashedToken := crypto.HashToken(token)
	req := createRequest(http.MethodGet, "/api", fmt.Sprintf("Bearer %s", token))
	c := e.NewContext(req, rec)

	m, mock := createMockMiddleware(t)
	mock.getSessionByHashedTokenFunc = func(ctx context.Context, ht string) (db.ShenSession, error) {
		return db.ShenSession{
			ID:          1,
			HashedToken: hashedToken,
			UserID:      100,
			Revoked:     false,
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(24 * time.Hour),
				Valid: true,
			},
		}, nil
	}
	mock.getUserByIDFunc = func(ctx context.Context, id int32) (db.ShenUser, error) {
		return db.ShenUser{
			ID:       100,
			Username: "regularuser",
			Active:   true,
			Role:     2, // user role
		}, nil
	}

	nextCalled := false
	handler := m.RequireRole("admin", "service")(func(c echo.Context) error {
		nextCalled = true
		return nil
	})

	err := handler(c)
	require.Error(t, err)
	assert.False(t, nextCalled)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusForbidden, httpErr.Code)
	assert.Equal(t, "Insufficient permissions", httpErr.Message)
}

func TestRequireRole_InvalidRoleInCache(t *testing.T) {
	e, rec := setupEcho()
	token := "validtoken123"
	hashedToken := crypto.HashToken(token)
	req := createRequest(http.MethodGet, "/api", fmt.Sprintf("Bearer %s", token))
	c := e.NewContext(req, rec)

	m, mock := createMockMiddleware(t)
	mock.getSessionByHashedTokenFunc = func(ctx context.Context, ht string) (db.ShenSession, error) {
		return db.ShenSession{
			ID:          1,
			HashedToken: hashedToken,
			UserID:      100,
			Revoked:     false,
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(24 * time.Hour),
				Valid: true,
			},
		}, nil
	}
	mock.getUserByIDFunc = func(ctx context.Context, id int32) (db.ShenUser, error) {
		return db.ShenUser{
			ID:       100,
			Username: "regularuser",
			Active:   true,
			Role:     2,
		}, nil
	}

	nextCalled := false
	handler := m.RequireRole("nonexistent")(func(c echo.Context) error {
		nextCalled = true
		return nil
	})

	err := handler(c)
	require.Error(t, err)
	assert.False(t, nextCalled)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	assert.Equal(t, "Invalid role configuration", httpErr.Message)
}
