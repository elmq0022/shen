package auth

import (
	"net/http"
	"time"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/crypto"
	"github.com/elmq0022/shen/internal/handlers"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	queries *db.Queries
}

func NewHandler(queries *db.Queries) Handler {
	return Handler{
		queries: queries,
	}
}

func (h *Handler) Login(c echo.Context) error {

	var lr LoginRequest
	if err := c.Bind(&lr); err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(err.Error()))
	}

	user, err := h.queries.GetUserByUsername(c.Request().Context(), lr.Username)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, handlers.NewErrorResponse("Invalid credentials"))
	}

	if !user.Active {
		return c.JSON(http.StatusForbidden, handlers.NewErrorResponse("Forbidden"))
	}

	role, err := h.queries.GetRoleByName(c.Request().Context(), "service") // there's probably a better way
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("Internal server error"))
	}
	if user.Role == role.ID {
		return c.JSON(http.StatusForbidden, handlers.NewErrorResponse("Forbidden"))
	}

	authorized, err := crypto.CheckPassword(lr.Password, user.HashedPassword.String)
	if err != nil || !authorized {
		return c.JSON(http.StatusUnauthorized, handlers.NewErrorResponse("Invalid credentials"))
	}

	token, err := crypto.GenerateSessionToken()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("Failed to generate token"))
	}

	hashedToken := crypto.HashToken(token)

	// TODO: configure token expiry
	_, err = h.queries.CreateSession(c.Request().Context(), db.CreateSessionParams{
		HashedToken: hashedToken,
		UserID:      user.ID,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(30 * 24 * time.Hour),
			Valid: true,
		},
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("Failed to create session"))
	}

	return c.JSON(http.StatusOK, LoginResponse{SessionToken: token})
}

func (h *Handler) Logout(c echo.Context) error {

	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, handlers.NewErrorResponse("Authorization header required"))
	}

	const prefix = "Bearer "
	if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
		return c.JSON(http.StatusUnauthorized, handlers.NewErrorResponse("Invalid authorization header format"))
	}

	sessionToken := authHeader[len(prefix):]
	hashedToken := crypto.HashToken(sessionToken)

	// Get the session to find the user ID
	session, err := h.queries.GetSessionByHashedToken(c.Request().Context(), hashedToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, handlers.NewErrorResponse("Invalid session"))
	}

	// Revoke all sessions for the user
	err = h.queries.RevokeAllUserSessions(c.Request().Context(), session.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("Failed to revoke sessions"))
	}

	return c.NoContent(http.StatusNoContent)
}
