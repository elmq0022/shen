package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/crypto"
	"github.com/labstack/echo/v4"
)

type Querier interface {
	GetSessionByHashedToken(ctx context.Context, hashedToken string) (db.ShenSession, error)
	GetUserByID(ctx context.Context, id int32) (db.ShenUser, error)
	ListRoles(ctx context.Context) ([]db.ShenUserRole, error)
}

type Middleware struct {
	queries   Querier
	roleCache map[string]int32
}

func NewMiddleware(ctx context.Context, queries Querier) (Middleware, error) {
	roles, err := queries.ListRoles(ctx)
	if err != nil {
		return Middleware{}, err
	}

	roleCache := make(map[string]int32)
	for _, role := range roles {
		roleCache[role.Name] = role.ID
	}

	return Middleware{
		queries:   queries,
		roleCache: roleCache,
	}, nil
}

// RequireRole returns a middleware that checks if the authenticated user has one of the specified roles
func (m *Middleware) RequireRole(roleNames ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Authorization header required")
			}

			prefix := "Bearer "
			if !strings.HasPrefix(authHeader, prefix) {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authorization header format")
			}

			sessionToken := strings.TrimPrefix(authHeader, prefix)
			hashedToken := crypto.HashToken(sessionToken)
			session, err := m.queries.GetSessionByHashedToken(c.Request().Context(), hashedToken)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired session")
			}

			if session.Revoked {
				return echo.NewHTTPError(http.StatusUnauthorized, "Session has been revoked")
			}

			if session.ExpiresAt.Valid && session.ExpiresAt.Time.Before(time.Now()) {
				return echo.NewHTTPError(http.StatusUnauthorized, "Session has expired")
			}

			user, err := m.queries.GetUserByID(c.Request().Context(), session.UserID)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
			}

			if !user.Active {
				return echo.NewHTTPError(http.StatusForbidden, "Account is not active")
			}

			for _, roleName := range roleNames {
				roleID, exists := m.roleCache[roleName]
				if !exists {
					return echo.NewHTTPError(http.StatusInternalServerError, "Invalid role configuration")
				}
				if user.Role == roleID {
					c.Set("user", user)
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
		}
	}
}

// IsAdmin is a convenience method that checks if the user has the admin role
func (m *Middleware) IsAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return m.RequireRole("admin")(next)
}
