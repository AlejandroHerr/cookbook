package testutil

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/AlejandroHerr/cook-book-go/internal/common/infra/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultTestDBHost     = "localhost"
	defaultTestDBPort     = "5439"
	defaultTestDBUser     = "tests"
	defaultTestDBPassword = "123456"
	defaultTestDBName     = "tests"
)

// DefaultTestDBConfig returns default test database configuration
// which can be overridden by environment variables
func DefaultTestDBConfig() (*db.Config, error) {
	user := getEnv("DB_USER", defaultTestDBUser)
	password := getEnv("DB_PASSWORD", defaultTestDBPassword)
	// string to int
	port, err := strconv.Atoi(getEnv("DB_PORT", defaultTestDBPort))
	if err != nil {
		return nil, fmt.Errorf("parsing port: %w", err)
	}

	return &db.Config{
		Host:     getEnv("DB_HOST", defaultTestDBHost),
		Port:     port,
		User:     &user,
		Password: &password,
		Database: getEnv("DB_DATABASE", defaultTestDBName),
	}, nil
}

// MustConnect creates a new database connection or panics
func MustConnect(ctx context.Context) *pgxpool.Pool {
	config, err := DefaultTestDBConfig()
	if err != nil {
		log.Fatalf("failed to get test database config: %v", err)
	}

	pool, err := db.Connect(ctx, config, 60, nil)
	if err != nil {
		log.Fatalf("failed to connect to test database: %v", err)
	}

	return pool
}

// helper function to get environment variable with default fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return fallback
}
