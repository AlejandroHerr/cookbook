package routers_test

import (
	"context"
	"testing"

	"github.com/AlejandroHerr/cook-book-go/internal/core/model"
	"github.com/AlejandroHerr/cook-book-go/internal/core/repo"
	"github.com/AlejandroHerr/cook-book-go/internal/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	fixtures *model.Fixtures
	pgPool   *pgxpool.Pool
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgPool = testutil.MustConnect(ctx)
	fixtures = model.MustMakeFixtures(10)
	repo.MustInstertFixtures(ctx, pgPool, fixtures)

	defer func() {
		repo.MustCleanup(ctx, pgPool)
		pgPool.Close()
	}()

	m.Run()
}
