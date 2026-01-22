package routes

import (
	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/handlers/tokens"
	"github.com/elmq0022/shen/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func RegisterTokenRoutes(g *echo.Group, pool *pgxpool.Pool, queries *db.Queries, mw middleware.Middleware) {
	h := tokens.NewHandler(pool, queries)
	g.POST("/:name/:application", h.CreatePAT, mw.IsAuthenticated)
	// g.GET() // list tokens endpoint
}
