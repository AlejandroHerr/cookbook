package db

import (
	"context"
	"fmt"

	"github.com/AlejandroHerr/cookbook/internal/common"
	"github.com/jackc/pgx/v5"
)

var (
	_ common.Transaction        = (*PgxTransaction)(nil)
	_ common.TransactionManager = (*PgxTransactionManager)(nil)
)

type PgxPool interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type PgxTransactionManager struct {
	pool PgxPool
}

func MakePgxTransactionManager(pool PgxPool) *PgxTransactionManager {
	return &PgxTransactionManager{
		pool: pool,
	}
}

func (m *PgxTransactionManager) Begin(ctx context.Context) (common.Transaction, error) {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	return &PgxTransaction{
		ctx: ctx,
		tx:  tx,
	}, nil
}

type PgxTransaction struct {
	ctx context.Context
	tx  pgx.Tx
}

func (s *PgxTransaction) Rollback() error {
	err := s.tx.Rollback(s.ctx)
	if err != nil {
		return fmt.Errorf("rollback transaction: %w", err)
	}

	return nil
}

func (s *PgxTransaction) Commit() error {
	err := s.tx.Commit(s.ctx)
	if err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *PgxTransaction) Transaction() any {
	return s.tx
}

func GetPGXDB(ctx context.Context, pgxDB PGXDB) PGXDB {
	if session, ok := ctx.Value(common.TransactionContextKey{}).(*PgxTransaction); ok {
		if tx, isTx := session.Transaction().(pgx.Tx); isTx {
			return tx
		}
	}

	return pgxDB
}
