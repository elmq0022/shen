package users

import (
	"errors"
	"net/http"
	"strconv"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/crypto"
	"github.com/elmq0022/shen/internal/handlers"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	queries *db.Queries
}

func NewHandler(queries *db.Queries) Handler {
	return Handler{queries: queries}
}

func (h *Handler) ListUsers(c echo.Context) error {
	cursor := c.QueryParam("cursor")
	limitStr := c.QueryParam("limit")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("invalid limit parameter"))
	}

	users, err := h.queries.ListUsers(c.Request().Context(), db.ListUsersParams{
		Column1: cursor,
		Limit:   int32(limit),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to list users"))
	}

	return c.JSON(http.StatusOK, users)
}

func (h *Handler) CreateUser(c echo.Context) error {
	var cur CreateUserRequest
	if err := c.Bind(&cur); err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(err.Error()))
	}

	hashedPassword, err := crypto.HashedPassword(cur.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to hash password"))
	}

	role, err := h.queries.GetRoleByName(c.Request().Context(), cur.Role)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("invalid role"))
	}

	user, err := h.queries.CreateUser(c.Request().Context(), db.CreateUserParams{
		Username: cur.UserName,
		HashedPassword: pgtype.Text{
			String: hashedPassword,
			Valid:  true,
		},
		Role: role.ID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return c.JSON(http.StatusConflict, handlers.NewErrorResponse("username already exists"))
		}
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to create user"))
	}

	return c.JSON(http.StatusCreated, user)
}

func (h *Handler) DeleteUser(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("no username provided"))
	}

	user, err := h.queries.GetUserByUsername(c.Request().Context(), username)
	if err != nil {
		return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("user does not exist"))
	}

	if err := h.queries.DeactivateUser(c.Request().Context(), user.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to deactivate user"))
	}

	return c.NoContent(http.StatusNoContent)
}
