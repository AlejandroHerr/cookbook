package ingredients_test

import (
	"context"
	"testing"

	"github.com/AlejandroHerr/cook-book-go/internal/infra"
	"github.com/AlejandroHerr/cook-book-go/internal/ingredients"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	testDBPool *pgxpool.Pool
	fixtures   *ingredients.Fixtures
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	testDBPool = infra.ConnectToTestDB(ctx)
	defer testDBPool.Close()

	fixtures = ingredients.GenerateFixtures(10)

	ingredients.LoadFixtures(ctx, testDBPool, fixtures)
	defer ingredients.CleanupFixtures(ctx, testDBPool)

	m.Run()
}
