package pg_test

import (
	"testing"

	"github.com/AlejandroHerr/cookbook/internal/common/pg"
	pgMocks "github.com/AlejandroHerr/cookbook/internal/common/pg/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mocks struct {
	mockPool *pgMocks.PgxPool
	mockTx   *pgMocks.Tx
}

func setup() *mocks {
	mockPool := &pgMocks.PgxPool{} //nolint:exhaustruct

	mockTx := &pgMocks.Tx{} //nolint:exhaustruct

	mockTx.On("Rollback", mock.Anything).Return(nil)
	mockTx.On("Commit", mock.Anything).Return(nil)

	mockPool.On("Begin", mock.Anything).Return(mockTx, nil)

	return &mocks{
		mockPool: mockPool,
		mockTx:   mockTx,
	}
}

func TestDBTransaction(t *testing.T) {
	t.Parallel()

	t.Run("TransactionManager", func(t *testing.T) {
		t.Parallel()
		t.Run("Begin", func(t *testing.T) {
			t.Run("calling begin returns a UnitOfWork that wraps a pgx.Tx", func(t *testing.T) {
				t.Parallel()

				mocks := setup()
				m := pg.NewTransactionManager(mocks.mockPool)
				ctx := t.Context()

				_, uow, err := m.Begin(ctx)
				require.NoError(t, err)

				err = uow.Commit(ctx)
				require.NoError(t, err)

				mocks.mockTx.AssertNumberOfCalls(t, "Commit", 1)

				err = uow.Rollback(ctx)
				require.NoError(t, err)

				mocks.mockTx.AssertNumberOfCalls(t, "Rollback", 1)
			})
			t.Run("calling begin returns a context with the UnitOfWork ", func(t *testing.T) {
				t.Parallel()

				mocks := setup()
				m := pg.NewTransactionManager(mocks.mockPool)
				ctx := t.Context()

				txCtx, _, err := m.Begin(ctx)
				require.NoError(t, err)

				require.Equal(t, mocks.mockTx, pg.GetDBTX(txCtx, mocks.mockPool), "uow should be equal to GetPGXDB")
			})
		})
	})
}
