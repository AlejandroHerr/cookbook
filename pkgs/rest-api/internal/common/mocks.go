package common

import (
	"context"

	"github.com/stretchr/testify/mock"
)

var (
	_ TransactionManager = (*MockTransactionManager)(nil)
	_ Transaction        = (*MockTransaction)(nil)
)

type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) Begin(ctx context.Context) (Transaction, error) {
	args := m.Called(ctx)
	return args.Get(0).(Transaction), args.Error(1)
}

type MockTransaction struct {
	mock.Mock
}

func (m *MockTransaction) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTransaction) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTransaction) Transaction() interface{} {
	return nil
}
