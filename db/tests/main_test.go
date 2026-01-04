//go:build integration

package db_tests

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		panic("DATABASE_URL not set")
	}

	var err error
	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}

	// Optional but useful: verify connectivity once
	if err := testPool.Ping(ctx); err != nil {
		panic(err)
	}

	code := m.Run()

	testPool.Close()
	os.Exit(code)
}
