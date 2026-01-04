//go:build integration

package db_tests

import (
	"context"
	"testing"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type TestDB struct {
	Pool    *pgxpool.Pool
	Queries *db.Queries
	Ctx     context.Context
}

func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	ctx := context.Background()

	if testPool == nil {
		t.Fatal("testPool is nil; did TestMain run?")
	}

	tx, err := testPool.Begin(ctx)
	require.NoError(t, err)

	queries := db.New(tx)

	t.Cleanup(func() {
		_ = tx.Rollback(ctx)
	})

	return &TestDB{
		Pool:    testPool, // optional now
		Queries: queries,
		Ctx:     ctx,
	}
}
