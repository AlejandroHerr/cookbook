package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
)

type MockPgxPool struct {
	mock.Mock
}

func (m *MockPgxPool) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	return args.Get(0).(pgx.Tx), args.Error(1) //nolint:wrapcheck
}

type MockTx struct {
	mock.Mock
}

func (m *MockTx) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)

	return args.Get(0).(pgx.Tx), args.Error(1) //nolint:wrapcheck
}

func (m *MockTx) Commit(ctx context.Context) error {
	args := m.Called(ctx)

	return args.Error(0) //nolint:wrapcheck
}

func (m *MockTx) Rollback(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0) //nolint:wrapcheck
}

func (m *MockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	args := m.Called(ctx, tableName, columnNames, rowSrc)

	return args.Get(0).(int64), args.Error(1) //nolint:wrapcheck
}

func (m *MockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	args := m.Called(ctx, b)

	return args.Get(0).(pgx.BatchResults)
}

func (m *MockTx) LargeObjects() pgx.LargeObjects {
	args := m.Called()

	return args.Get(0).(pgx.LargeObjects)
}

func (m *MockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	args := m.Called(ctx, name, sql)

	return args.Get(0).(*pgconn.StatementDescription), args.Error(1) //nolint:wrapcheck
}

func (m *MockTx) Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error) {
	args := m.Called(ctx, sql, arguments)

	return args.Get(0).(pgconn.CommandTag), args.Error(1) //nolint:wrapcheck
}

func (m *MockTx) Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error) {
	args := m.Called(ctx, sql, arguments)

	return args.Get(0).(pgx.Rows), args.Error(1) //nolint:wrapcheck
}

func (m *MockTx) QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row {
	args := m.Called(ctx, sql, arguments)

	return args.Get(0).(pgx.Row)
}

func (m *MockTx) Conn() *pgx.Conn {
	args := m.Called()

	return args.Get(0).(*pgx.Conn)
}
