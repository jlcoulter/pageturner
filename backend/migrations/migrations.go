package migrations

import (
	//"bookTracker/config"
	"fmt"
	//"go/build"
	"log"

	//"log"
	"database/sql"
	"os"

	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func buildConnString() string {
	host := getEnv("DB_HOST", "10.1.1.50")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "username")
	password := getEnv("DB_PASSWORD", "password")
	dbname := getEnv("DB_NAME", "default_database")
	sslmode := getEnv("SSL_MODE", "disable")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	return connStr
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func OpenSqlConnection() *sql.DB {
	connStr := buildConnString()
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		panic(err)
	}

	return db
}

func OpenPgxConnection() *pgxpool.Pool {
	connStr := buildConnString()
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		panic(err)
	}

	return pool
}

func MigrationsUp() {
	db := OpenSqlConnection()
	defer db.Close()

	if err := goose.Up(db, "./migrations"); err != nil {
		log.Fatal(err)
	}
}

func MigrationsDown() {
	db := OpenSqlConnection()
	defer db.Close()

	if err := goose.Down(db, "./migrations"); err != nil {
		log.Fatal(err)
	}
}
