package repo_test

import (
	"context"
	"testing"

	"github.com/AlejandroHerr/cook-book-go/internal/core/model"
	"github.com/AlejandroHerr/cook-book-go/internal/core/repo"
	"github.com/AlejandroHerr/cook-book-go/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestPgIngredientsRecipe(t *testing.T) {
	t.Run("PgIngredientsRecipe", func(t *testing.T) {
		t.Parallel()

		repo := repo.NewPgIngredientsRepo(pgPool)

		t.Run("UpsertMany", func(t *testing.T) {
			t.Run("upserts ingredients by name and return the full ingredient", func(t *testing.T) {
				newIngredient := model.Ingredient{} //nolint:exhaustruct
				testutil.MustMakeStructFixture(&newIngredient)

				existing := []model.Ingredient{
					*fixtures.Ingredients[0],
					*fixtures.Ingredients[1],
					*fixtures.Ingredients[2],
				}

				names := make([]string, len(existing)+1)
				for i, ingredient := range existing {
					names[i] = ingredient.Name
				}

				names[len(existing)] = newIngredient.Name

				upserted, err := repo.UpsertMany(context.Background(), names)
				require.NoError(t, err, "error should be nil")
				require.Equal(t, existing[0:3], upserted[0:3], "ingredients should be equal")

				require.NotEqual(t, upserted[3].ID, newIngredient.ID, "new ingredient should have a new ID")
				require.Equal(t, newIngredient.Name, upserted[3].Name, "new ingredient should have the same name")
				require.Nil(t, upserted[3].Kind, "new ingredient should have the same kind")
			})
		})
	})
}
