package recipes_test

import (
	"context"
	"testing"

	"github.com/AlejandroHerr/cookbook/internal/common/testutil"
	"github.com/AlejandroHerr/cookbook/internal/recipes"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPgIngredients(t *testing.T) {
	repo := recipes.NewPgIngredientsRepo(pgPool)

	t.Run("UpsertMany", func(t *testing.T) {
		t.Parallel()
		t.Run("upserts the ingredients and returns the full RecipeIngredient", func(t *testing.T) {
			NumExistingIngredients := 3
			NumIngredients := NumExistingIngredients + 1

			recipeIngredients := make([]recipes.RecipeIngredient, NumIngredients)
			recipeIngredientsDto := make([]recipes.CreateRecipeIngredientDTO, NumIngredients)

			for i := range recipeIngredients {
				testutil.MustMakeStructFixture(&recipeIngredients[i])
				recipeIngredientsDto[i] = recipes.CreateRecipeIngredientDTO{
					Name:     recipeIngredients[i].Name,
					Quantity: recipeIngredients[i].Quantity,
					Unit:     recipeIngredients[i].Unit,
				}
			}

			rows, err := pgPool.Query(
				context.Background(),
				`
          SELECT
            id, name, kind
          FROM
            ingredients
          LIMIT $1
        `,
				NumExistingIngredients,
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

			got, err := repo.UpsertMany(context.Background(), recipeIngredientsDto)
			require.NoError(t, err, "should not fail")
			require.Equal(t, recipeIngredients[0:NumExistingIngredients], got[0:NumExistingIngredients], "existing upserted ingredients should equal original ones")

			newIngredient := got[NumExistingIngredients]
			require.NotEqual(t, uuid.Nil, newIngredient.ID, "new ingredient should have a new ID")
			require.Equal(t, recipeIngredients[NumExistingIngredients].Name, newIngredient.Name, "new ingredient should have the same name")
			require.Nil(t, newIngredient.Kind, "new ingredient should have a nil kind")
			require.Equal(t, recipeIngredients[NumExistingIngredients].Quantity, newIngredient.Quantity, "new ingredient should have the same quantity")
			require.Equal(t, recipeIngredients[NumExistingIngredients].Unit, newIngredient.Unit, "new ingredient should have the same unit")
		})
	})
}
