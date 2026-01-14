package main

import (
	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/handlers/auth"
	"github.com/labstack/echo/v4"
)

func NewAuthGroup(g *echo.Group, queries *db.Queries) {
	authHandler := auth.NewHandler(queries)
	g.POST("login", authHandler.Login)
	g.POST("logout", authHandler.Logout)
}
