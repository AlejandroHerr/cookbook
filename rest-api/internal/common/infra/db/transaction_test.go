package db_test

import (
	"context"
	"testing"

	"github.com/AlejandroHerr/cookbook/internal/common"
	"github.com/AlejandroHerr/cookbook/internal/common/infra/db"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mocks struct {
	mockPool *db.MockPgxPool
	mockTx   *db.MockTx
}

func setup() *mocks {
	mockPool := &db.MockPgxPool{} //nolint:exhaustruct

	mockTx := &db.MockTx{} //nolint:exhaustruct

	mockTx.On("Rollback", mock.Anything).Return(nil)
	mockTx.On("Commit", mock.Anything).Return(nil)

	mockPool.On("Begin", mock.Anything).Return(mockTx, nil)

	return &mocks{
		mockPool: mockPool,
		mockTx:   mockTx,
	}
}

func TestPgxTransactionManager(t *testing.T) {
	t.Parallel()

	t.Run("calling begin returns a PgxTransaction that wraps a pgx.Tx", func(t *testing.T) {
		t.Parallel()

		mocks := setup()
		m := db.MakePgxTransactionManager(mocks.mockPool)

		transaction, err := m.Begin(context.Background())
		require.NoError(t, err)

		err = transaction.Commit()
		require.NoError(t, err)

		mocks.mockTx.AssertNumberOfCalls(t, "Commit", 1)

		err = transaction.Rollback()
		require.NoError(t, err)

		mocks.mockTx.AssertNumberOfCalls(t, "Rollback", 1)

		tx := transaction.Transaction()

		require.Equal(t, mocks.mockTx, tx)

		beq := db.GetBatcherExecutorQuerier(context.WithValue(context.Background(), common.TransactionContextKey{}, transaction), &pgx.Conn{})

		require.Equal(t, mocks.mockTx, beq)

		mocks.mockPool.AssertExpectations(t)
	})
	t.Run("GetBatcherExecutorQuerier", func(t *testing.T) {
		t.Parallel()

		t.Run("returns the pgx.Tx if present in the context", func(t *testing.T) {
			mocks := setup()
			m := db.MakePgxTransactionManager(mocks.mockPool)

			transaction, err := m.Begin(context.Background())
			require.NoError(t, err)

			beq := db.GetBatcherExecutorQuerier(context.WithValue(context.Background(), common.TransactionContextKey{}, transaction), &pgx.Conn{})

			require.Equal(t, mocks.mockTx, beq)

			mocks.mockPool.AssertExpectations(t)
		})
		t.Run("return fallback if there is no transaction in the context", func(t *testing.T) {
			conn := &pgx.Conn{}

			beq0 := db.GetBatcherExecutorQuerier(context.Background(), conn)
			require.Equal(t, conn, beq0)

			beq1 := db.GetBatcherExecutorQuerier(
				context.WithValue(context.Background(), common.TransactionContextKey{}, nil),
				conn,
			)
			require.Equal(t, conn, beq1)
		})
	})
}
