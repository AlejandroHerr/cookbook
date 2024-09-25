package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
)

type Config struct {
	Password *string `env:"DB_PASS" json:"-"`
	User     *string `env:"DB_USER" json:"user,omitempty"`
	Port     int     `env:"DB_PORT,notEmpty,required" json:"port"`
	Host     string  `env:"DB_HOST,notEmpty,required" json:"host"`
	Database string  `env:"DB_DATABASE,notEmpty,required" json:"database"`
}

func Connect(
	ctx context.Context,
	config *Config,
	connTimeout time.Duration,
	logger pgx.QueryTracer,
) (*pgxpool.Pool, error) {
	userPass := ""

	if config.User != nil && config.Password != nil {
		userPass = fmt.Sprintf("user=%s password=%s", *config.User, *config.Password)
	}

	pgxConfig, err := pgxpool.ParseConfig(fmt.Sprintf(
		"host=%s port=%d dbname=%s %s sslmode=disable",
		config.Host,
		config.Port,
		config.Database,
		userPass,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}

	if logger != nil {
		pgxConfig.ConnConfig.Tracer = logger
	}

	pgxConfig.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())

		return nil
	}

	timeout := connTimeout
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
