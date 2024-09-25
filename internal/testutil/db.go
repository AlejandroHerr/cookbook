package testutil

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/AlejandroHerr/cook-book-go/internal/common/infra/db"
	"github.com/jackc/pgx/v5"
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
	port, err := strconv.Atoi(getEnv("TEST_DB_PORT", defaultTestDBPort))
	if err != nil {
		return nil, fmt.Errorf("parsing port: %w", err)
	}

	return &db.Config{
		Host:     getEnv("TEST_DB_HOST", defaultTestDBHost),
		Port:     port,
		User:     &user,
		Password: &password,
		Database: getEnv("TEST_DB_NAME", defaultTestDBName),
	}, nil
}

// MustConnect creates a new database connection or panics
func MustConnect(ctx context.Context) *pgxpool.Pool {
	config, err := DefaultTestDBConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to get test database config: %v", err))
	}

	pool, err := db.Connect(ctx, config, 60, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to test database: %v", err))
	}

	return pool
}

// ResetDB truncates all tables in the test database
// Useful for cleaning up between tests
func ResetDB(ctx context.Context, pool *pgxpool.Pool) error {
	// Get all table names
	rows, err := pool.Query(ctx, `
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public'
	`)
	if err != nil {
		return fmt.Errorf("getting table names: %w", err)
	}
	defer rows.Close()

	var tables []string

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("scanning table name: %w", err)
		}

		tables = append(tables, tableName)
	}

	// Truncate all tables
	if len(tables) > 0 {
		query := fmt.Sprintf("TRUNCATE %s CASCADE", pgx.Identifier(tables).Sanitize())
		if _, err := pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("truncating tables: %w", err)
		}
	}

	return nil
}

// helper function to get environment variable with default fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return fallback
}
