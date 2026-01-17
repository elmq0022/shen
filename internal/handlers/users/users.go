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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewHandler(pool *pgxpool.Pool, queries *db.Queries) Handler {
	return Handler{pool: pool, queries: queries}
}

func (h *Handler) ListUsers(c echo.Context) error {
	cursor := c.QueryParam("cursor")
	limitStr := c.QueryParam("limit")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("invalid limit parameter"))
	}

	users, err := h.queries.ListActiveUsers(c.Request().Context(), db.ListActiveUsersParams{
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

func (h *Handler) UpdateUser(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("no username provided"))
	}

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(err.Error()))
	}

	if req.Role == nil && req.Password == nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("no fields to update"))
	}

	ctx := c.Request().Context()
	currentUser := c.Get("user").(db.ShenUser)

	targetUser, err := h.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("user not found"))
	}

	currentUserRole, err := h.queries.GetRoleByID(ctx, currentUser.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to get current user role"))
	}
	isAdmin := currentUserRole.Name == "admin"
	isSelf := currentUser.ID == targetUser.ID

	// Validate permissions before starting transaction
	if req.Role != nil && !isAdmin {
		return c.JSON(http.StatusForbidden, handlers.NewErrorResponse("only admins can change user roles"))
	}
	if req.Password != nil && !isAdmin && !isSelf {
		return c.JSON(http.StatusForbidden, handlers.NewErrorResponse("can only change your own password"))
	}

	// Validate role exists before starting transaction
	var newRoleID int32
	if req.Role != nil {
		newRole, err := h.queries.GetRoleByName(ctx, *req.Role)
		if err != nil {
			return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("invalid role"))
		}
		newRoleID = newRole.ID
	}

	// Hash password before starting transaction
	var hashedPassword string
	if req.Password != nil {
		hashedPassword, err = crypto.HashedPassword(*req.Password)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to hash password"))
		}
	}

	// Start transaction for the updates
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to start transaction"))
	}
	defer tx.Rollback(ctx)

	qtx := h.queries.WithTx(tx)

	if req.Role != nil {
		if err := qtx.UpdateUserRole(ctx, db.UpdateUserRoleParams{
			ID:   targetUser.ID,
			Role: newRoleID,
		}); err != nil {
			return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to update user role"))
		}
	}

	if req.Password != nil {
		if err := qtx.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
			ID:             targetUser.ID,
			HashedPassword: pgtype.Text{String: hashedPassword, Valid: true},
		}); err != nil {
			return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to update password"))
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to commit transaction"))
	}

	return c.NoContent(http.StatusNoContent)
}
