package recipes_test

import (
	"context"
	"testing"

	"github.com/AlejandroHerr/cookbook/internal/infra"
	"github.com/AlejandroHerr/cookbook/internal/recipes"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	testDBPool *pgxpool.Pool
	fixtures   *recipes.Fixtures
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	testDBPool = infra.ConnectToTestDB(ctx)
	defer testDBPool.Close()

	fixtures = recipes.GenerateFixtures(10)

	recipes.LoadFixtures(ctx, testDBPool, fixtures)
	defer recipes.CleanupFixtures(ctx, testDBPool)

	m.Run()
}
