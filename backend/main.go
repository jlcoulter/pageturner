package main

import (
	"bookTracker/internal/api"
	"bookTracker/internal/db/generated"
	"bookTracker/internal/repository"
	"bookTracker/migrations"
	"bookTracker/scripts"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	var action string

	// Check for non-interactive mode (for Docker/automation)
	if len(os.Args) > 1 {
		action = os.Args[1]
	} else {
		fmt.Print("What would you like to do?\n(M): Migrate\n(I): Import\n(S): Start Server\n")
		fmt.Scan(&action)
	}

	if action == "M" || action == "m" || action == "--migrate" {
		log.Println("Running database migrations...")
		migrations.MigrationsUp()
		log.Println("Migrations completed successfully")
		return
	}

	if action == "I" {
		scripts.ImportOpenLibraryData(ctx)
	}

	if action == "S" || action == "s" {
		connStr := getDatabaseURL()
		pool, err := pgxpool.New(ctx, connStr)
		if err != nil {
			log.Fatalf("failed to create connection pool: %v", err)
		}
		defer pool.Close()

		// Verify connection
		if err := pool.Ping(ctx); err != nil {
			log.Fatalf("failed to ping database: %v", err)
		}

		queries := generated.New(pool)
		bookRepo := repository.NewBookRepo(queries)
		openLibraryRepo := repository.NewOpenLibraryRepo(queries)

		server := api.NewServer(bookRepo, openLibraryRepo)
		r := chi.NewRouter()
		r.Mount("/api", server.Routes())

		srv := &http.Server{
			Addr:    ":8080",
			Handler: r,
		}

		// Start server in goroutine
		go func() {
			log.Println("Starting server on :8080")
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("server failed: %v", err)
			}
		}()

		// Wait for interrupt signal
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")

		// Graceful shutdown with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("server forced to shutdown: %v", err)
		}

		log.Println("Server exited")
	}
}

func getDatabaseURL() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "username")
	password := getEnv("DB_PASSWORD", "password")
	dbname := getEnv("DB_NAME", "default_database")
	sslmode := getEnv("SSL_MODE", "disable")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}