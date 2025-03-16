//nolint:errcheck,wrapcheck
package mocks

import (
	"context"

	"github.com/AlejandroHerr/cookbook/internal/common"
	"github.com/stretchr/testify/mock"
)

var (
	_ common.TransactionManager = (*TransactionManager)(nil)
	_ common.UnitOfWork         = (*UnitOfWork)(nil)
)

type TransactionManager struct {
	mock.Mock
}

func (m *TransactionManager) Begin(ctx context.Context) (context.Context, common.UnitOfWork, error) {
	args := m.Called(ctx)
	return ctx, args.Get(1).(common.UnitOfWork), args.Error(2)
}

type UnitOfWork struct {
	mock.Mock
}

func (m *UnitOfWork) Rollback(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *UnitOfWork) Commit(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
