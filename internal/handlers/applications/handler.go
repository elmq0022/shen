package applications

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/handlers"
	"github.com/jackc/pgx/v5/pgconn"
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

func (h *Handler) CreateApplication(c echo.Context) error {
	var req CreateApplicationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(err.Error()))
	}

	// Normalize name to lowercase
	name := strings.ToLower(strings.TrimSpace(req.Name))
	if name == "" {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("application name is required"))
	}

	app, err := h.queries.CreateApplication(c.Request().Context(), name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return c.JSON(http.StatusConflict, handlers.NewErrorResponse("application already exists"))
		}
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to create application"))
	}

	return c.JSON(http.StatusCreated, app)
}

func (h *Handler) ListApplications(c echo.Context) error {
	cursor := c.QueryParam("cursor")
	limitStr := c.QueryParam("limit")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("invalid limit parameter"))
	}

	apps, err := h.queries.ListActiveApplications(c.Request().Context(), db.ListActiveApplicationsParams{
		Column1: cursor,
		Limit:   int32(limit),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to list applications"))
	}

	return c.JSON(http.StatusOK, apps)
}

func (h *Handler) DeleteApplication(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("no application name provided"))
	}

	app, err := h.queries.GetApplicationByName(c.Request().Context(), name)
	if err != nil {
		return c.JSON(http.StatusNotFound, handlers.NewErrorResponse("application does not exist"))
	}

	if err := h.queries.DeactivateApplication(c.Request().Context(), app.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse("failed to deactivate application"))
	}

	return c.NoContent(http.StatusNoContent)
}
