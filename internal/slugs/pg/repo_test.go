package pg_test

import (
	"context"
	"strconv"
	"testing"

	pgtestutil "github.com/AlejandroHerr/cookbook/internal/common/pg/testutil"
	"github.com/AlejandroHerr/cookbook/internal/slugs/pg"
	"github.com/AlejandroHerr/cookbook/internal/sql"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPgSlugsRepo(t *testing.T) {
	t.Parallel()

	db := pgtestutil.MustConnect(t)

	queries := sql.New()

	tx, err := db.Begin(t.Context())
	if err != nil {
		panic(err)
	}

	t.Cleanup(func() {
		tx.Rollback(context.Background()) //nolint:usetesting
	})

	fixtures := make([]string, 100)
	for i := range fixtures {
		title := gofakeit.Sentence(10)

		var slug string
		if i%2 == 0 {
			slug = `slug-even-` + strconv.Itoa(i)
		} else {
			slug = `slug-odd-` + strconv.Itoa(i)
		}

		_, err = queries.CreateRecipe(t.Context(), tx, sql.CreateRecipeParams{ //nolint:exhaustruct
			ID:    uuid.NewString(),
			Title: title,
			Slug:  slug,
		})
		if err != nil {
			panic(err)
		}

		fixtures[i] = slug
	}

	err = tx.Commit(t.Context())
	if err != nil {
		panic(err)
	}

	t.Cleanup(func() {
		db.Exec(context.Background(), "TRUNCATE TABLE recipes CASCADE") //nolint:usetesting
	})

	t.Run("GetSlugs", func(t *testing.T) {
		t.Parallel()
		t.Run("should return slugs if the entity exists", func(t *testing.T) {
			repo := pg.NewRepo(db)

			slugs, err := repo.GetSlugs(t.Context(), "recipes", "slug-even")
			require.NoError(t, err)

			require.Len(t, slugs, 50, "should return 50 slugs")
			require.Contains(t, slugs, "slug-even-0")
		})
		t.Run("should return an error if the entity does not exists", func(t *testing.T) {
			repo := pg.NewRepo(db)

			_, err = repo.GetSlugs(t.Context(), "nonexistent", "slug-even")
			require.Error(t, err, "should return an error")
		})
	})
}
