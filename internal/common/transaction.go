package common

import (
	"context"
)

type TransactionManager interface {
	Begin(ctx context.Context) (Transaction, error)
}

type Transaction interface {
	Rollback() error
	Commit() error
	Transaction() any
}

type TransactionContextKey struct{}
