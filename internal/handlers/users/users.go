package users

import (
	"net/http"
	"strconv"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/handlers"
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
