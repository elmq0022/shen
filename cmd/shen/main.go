package main

import (
	"context"
	"log"
	"net/http"
	"os"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/auth"
	"github.com/elmq0022/shen/internal/bootstrap"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	ctx := context.Background()

	// Initialize database connection
	pool, queries := initDatabase(ctx)
	defer pool.Close()

	// Run bootstrap functions on startup
	runBootstrap(ctx, queries)

	log.Println("Shen bootstrap completed successfully")

	// Initialize Echo server
	e := initServer()

	// Start server
	port := getEnv("SHEN_PORT", "8080")
	log.Printf("Starting Shen server on port %s", port)
	if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
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

func initServer() *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	// Health check endpoint
	e.GET("/api/v1/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
		})
	})

	return e
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
