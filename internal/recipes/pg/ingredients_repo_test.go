//nolint:paralleltest
package pg_test

import (
	"testing"

	pgtestutil "github.com/AlejandroHerr/cookbook/internal/common/pg/testutil"
	"github.com/AlejandroHerr/cookbook/internal/recipes"
	"github.com/AlejandroHerr/cookbook/internal/recipes/pg"
	"github.com/AlejandroHerr/cookbook/internal/recipes/pg/fixtures"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPgIngredients(t *testing.T) {
	db := pgtestutil.MustConnect(t)
	_ = fixtures.MustMakeFixtures(t, 100, db)
	repo := pg.NewIngredientsRepo(db)

	t.Run("UpsertMany", func(t *testing.T) {
		t.Run("upserts the ingredients and returns the full RecipeIngredient", func(t *testing.T) {
			numExistingIngredients := 3
			numIngredients := numExistingIngredients + 1

			recipeIngredients := make([]recipes.RecipeIngredient, numIngredients)
			recipeIngredientsDto := make([]recipes.CreateRecipeIngredientDTO, numIngredients)

			for i := range recipeIngredients {
				gofakeit.Struct(&recipeIngredients[i])
				recipeIngredientsDto[i] = recipes.CreateRecipeIngredientDTO{
					Name:     recipeIngredients[i].Name,
					Quantity: recipeIngredients[i].Quantity,
					Unit:     recipeIngredients[i].Unit,
				}
			}

			rows, err := db.Query(
				t.Context(),
				`
          SELECT
            id, name, kind
          FROM
            ingredients
          LIMIT $1
        `,
				numExistingIngredients,
			)
			require.NoError(t, err, "error should be nil")

			defer rows.Close()

			for i := 0; rows.Next(); i++ {
				err = rows.Scan(
					&recipeIngredients[i].ID,
					&recipeIngredients[i].Name,
					&recipeIngredients[i].Kind,
				)

				recipeIngredientsDto[i].Name = recipeIngredients[i].Name

				require.NoError(t, err, "error should be nil")
			}

			got, err := repo.UpsertMany(t.Context(), recipeIngredientsDto)
			require.NoError(t, err, "should not fail")
			require.Equal(t, recipeIngredients[0:numExistingIngredients], got[0:numExistingIngredients], "existing upserted ingredients should equal original ones")

			newIngredient := got[numExistingIngredients]
			require.NotEqual(t, uuid.Nil, newIngredient.ID, "new ingredient should have a new ID")
			require.Equal(t, recipeIngredients[numExistingIngredients].Name, newIngredient.Name, "new ingredient should have the same name")
			require.Nil(t, newIngredient.Kind, "new ingredient should have a nil kind")
			require.InEpsilon(t, recipeIngredients[numExistingIngredients].Quantity, newIngredient.Quantity, 0.001, "new ingredient should have the same quantity")
			require.Equal(t, recipeIngredients[numExistingIngredients].Unit, newIngredient.Unit, "new ingredient should have the same unit")
		})
	})
}
