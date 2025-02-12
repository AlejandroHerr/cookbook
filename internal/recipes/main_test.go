package recipes_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AlejandroHerr/cookbook/internal/common/testutil"
	"github.com/AlejandroHerr/cookbook/internal/recipes"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	fixtures []*recipes.Recipe
	pgPool   *pgxpool.Pool
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgPool = testutil.MustConnect(ctx)
	defer func() {
		fmt.Println("Cleaning up fixtures")
		recipes.MustCleanUpFixtures(ctx, pgPool)
		pgPool.Close()
	}()

	fixtures = recipes.MustMakeFixtures(100)

	err := recipes.InstertFixtures(context.Background(), pgPool, fixtures)
	if err != nil {
		panic(fmt.Errorf("inserting fixtures: %w", err))
	}

	m.Run()
}
