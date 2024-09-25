package repo_test

import (
	"context"
	"testing"
	"time"

	"github.com/AlejandroHerr/cook-book-go/internal/common"
	"github.com/AlejandroHerr/cook-book-go/internal/core/model"
	"github.com/AlejandroHerr/cook-book-go/internal/core/repo"
	"github.com/AlejandroHerr/cook-book-go/internal/core/testutil"
	ctu "github.com/AlejandroHerr/cook-book-go/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPgRecipesRepository(t *testing.T) {
	t.Run("PgRecipesRepository", func(t *testing.T) {
		t.Parallel()

		repo := repo.NewPgRecipesRepository(pgPool)

		t.Run("GetAll", func(t *testing.T) {
			t.Run("When there are recipes it returns the recipes", func(t *testing.T) {
				recipes, err := repo.GetAll(context.Background())

				got := make([]model.Recipe, len(recipes))
				for i, recipe := range recipes {
					got[i] = *recipe
				}

				expected := make([]model.Recipe, len(fixtures.Recipes))

				for i, recipe := range fixtures.Recipes {
					expected[i] = *recipe
					expected[i].Ingredients = nil
				}

				require.NoError(t, err, "error should be nil")
				require.Equal(t, got, expected, "recipes should be equal")
			})
		})
		t.Run("GetById", func(t *testing.T) {
			t.Run("When recipe exists it returns the recipe", func(t *testing.T) {
				recipe, err := repo.GetByID(context.Background(), fixtures.Recipes[0].ID.String())

				require.NoError(t, err, "error should be nil")
				require.Equal(t, fixtures.Recipes[0], recipe, "recipe should be equal")
			})
			t.Run("When recipe does not exists it returns an error", func(t *testing.T) {
				_, err := repo.GetByID(context.Background(), uuid.NewString())

				var errNotFound *common.ErrNotFound

				require.ErrorAs(t, err, &errNotFound, "error should be ErrNotFound")
			})
		})
		t.Run("GetBySlug", func(t *testing.T) {
			t.Run("When recipe exists it returns the recipe", func(t *testing.T) {
				recipe, err := repo.GetBySlug(context.Background(), fixtures.Recipes[1].Slug())

				require.NoError(t, err, "error should be nil")
				require.Equal(t, fixtures.Recipes[1], recipe, "recipe should be equal")
			})
			t.Run("When recipe does not exists it returns an error", func(t *testing.T) {
				_, err := repo.GetBySlug(context.Background(), "not-found")

				var errNotFound *common.ErrNotFound

				require.ErrorAs(t, err, &errNotFound, "error should be ErrNotFound")
			})
		})
		t.Run("Create", func(t *testing.T) {
			t.Run("creates a recipe", func(t *testing.T) {
				recipeIngredients := make([]model.RecipeIngredient, 3)

				for i := 0; i < 3; i++ {
					ingredient := fixtures.Ingredients[i]
					recipeIngredients[i] = model.RecipeIngredient{
						ID:       ingredient.ID,
						Name:     ingredient.Name,
						Kind:     ingredient.Kind,
						Unit:     model.Kilo,
						Quantity: 1.0,
					}
				}

				recipe := new(model.Recipe)
				ctu.MustMakeStructFixture(recipe)

				recipe.Ingredients = recipeIngredients

				created, err := repo.Create(context.Background(), *recipe)

				require.NoError(t, err, "error should be nil")
				require.NotEqual(t, uuid.Nil.String(), created.ID.String())
				require.True(t, recipe.CreatedAt.Before(time.Now()))
				require.True(t, recipe.UpdatedAt.Before(time.Now()))
				testutil.RequireRecipeEquals(t, *recipe, *created, true, true, "recipe should be equal")

				dbRecipe, err := repo.GetByID(context.Background(), recipe.ID.String())

				require.NoError(t, err, "error should be nil")
				require.NotEqual(t, uuid.Nil.String(), dbRecipe.ID.String())
				require.True(t, recipe.CreatedAt.Before(time.Now()))
				require.True(t, recipe.UpdatedAt.Before(time.Now()))
				testutil.RequireRecipeEquals(t, *recipe, *dbRecipe, true, false, "recipe should be equal")
			})
			t.Run("fails if the recipe is duplicated", func(t *testing.T) {
				recipe := new(model.Recipe)
				ctu.MustMakeStructFixture(recipe)

				recipe.ID = fixtures.Recipes[0].ID

				_, err := repo.Create(context.Background(), *recipe)

				var errDuplicatedKey *common.ErrDuplicateKey

				require.ErrorAs(t, err, &errDuplicatedKey, "error should be ErrDuplicateKey")

				dbRecipe, err := repo.GetByID(context.Background(), recipe.ID.String())

				require.NoError(t, err, "error should be nil")
				require.Equal(t, fixtures.Recipes[0], dbRecipe, "should not create the recipe")
			})
			t.Run("when one ingredients does not exist, it creates the recipe but fails", func(t *testing.T) {
				recipe := new(model.Recipe)
				ctu.MustMakeStructFixture(recipe)

				recipe.Ingredients = []model.RecipeIngredient{
					{
						ID:       uuid.New(),
						Name:     "not-found",
						Kind:     nil,
						Unit:     model.Kilo,
						Quantity: 1.0,
					},
				}
				_, err := repo.Create(context.Background(), *recipe)

				var errConstrain *common.ErrConstrain

				require.ErrorAs(t, err, &errConstrain, "error should be nil")

				dbRecipe, err := repo.GetByID(context.Background(), recipe.ID.String())

				require.NoError(t, err, "error should be nil")
				require.Len(t, dbRecipe.Ingredients, 0, "should create a recipe without ingredients")

				recipe.Ingredients = make([]model.RecipeIngredient, 0)
				testutil.RequireRecipeEquals(t, *recipe, *dbRecipe, true, false, "recipe should be equal without ingredients")
			})
		})

		t.Run("Update", func(t *testing.T) {
			t.Run("updates a recipe", func(t *testing.T) {
				baseRecipe := *fixtures.Recipes[5]

				newIngredients := make([]model.RecipeIngredient, 2)

				for _, i := range fixtures.Ingredients {
					found := false

					for _, ri := range baseRecipe.Ingredients {
						if i.ID != ri.ID {
							newRecipeIngredient := model.RecipeIngredient{
								ID:       i.ID,
								Name:     i.Name,
								Kind:     i.Kind,
								Unit:     model.Liter,
								Quantity: 4.5,
							}
							newIngredients[0] = newRecipeIngredient

							found = true

							break
						}
					}

					if found {
						break
					}
				}

				newIngredients[1] = model.RecipeIngredient{
					ID:       baseRecipe.Ingredients[0].ID,
					Name:     baseRecipe.Ingredients[0].Name,
					Kind:     baseRecipe.Ingredients[0].Kind,
					Unit:     model.Gram,
					Quantity: 23.76,
				}

				recipe := new(model.Recipe)
				ctu.MustMakeStructFixture(&recipe)
				recipe.ID = baseRecipe.ID
				recipe.Ingredients = newIngredients

				updated, err := repo.Update(context.Background(), *recipe)

				require.NoError(t, err, "error should be nil")

				testutil.RequireRecipeEquals(t, *recipe, *updated, true, false, "recipe should be updated")
				require.True(t, updated.UpdatedAt.After(baseRecipe.UpdatedAt), "updated at should be after the base recipe updated at")
				require.Equal(t, baseRecipe.CreatedAt, updated.CreatedAt, "created at should not change")

				dbRecipe, err := repo.GetByID(context.Background(), recipe.ID.String())
				//
				require.NoError(t, err, "error should be nil")
				testutil.RequireRecipeEquals(t, *recipe, *dbRecipe, true, false, "recipe should be updated")
				require.True(t, dbRecipe.UpdatedAt.After(baseRecipe.UpdatedAt), "updated at should be after the base recipe updated at")
				require.Equal(t, baseRecipe.CreatedAt, dbRecipe.CreatedAt, "created at should not change")
			})
			t.Run("fails if one of the ingredients does not exist", func(t *testing.T) {
				baseRecipe := fixtures.Recipes[5]
				recipe := *baseRecipe
				recipe.Ingredients[2].ID = uuid.New()

				_, err := repo.Update(context.Background(), recipe)

				var errConstrain *common.ErrConstrain

				require.ErrorAs(t, err, &errConstrain, "error should be ErrConstrain")
			})
			t.Run("returns not found if recipe does not exist", func(t *testing.T) {
				updatedRecipe := new(model.Recipe)
				ctu.MustMakeStructFixture(updatedRecipe)

				_, err := repo.Update(context.Background(), *updatedRecipe)

				var errNotFound *common.ErrNotFound

				require.ErrorAs(t, err, &errNotFound, "error should be ErrNotFound")
			})
		})
		t.Run("Delete", func(t *testing.T) {
			t.Run("deletes a recipe", func(t *testing.T) {
				err := repo.Delete(context.Background(), fixtures.Recipes[0].ID.String())

				require.NoError(t, err, "error should be nil")

				_, err = repo.GetByID(context.Background(), fixtures.Recipes[0].ID.String())

				var errNotFound *common.ErrNotFound

				require.ErrorAs(t, err, &errNotFound, "should not find the recipe after deleting")
			})
			t.Run("returns no error if recipe does not exists", func(t *testing.T) {
				err := repo.Delete(context.Background(), uuid.NewString())

				require.NoError(t, err, "error should be nil")
			})
		})
	})
}
