package pg

import (
	"errors"

	"github.com/AlejandroHerr/cookbook/internal/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func HandleScanError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return &common.NotFoundError{Err: err}
	}

	return &common.UnexpectedError{Err: err}
}

func HandleExecError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return &common.NotFoundError{Err: err}
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return &common.DuplicateError{Key: pgErr.ConstraintName, Err: err}
		case "23503": // foreign_key_violation
			return &common.ConstrainError{Constraint: pgErr.ConstraintName, Err: err}
		}
	}

	return &common.UnexpectedError{Err: err}
}
