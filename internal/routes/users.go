package routes

import (
	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/handlers/users"
	"github.com/elmq0022/shen/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func RegisterUserRoutes(g *echo.Group, pool *pgxpool.Pool, queries *db.Queries, mw middleware.Middleware) {
	h := users.NewHandler(pool, queries)
	g.GET("", h.ListUsers, mw.IsAdmin)
	g.POST("", h.CreateUser, mw.IsAdmin)
	g.DELETE(":username", h.DeleteUser, mw.IsAdmin)
	g.PATCH(":username", h.UpdateUser, mw.IsAuthenticated)
}
