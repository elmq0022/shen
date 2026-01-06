package main

import (
	"context"
	"log"
	"os"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/auth"
	"github.com/elmq0022/shen/internal/bootstrap"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	// Initialize database connection
	pool, queries := initDatabase(ctx)
	defer pool.Close()

	// Run bootstrap functions on startup
	runBootstrap(ctx, queries)

	log.Println("Shen bootstrap completed successfully")

	// TODO: Add server startup code here
	// Pass queries to your handlers/services as needed
	// Example: handler := NewUserHandler(queries)
}

func initDatabase(ctx context.Context) (*pgxpool.Pool, *db.Queries) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	queries := db.New(pool)
	log.Println("Database connection established")

	return pool, queries
}

func runBootstrap(ctx context.Context, queries *db.Queries) {
	log.Println("Running bootstrap: generating initial JWT keys...")
	bootstrap.GenerateInitialKeys(ctx, queries)
	log.Println("Bootstrap: JWT keys initialized")

	log.Println("Running bootstrap: creating admin user...")
	bootstrap.CreateAdmin(ctx, queries, auth.HashedPassword)
	log.Println("Bootstrap: admin user initialized")
}
