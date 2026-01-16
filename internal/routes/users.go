package routes

import (
	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/handlers/users"
	"github.com/labstack/echo/v4"
)

func RegisterUserRoutes(g *echo.Group, queries *db.Queries) {
	h := users.NewHandler(queries)
	g.GET("", h.ListUsers)
	g.POST("", h.CreateUser)
	g.DELETE(":username", h.DeleteUser)
}
