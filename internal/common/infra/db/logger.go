package db

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

type PgxLogger struct {
	logger *slog.Logger
}

func NewPgxLogger(logger *slog.Logger) *PgxLogger {
	return &PgxLogger{
		logger: logger.With(slog.String("service", "db")),
	}
}

func (l PgxLogger) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	l.logger.DebugContext(ctx, "query start",
		"query", data.SQL,
		"args", data.Args,
	)

	return ctx
}

func (l PgxLogger) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	if data.Err != nil {
		l.logger.WarnContext(ctx, "query end",
			slog.Any("error", data.Err),
			slog.String("commandTag", data.CommandTag.String()),
		)
	} else {
		l.logger.DebugContext(ctx, "query end",
			slog.String("commandTag", data.CommandTag.String()),
		)
	}
}
