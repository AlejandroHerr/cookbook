package db

import (
	"context"

	"github.com/AlejandroHerr/cookbook/internal/common/logging"
	"github.com/jackc/pgx/v5"
)

type PgxLogger struct {
	logger logging.Logger
}

func NewPgxLogger(logger logging.Logger) *PgxLogger {
	return &PgxLogger{
		logger: logger,
	}
}

func (l PgxLogger) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	l.logger.Debugw("query start", "query", data.SQL, "args", data.Args)

	return ctx
}

func (l PgxLogger) TraceQueryEnd(_ context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	if data.Err != nil {
		l.logger.Errorw("query end", "error", data.Err.Error(), "commandTag", data.CommandTag.String())
	} else {
		l.logger.Debugw("query end", "commandTag", data.CommandTag.String())
	}
}
