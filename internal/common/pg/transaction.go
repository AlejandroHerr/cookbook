package pg

import (
	"context"
	"fmt"

	"github.com/AlejandroHerr/cookbook/internal/common"
	"github.com/jackc/pgx/v5"
)

var (
	_ common.UnitOfWork         = (*UnitOfWork)(nil)
	_ common.TransactionManager = (*TransactionManager)(nil)
)

type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type txKey struct{}

type UnitOfWork struct {
	tx pgx.Tx
}

func (u *UnitOfWork) Commit(ctx context.Context) error {
	err := u.tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("tx commit: %w", err)
	}

	return nil
}

func (u *UnitOfWork) Rollback(ctx context.Context) error {
	err := u.tx.Rollback(ctx)
	if err != nil {
		return fmt.Errorf("tx rollback: %w", err)
	}

	return nil
}

type TransactionManager struct {
	db TxBeginner
}

func NewTransactionManager(db TxBeginner) *TransactionManager {
	return &TransactionManager{db: db}
}

func (tm *TransactionManager) Begin(ctx context.Context) (context.Context, common.UnitOfWork, error) {
	tx, err := tm.db.Begin(ctx)
	if err != nil {
		return ctx, nil, fmt.Errorf("db begin: %w", err)
	}

	uow := &UnitOfWork{tx: tx}
	newCtx := context.WithValue(ctx, txKey{}, tx)

	return newCtx, uow, nil
}

func GetDBTX(ctx context.Context, dbtx DBTX) DBTX {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}

	return dbtx
}
