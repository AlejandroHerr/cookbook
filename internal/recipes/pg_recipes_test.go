//nolint:exhaustruct
package recipes_test

import (
	"context"
	"testing"
	"time"

	"github.com/AlejandroHerr/cookbook/internal/common"
	"github.com/AlejandroHerr/cookbook/internal/common/infra/db"
	"github.com/AlejandroHerr/cookbook/internal/common/testutil"
	"github.com/AlejandroHerr/cookbook/internal/recipes"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestPgRecipesRepository(t *testing.T) {
	t.Run("PgRecipesRepository", func(t *testing.T) {
		t.Parallel()

		repo := recipes.MakePgRecipesRepository(pgPool)

		t.Run("GetAll", func(t *testing.T) {
			t.Run("When there are recipes it returns the recipes", func(t *testing.T) {
				got, err := repo.GetAll(context.Background())
				require.NoError(t, err, "error should be nil")

				expected := make([]recipes.Recipe, len(fixtures))

				for i, recipe := range fixtures {
					expected[i] = *recipe
					expected[i].Ingredients = nil
				}

				require.Equal(t, len(expected), len(got), "should return the same number of recipes")
				RequireRecipesEqual(t, expected, got, "recipes should be equal")
			})
		})

		t.Run("GetById", func(t *testing.T) {
			t.Run("When recipe exists it returns the recipe", func(t *testing.T) {
				recipe, err := repo.GetByID(context.Background(), fixtures[0].ID.String())

				require.NoError(t, err, "error should be nil")
				RequireRecipeEqual(t, *fixtures[0], *recipe, RecipeEqualityOptions{}, "recipe should be equal")
			})
			t.Run("When recipe does not exists it returns an error", func(t *testing.T) {
				_, err := repo.GetByID(context.Background(), uuid.NewString())

				var errNotFound *common.NotFoundError

				require.ErrorAs(t, err, &errNotFound, "error should be ErrNotFound")
			})
		})
		t.Run("GetBySlug", func(t *testing.T) {
			t.Run("When recipe exists it returns the recipe", func(t *testing.T) {
				recipe, err := repo.GetBySlug(context.Background(), fixtures[0].Slug)

				require.NoError(t, err, "error should be nil")
				RequireRecipeEqual(t, *fixtures[0], *recipe, RecipeEqualityOptions{}, "recipe should be equal")
			})
			t.Run("When recipe does not exists it returns an error", func(t *testing.T) {
				_, err := repo.GetBySlug(context.Background(), "not-found")

				var errNotFound *common.NotFoundError

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

				require.NotEqual(t, uuid.Nil, created.ID, "created recipe should have an id")
				require.WithinDuration(t, created.CreatedAt, time.Now(), time.Second, "created at should be before now")
				require.WithinDuration(t, created.UpdatedAt, time.Now(), time.Second, "updated at should be before now")

				RequireRecipeEqual(t, *recipe, *created, RecipeEqualityOptions{
					IgnoreCreatedAt: true,
					IgnoreUpdatedAt: true,
					IgnoreID:        true,
				}, "should return the created recipe")

				dbRecipe, err := repo.GetByID(context.Background(), recipe.ID.String())
				require.NoError(t, err, "error should be nil")

				RequireRecipeEqual(t, *created, *dbRecipe, RecipeEqualityOptions{}, "recipe should be created in the db")
			})
			t.Run("fails if the recipe is duplicated", func(t *testing.T) {
				recipe := *fixtures[0]
				recipe.ID = uuid.New()
				_, err := repo.Create(context.Background(), recipe)

				var errDuplicatedKey *common.DuplicateError

				require.ErrorAs(t, err, &errDuplicatedKey, "should fail with ErrDuplicateKey")

				_, err = repo.GetByID(context.Background(), recipe.ID.String())

				var errNotFound *common.NotFoundError

				require.ErrorAs(t, err, &errNotFound, "should not have created the recipe")
			})
			t.Run("fails if ingredient does not exists", func(t *testing.T) {
				recipe := recipes.Recipe{}
				testutil.MustMakeStructFixture(&recipe)

				recipe.Ingredients[0].ID = uuid.New()
				_, err := repo.Create(context.Background(), recipe)

				var errUnexpected *common.UnexpectedError

				require.ErrorAs(t, err, &errUnexpected, "should fail with UnexpectedError")
			})
			t.Run("fails if ingredient is duplicated", func(t *testing.T) {
				recipe := recipes.Recipe{}
				testutil.MustMakeStructFixture(&recipe)

				recipe.Ingredients = append(recipe.Ingredients, recipe.Ingredients[0])
				_, err := repo.Create(context.Background(), recipe)

				var errUnexpected *common.UnexpectedError

				require.ErrorAs(t, err, &errUnexpected, "should fail with UnexpectedError")
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
						Unit:     recipes.Unit(recipes.Units[gofakeit.Number(0, len(recipes.Units)-1)]),
					}
				}

				updated, err := repo.Update(context.Background(), recipe)
				require.NoError(t, err, "error should be nil")

				require.True(t, updated.UpdatedAt.After(recipe.UpdatedAt), "updated at should be updated")
				require.WithinDuration(t, updated.UpdatedAt, time.Now(), time.Second, "updated at should be before now")

				RequireRecipeEqual(t, recipe, *updated, RecipeEqualityOptions{
					IgnoreUpdatedAt: true,
				}, "should return the updated recipe")

				dbRecipe, err := repo.GetByID(context.Background(), recipe.ID.String())

				require.NoError(t, err, "error should be nil")
				RequireRecipeEqual(t, *updated, *dbRecipe, RecipeEqualityOptions{}, "recipe should be updated in the db")
			})
			t.Run("fails if ingredient does not exists", func(t *testing.T) {
				recipe := new(recipes.Recipe)
				testutil.MustMakeStructFixture(recipe)
				recipe.ID = fixtures[1].ID

				_, err := repo.Update(context.Background(), *recipe)

				var errUnexpected *common.UnexpectedError

				require.ErrorAs(t, err, &errUnexpected, "error should be UnexpectedError")
			})
			t.Run("returns ErrNotFound if recipe does not exist", func(t *testing.T) {
				recipe := new(recipes.Recipe)
				testutil.MustMakeStructFixture(recipe)

				_, err := repo.Update(context.Background(), *recipe)

				var errNotFound *common.NotFoundError

				require.ErrorAs(t, err, &errNotFound, "error should be ErrNotFound")
			})
		})
		t.Run("Delete", func(t *testing.T) {
			t.Run("deletes a recipe", func(t *testing.T) {
				err := repo.Delete(context.Background(), fixtures[9].ID.String())

				require.NoError(t, err, "error should be nil")

				_, err = repo.GetByID(context.Background(), fixtures[9].ID.String())

				var errNotFound *common.NotFoundError

				require.ErrorAs(t, err, &errNotFound, "should not find the recipe after deleting")
			})
			t.Run("returns no error if recipe does not exists", func(t *testing.T) {
				err := repo.Delete(context.Background(), uuid.NewString())

				require.NoError(t, err, "error should be nil")
			})
		})
		t.Run("GetUniqueSlug", func(t *testing.T) {
			t.Run("returs default slug if empty", func(t *testing.T) {
				title := gofakeit.Adjective() + " " + gofakeit.Adjective() + " " + gofakeit.Dinner()
				expected := slug.Make(title)

				got, err := repo.GetUniqueSlug(context.Background(), title)
				require.NoError(t, err, "error should be nil")
				require.Equal(t, expected, got, "should return the same slug")
			})
			t.Run("returns next available slug", func(t *testing.T) {
				title := gofakeit.Adjective() + " " + gofakeit.Adjective() + " " + gofakeit.Dinner()
				slug := slug.Make(title)

				txm := db.MakePgxTransactionManager(pgPool)
				tx, err := txm.Begin(context.Background())
				require.NoError(t, err, "error should be nil")

				defer tx.Rollback()

				type ForTests struct {
					ID    uuid.UUID
					Title string
					Slug  string
				}

				testRecipes := []ForTests{
					{ID: uuid.New(), Title: title, Slug: slug},
					{ID: uuid.New(), Title: title, Slug: slug + "-1"},
					{ID: uuid.New(), Title: title, Slug: slug + "-2"},
					{ID: uuid.New(), Title: title, Slug: slug + "-3"},
					{ID: uuid.New(), Title: title, Slug: slug + "-5"},
					{ID: uuid.New(), Title: title, Slug: slug + "-vegan"},
				}

				pgtx := tx.Transaction().(pgx.Tx)
				for _, r := range testRecipes {
					_, err = pgtx.Exec(context.Background(),
						`
            INSERT INTO
              recipes (id, title, slug)
            VALUES
              ($1,$2,$3)
          `,
						r.ID, r.Title, r.Slug,
					)

					require.NoError(t, err, "error inserting test data")
				}

				expected := slug + "-4"
				got, err := repo.GetUniqueSlug(context.WithValue(context.Background(), common.TransactionContextKey{}, tx), title)
				require.NoError(t, err, "error should be nil")
				require.Equal(t, expected, got, "should return the next available slug")

				tx.Rollback()
			})
		})
	})
}
