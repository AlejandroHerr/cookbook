package infra

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectToTestDB(ctx context.Context) *pgxpool.Pool {
	user := "tests"
	password := "123456"
	config := Config{
		Host:     "localhost",
		Port:     5433,
		Database: "tests",
		User:     &user,
		Password: &password,
	}

	var err error

	testPool, err := Connect(ctx, &config)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	return testPool
}
