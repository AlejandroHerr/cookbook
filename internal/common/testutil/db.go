package testutil

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/AlejandroHerr/cookbook/internal/common/infra/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultTestDBHost     = "localhost"
	defaultTestDBPort     = "5439"
	defaultTestDBUser     = "tests"
	defaultTestDBPassword = "123456"
	defaultTestDBName     = "tests"
)

// defaultTestDBConfig returns default test database configuration.
// which can be overridden by environment variables.
func defaultTestDBConfig() (*db.Config, error) {
	user := getEnv("POSTGRES_USER", defaultTestDBUser)
	password := getEnv("POSTGRES_PASSWORD", defaultTestDBPassword)
	// string to int
	port, err := strconv.Atoi(getEnv("POSTGRES_PORT", defaultTestDBPort))
	if err != nil {
		return nil, fmt.Errorf("parsing port: %w", err)
	}

	return &db.Config{
		Host:     getEnv("POSTGRES_HOST", defaultTestDBHost),
		Port:     port,
		User:     &user,
		Password: &password,
		Database: getEnv("POSTGRES_DB", defaultTestDBName),
	}, nil
}

// MustConnect creates a new database connection or panics.
func MustConnect(ctx context.Context) *pgxpool.Pool {
	config, err := defaultTestDBConfig()
	if err != nil {
		panic(fmt.Errorf("defaultTestDBConfig: %w", err))
	}

	pool, err := db.Connect(ctx, config, 60, nil)
	if err != nil {
		panic(fmt.Errorf("connect: %w", err))
	}

	return pool
}

// helper function to get environment variable with default fallback.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return fallback
}
