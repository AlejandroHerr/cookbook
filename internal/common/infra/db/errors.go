package db

import (
	"errors"

	"github.com/AlejandroHerr/cook-book-go/internal/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func HandlePgError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return &common.ErrNotFound{Err: err}
	}

	if errors.Is(err, pgx.ErrTooManyRows) {
		return &common.ErrTooManyRows{Err: err}
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return &common.ErrDuplicateKey{Key: pgErr.ConstraintName, Err: err}
		case "23503": // foreign_key_violation
			return &common.ErrConstrain{Constraint: pgErr.ConstraintName, Err: err}
		}
	}

	return &common.ErrUnexpected{Err: err}
}
