package recipes_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/AlejandroHerr/cook-book-go/internal/common/testutil"
	"github.com/AlejandroHerr/cook-book-go/internal/recipes"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	fixtures []*recipes.Recipe
	pgPool   *pgxpool.Pool
)

func TestMain(m *testing.M) {
	err := godotenv.Load("../../.env.test")
	if err != nil {
		log.Fatalf("loading .env.test file: %v", err)
	}

	ctx := context.Background()

	pgPool = testutil.MustConnect(ctx)
	defer func() {
		recipes.MustCleanUpFixtures(ctx, pgPool)
		pgPool.Close()
	}()

	fixtures = recipes.MustMakeFixtures(100)

	err = recipes.InstertFixtures(context.Background(), pgPool, fixtures)
	if err != nil {
		panic(fmt.Errorf("inserting fixtures: %w", err))
	}

	m.Run()
}
