package routes

import (
	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/handlers/groups"
	"github.com/elmq0022/shen/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func RegisterGroupRoutes(g *echo.Group, pool *pgxpool.Pool, queries *db.Queries, mw middleware.Middleware) {
	h := groups.NewHandler(pool, queries)
	g.GET("", h.ListGroups, mw.IsAdmin)
	g.POST("", h.CreateGroup, mw.IsAdmin)
	g.DELETE(":name", h.DeleteGroup, mw.IsAdmin)
	g.GET(":name/members", h.ListGroupMembers, mw.IsAdmin)
	g.POST(":name/members", h.AddGroupMembers, mw.IsAdmin)
	g.DELETE(":name/members", h.RemoveGroupMembers, mw.IsAdmin)
}
