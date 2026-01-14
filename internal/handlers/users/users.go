package users

import db "github.com/elmq0022/shen/db/sqlc"

type Handler struct {
	queries *db.Queries
}

func NewHandler(queries *db.Queries) Handler {
	return Handler{queries: queries}
}
