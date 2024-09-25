package db

import (
	"context"
	"fmt"

	"github.com/AlejandroHerr/cook-book-go/internal/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func NewPgxTransactionManager(pool PgxPool) *PgxTransactionManager {
	return &PgxTransactionManager{
		pool: pool,
	}
}

func (m *PgxTransactionManager) Begin(ctx context.Context) (common.Transaction, error) {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	fmt.Println("PgxTransactionManager Begin")

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

type Executor interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Batcher interface {
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

type BatcherExecutorQuerier interface {
	Executor
	Querier
	Batcher
}

func GetBatcherExecutorQuerier(ctx context.Context, beq BatcherExecutorQuerier) BatcherExecutorQuerier {
	if session, ok := ctx.Value(common.TransactionContextKey{}).(*PgxTransaction); ok {
		if tx, ok := session.Transaction().(pgx.Tx); ok {
			return tx
		}
	}

	return beq
}
