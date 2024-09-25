package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
)

type Config struct {
	Password *string
	User     *string
	Port     int
	Host     string
	Database string
	Timeout  time.Duration
}

func Connect(ctx context.Context, config *Config) (*pgxpool.Pool, error) {
	userPass := ""

	if config.User != nil && config.Password != nil {
		userPass = fmt.Sprintf("user=%s password=%s", *config.User, *config.Password)
	}

	pgxConfig, err := pgxpool.ParseConfig(fmt.Sprintf(
		"host=%s port=%d dbname=%s %s",
		config.Host,
		config.Port,
		config.Database,
		userPass,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}

	pgxConfig.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())

		return nil
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	handlerCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(handlerCtx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return pool, nil
}
