package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	db "github.com/elmq0022/shen/db/sqlc"
	"github.com/elmq0022/shen/internal/bootstrap"
	"github.com/elmq0022/shen/internal/crypto"
	"github.com/elmq0022/shen/internal/handlers/jwks"
	mw "github.com/elmq0022/shen/internal/middleware"
	"github.com/elmq0022/shen/internal/routes"
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
	e := initServer(ctx, pool, queries)

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
	bootstrap.CreateAdmin(ctx, queries, crypto.HashedPassword)
	log.Println("Bootstrap: admin user initialized")
}

func initServer(ctx context.Context, pool *pgxpool.Pool, queries *db.Queries) *echo.Echo {
	e := echo.New()
	shenMiddleware, err := mw.NewMiddleware(ctx, queries)
	if err != nil {
		panic(fmt.Errorf("%w", err))
	}

	// Middleware
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	// Health check endpoint
	e.GET("/api/v1/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
		})
	})

	// JWKS endpoint (RFC 7517) - kept at root level per spec
	jwksHandler := jwks.NewHandler(queries)
	e.GET("/.well-known/jwks.json", jwksHandler.GetJWKS)

	// Auth routes
	api := e.Group("/api/v1")
	routes.RegisterAuthRoutes(api.Group("/auth"), queries)
	routes.RegisterUserRoutes(api.Group("/users"), pool, queries, shenMiddleware)
	routes.RegisterApplicationRoutes(api.Group("/applications"), pool, queries, shenMiddleware)
	routes.RegisterGroupRoutes(api.Group("/groups"), pool, queries, shenMiddleware)
	routes.RegisterTokenRoutes(api.Group("/tokens"), pool, queries, shenMiddleware)

	return e
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
