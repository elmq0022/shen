package tokens

import (
	"net/http"
	"strconv"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/handlers"
	"github.com/labstack/echo/v4"
)

func (h *Handler) ListPATs(c echo.Context) error {
	requestor := c.Get("user").(db.ShenUser)

	admin, err := h.Queries.GetRoleByName(c.Request().Context(), "admin")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("could not resolve roles"))
	}

	var user db.ShenUser = requestor
	if queryUser := c.QueryParam("user"); queryUser != "" && queryUser != requestor.Username {
		if requestor.Role != admin.ID {
			return c.JSON(http.StatusUnauthorized, handlers.NewErrorResponse("you must be an administrator to view another user's PATs"))
		}
		user, err = h.Queries.GetUserByUsername(c.Request().Context(), queryUser)
		if err != nil {
			return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("user not found"))
		}
	}

	var cursorID int32
	if cursor := c.QueryParam("cursor"); cursor != "" {
		parsed, err := strconv.ParseInt(cursor, 10, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("invalid cursor parameter"))
		}
		cursorID = int32(parsed)
	}

	var limit int32 = 10
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		parsed, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("invalid limit parameter"))
		}
		limit = int32(parsed)
	}

	tokens, err := h.Queries.ListTokensByUser(c.Request().Context(), db.ListTokensByUserParams{
		UserID:   user.ID,
		Limit:    limit,
		CursorID: cursorID,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to list tokens"))
	}

	return c.JSON(http.StatusOK, tokens)
}
