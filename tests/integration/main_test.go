//go:build cli_integration

package integration

import (
	"context"
	"os"
	"testing"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	shen    string
	shenctl string
	queries *db.Queries // use only for verifying db state during tests; server handles writes
)

func TestMain(m *testing.M) {
	shen = compileBinary("../../cmd/shen", "shen")
	shenctl = compileBinary("../../cli/shenctl", "shenctl")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		panic("DATABASE_URL environment variable is not set")
	}

	testPool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic("failed to create database connection pool: " + err.Error())
	}

	queries = db.New(testPool)

	exitCode := m.Run()

	testPool.Close()
	os.Remove(shen)
	os.Remove(shenctl)
	os.Exit(exitCode)
}
