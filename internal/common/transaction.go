package common

import "context"

type UnitOfWork interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type TransactionManager interface {
	Begin(ctx context.Context) (context.Context, UnitOfWork, error)
}
