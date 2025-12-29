//go:build integration

package db_tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TestDB struct {
	Pool    *pgxpool.Pool
	Queries *db.Queries
	Ctx     context.Context
}

func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL environment variable is not set")
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("Failed to create connection pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	queries := db.New(pool)

	testDB := &TestDB{
		Pool:    pool,
		Queries: queries,
		Ctx:     ctx,
	}

	// Register cleanup function to truncate tables and close pool
	t.Cleanup(func() {
		tables := []string{
			"shen_token",
			"shen_session",
			"shen_group_application_permission",
			"shen_user_group_manager",
			"shen_user_group_member",
			"shen_user",
			"shen_group",
			"shen_application",
			// skip shen_permission, shen_user_role as they have data seeded in the migrations
		}

		for _, table := range tables {
			query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
			if _, err := testDB.Pool.Exec(testDB.Ctx, query); err != nil {
				t.Logf("warning: failed to truncate %s: %v", table, err)
			}
		}

		testDB.Pool.Close()
	})

	return testDB
}
