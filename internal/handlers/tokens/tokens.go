package tokens

import (
	"context"
	"crypto/rsa"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/keys"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	Pool       *pgxpool.Pool
	Queries    *db.Queries
	privateKey *rsa.PrivateKey
}

func NewHandler(pool *pgxpool.Pool, queries *db.Queries) Handler {
	signingKey, err := queries.GetActiveSigningKey(context.Background())
	if err != nil {
		panic("could not load signing keys")
	}

	kek, err := keys.LoadKEK()
	if err != nil {
		panic("could not load kek")
	}

	privateKey, err := keys.GetSigningKey(signingKey.EncryptedPrivateKey, kek)
	if err != nil {
		panic("could not decrypt private key")
	}

	return Handler{Pool: pool, Queries: queries, privateKey: privateKey}
}
