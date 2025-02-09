//nolint:exhaustruct
package recipes_test

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/AlejandroHerr/cook-book-go/internal/common"
	"github.com/AlejandroHerr/cook-book-go/internal/common/testutil"
	"github.com/AlejandroHerr/cook-book-go/internal/recipes"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPgRecipesRepository(t *testing.T) {
	t.Run("PgRecipesRepository", func(t *testing.T) {
		t.Parallel()

		repo := recipes.NewPgRecipesRepository(pgPool)

		t.Run("GetAll", func(t *testing.T) {
			t.Run("When there are recipes it returns the recipes", func(t *testing.T) {
				rs, err := repo.GetAll(context.Background())

				got := make([]recipes.Recipe, len(rs))
				for i, recipe := range rs {
					got[i] = *recipe
				}

				expected := make([]recipes.Recipe, len(fixtures))

				for i, recipe := range fixtures {
					expected[i] = *recipe
					expected[i].Ingredients = nil
				}

				require.NoError(t, err, "error should be nil")
				require.Equal(t, len(got), len(expected), "should return the same number of recipes")
				require.Equal(t, got, expected, "recipes should be equal")
			})
		})

		t.Run("GetById", func(t *testing.T) {
			t.Run("When recipe exists it returns the recipe", func(t *testing.T) {
				recipe, err := repo.GetByID(context.Background(), fixtures[0].ID.String())

				require.NoError(t, err, "error should be nil")
				RequireRecipeEqual(t, *fixtures[0], *recipe, "recipe should be equal")
			})
			t.Run("When recipe does not exists it returns an error", func(t *testing.T) {
				_, err := repo.GetByID(context.Background(), uuid.NewString())

				var errNotFound *common.ErrNotFound

				require.ErrorAs(t, err, &errNotFound, "error should be ErrNotFound")
			})
		})
		t.Run("GetBySlug", func(t *testing.T) {
			t.Run("When recipe exists it returns the recipe", func(t *testing.T) {
				recipe, err := repo.GetBySlug(context.Background(), fixtures[0].Slug())

				require.NoError(t, err, "error should be nil")
				RequireRecipeEqual(t, *fixtures[0], *recipe, "recipe should be equal")
			})
			t.Run("When recipe does not exists it returns an error", func(t *testing.T) {
				_, err := repo.GetBySlug(context.Background(), "not-found")

				var errNotFound *common.ErrNotFound

				require.ErrorAs(t, err, &errNotFound, "error should be ErrNotFound")
			})
		})
		t.Run("Create", func(t *testing.T) {
			t.Run("creates a recipe", func(t *testing.T) {
				recipe := new(recipes.Recipe)
				testutil.MustMakeStructFixture(recipe)

				recipe.Ingredients = fixtures[0].Ingredients

				created, err := repo.Create(context.Background(), *recipe)

				require.NoError(t, err, "error should be nil")

				// compare the recipe with the created one
				// ignoring created and updayed at
				recipe.CreatedAt = created.CreatedAt
				recipe.UpdatedAt = created.UpdatedAt

				require.NotEqual(t, uuid.Nil, created.ID, "created recipe should have an id")
				require.NotEqual(t, time.Time{}, created.CreatedAt, "created at should not be zero")
				require.True(t, created.CreatedAt.Before(time.Now()), "created at should be before now")
				require.NotEqual(t, time.Time{}, created.UpdatedAt, "updated at should not be zero")
				require.True(t, created.UpdatedAt.Before(time.Now()), "updated at should be before now")
				RequireRecipeEqual(t, *recipe, *created, "should return the created recipe")

				dbRecipe, err := repo.GetByID(context.Background(), recipe.ID.String())

				require.NoError(t, err, "error should be nil")
				RequireRecipeEqual(t, *recipe, *dbRecipe, "recipe should be created in the db")
			})
			t.Run("fails if the recipe is duplicated", func(t *testing.T) {
				recipe := *fixtures[0]
				recipe.ID = uuid.New()
				_, err := repo.Create(context.Background(), recipe)

				var errDuplicatedKey *common.ErrDuplicateKey

				require.ErrorAs(t, err, &errDuplicatedKey, "should fail with ErrDuplicateKey")

				_, err = repo.GetByID(context.Background(), recipe.ID.String())

				var errNotFound *common.ErrNotFound

				require.ErrorAs(t, err, &errNotFound, "should not have created the recipe")
			})
			t.Run("fails if the recipe is duplicated2", func(t *testing.T) {
				recipe := recipes.Recipe{}
				testutil.MustMakeStructFixture(&recipe)

				recipe.Ingredients[0].ID = uuid.New()
				_, err := repo.Create(context.Background(), recipe)

				var errCoinstrain *common.ErrConstrain

				require.ErrorAs(t, err, &errCoinstrain, "should fail with ErrDuplicateKey")
			})
		})
		t.Run("Update", func(t *testing.T) {
			t.Run("updates a recipe", func(t *testing.T) {
				recipe := *fixtures[1]
				for j, i := range recipe.Ingredients {
					recipe.Ingredients[j] = recipes.RecipeIngredient{
						ID:       i.ID,
						Name:     i.Name,
						Kind:     i.Kind,
						Quantity: gofakeit.Float64Range(0.1, 200),
						Unit:     recipes.Units[gofakeit.Number(0, len(recipes.Units)-1)],
					}
				}

				updated, err := repo.Update(context.Background(), recipe)
				require.NoError(t, err, "error should be nil")

				require.True(t, updated.UpdatedAt.After(recipe.UpdatedAt), "updated at should be updated")

				recipe.UpdatedAt = updated.UpdatedAt
				RequireRecipeEqual(t, recipe, *updated, "should return the updated recipe")

				dbRecipe, err := repo.GetByID(context.Background(), recipe.ID.String())

				require.NoError(t, err, "error should be nil")
				RequireRecipeEqual(t, recipe, *dbRecipe, "recipe should be updated in the db")
			})
			t.Run("return ErrConstrain if the ingredients do not exist", func(t *testing.T) {
				recipe := new(recipes.Recipe)
				testutil.MustMakeStructFixture(recipe)
				recipe.ID = fixtures[1].ID

				_, err := repo.Update(context.Background(), *recipe)

				var errConstrain *common.ErrConstrain

				require.ErrorAs(t, err, &errConstrain, "error should be ErrConstrain")
			})
			t.Run("returns ErrNotFound if recipe does not exist", func(t *testing.T) {
				recipe := new(recipes.Recipe)
				testutil.MustMakeStructFixture(recipe)

				_, err := repo.Update(context.Background(), *recipe)

				var errNotFound *common.ErrNotFound

				require.ErrorAs(t, err, &errNotFound, "error should be ErrNotFound")
			})
		})
		t.Run("Delete", func(t *testing.T) {
			t.Run("deletes a recipe", func(t *testing.T) {
				err := repo.Delete(context.Background(), fixtures[9].ID.String())

				require.NoError(t, err, "error should be nil")

				_, err = repo.GetByID(context.Background(), fixtures[9].ID.String())

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

func RequireRecipeEqual(t *testing.T, expected, got recipes.Recipe, msgAndArgs ...interface{}) {
	t.Helper()

	slices.SortFunc(expected.Ingredients, func(l, r recipes.RecipeIngredient) int {
		return strings.Compare(l.Name, r.Name)
	})
	slices.SortFunc(got.Ingredients, func(l, r recipes.RecipeIngredient) int {
		return strings.Compare(l.Name, r.Name)
	})

	require.Equal(t, expected, got, msgAndArgs...)
}
