//nolint:exhaustruct,paralleltest
package pg_test

import (
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/AlejandroHerr/cookbook/internal/common"
	pgtestutil "github.com/AlejandroHerr/cookbook/internal/common/pg/testutil"
	"github.com/AlejandroHerr/cookbook/internal/recipes"
	"github.com/AlejandroHerr/cookbook/internal/recipes/pg"
	"github.com/AlejandroHerr/cookbook/internal/recipes/pg/fixtures"
	"github.com/AlejandroHerr/cookbook/internal/sql"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/stretchr/testify/require"
)

func TestPgRecipesRepository(t *testing.T) {
	db := pgtestutil.MustConnect(t)
	recipeFixtures := fixtures.MustMakeFixtures(t, 100, db)

	repo := pg.NewRecipesRepo(db)

	t.Run("List", func(t *testing.T) {
		t.Run("returns a list of Recipes without ingredients", func(t *testing.T) {
			got, err := repo.List(t.Context())
			require.NoError(t, err, "error should be nil")

			require.Len(t, got, len(recipeFixtures), "should return the same number of recipes")

			wanted := make([]recipes.Recipe, 0)

			copy(wanted, recipeFixtures)

			slices.SortFunc(wanted, func(i, j recipes.Recipe) int {
				if i.CreatedAt.Before(j.CreatedAt) {
					return 1
				}

				return 0
			})

			for i, r := range wanted {
				received := got[i]
				require.Equal(t, r.ID, received.ID, "recipe "+strconv.Itoa(i)+" should have the same id")
				require.Equal(t, r.Title, received.Title, "recipe "+strconv.Itoa(i)+" should have the same title")
				require.Equal(t, *r.Headline, *received.Headline, "recipe "+strconv.Itoa(i)+" should have the same headline")
				require.Equal(t, *r.Description, *received.Description, "recipe "+strconv.Itoa(i)+" should have the same description")
				require.Equal(t, *r.URL, *received.URL, "recipe "+strconv.Itoa(i)+" should have the same URL")
				require.Equal(t, r.Tags, received.Tags, "recipe "+strconv.Itoa(i)+" should have the same tags")
				require.Equal(t, *r.Steps, *received.Steps, "recipe "+strconv.Itoa(i)+" should have the same steps")
				require.Equal(t, r.Servings, received.Servings, "recipe "+strconv.Itoa(i)+" should have the same servings")
				require.Equal(t, *r.PrepTime, *received.PrepTime, "recipe "+strconv.Itoa(i)+" should have the same prep time")
				require.Equal(t, slug.Make(r.Title), received.Slug, "recipe "+strconv.Itoa(i)+" should have the same slug")
				require.WithinDuration(t, r.CreatedAt, received.CreatedAt, time.Second, "recipe "+strconv.Itoa(i)+" should have the same created at")
				require.WithinDuration(t, r.UpdatedAt, received.UpdatedAt, time.Second, "recipe "+strconv.Itoa(i)+" should have the same updated at")

				require.Nil(t, received.Ingredients, "recipe "+strconv.Itoa(i)+" should not have ingredients")
			}
		})
	})

	t.Run("GetById", func(t *testing.T) {
		t.Run("returns the Recipe with Ingredients by ID", func(t *testing.T) {
			wanted := recipeFixtures[1]
			got, err := repo.GetByID(t.Context(), wanted.ID.String())
			require.NoError(t, err, "error should be nil")

			require.Equal(t, wanted, *got, "recipe should yyybe equal")
		})
		t.Run("returns a NotFoundError when the recipe does not exist", func(t *testing.T) {
			_, err := repo.GetByID(t.Context(), uuid.NewString())

			var errNotFound *common.NotFoundError

			require.ErrorAs(t, err, &errNotFound, "error should be ErrNotFound")
		})
	})

	t.Run("GetBySlug", func(t *testing.T) {
		t.Run("returns the recipe with Ingredients by Slug", func(t *testing.T) {
			wanted := recipeFixtures[1]
			got, err := repo.GetBySlug(t.Context(), wanted.Slug)
			require.NoError(t, err, "error should be nil")

			require.Equal(t, wanted, *got, "recipe should yyybe equal")
		})
		t.Run("returns a NotFoundError when the recipe does not exist", func(t *testing.T) {
			_, err := repo.GetBySlug(t.Context(), "not-found")

			var errNotFound *common.NotFoundError

			require.ErrorAs(t, err, &errNotFound, "error should be ErrNotFound")
		})
	})
	t.Run("Create", func(t *testing.T) {
		t.Run("creates a new Recipe and returns it with the Ingredients", func(t *testing.T) {
			wanted := new(recipes.Recipe)
			gofakeit.Struct(wanted)

			wanted.Ingredients = recipeFixtures[0].Ingredients

			got, err := repo.Create(t.Context(), *wanted)

			require.NoError(t, err, "error should be nil")

			require.NotZero(t, got.ID, "created recipe should have an id")
			require.WithinDuration(t, got.CreatedAt, time.Now(), time.Second, "created at should be before now")
			require.WithinDuration(t, got.UpdatedAt, time.Now(), time.Second, "updated at should be before now")
			require.Equal(t, wanted.Title, got.Title, "recipe should have the same title")
			require.Equal(t, *wanted.Headline, *got.Headline, "recipe should have the same headline")
			require.Equal(t, *wanted.Description, *got.Description, "recipe should have the same description")
			require.Equal(t, *wanted.URL, *got.URL, "recipe should have the same URL")
			require.Equal(t, wanted.Tags, got.Tags, "recipe should have the same tags")
			require.Equal(t, *wanted.Steps, *got.Steps, "recipe should have the same steps")
			require.Equal(t, wanted.Servings, got.Servings, "recipe should have the same servings")
			require.Equal(t, *wanted.PrepTime, *got.PrepTime, "recipe should have the same prep time")
			require.Equal(t, slug.Make(wanted.Title), got.Slug, "recipe should have the same slug")
			require.Len(t, got.Ingredients, len(wanted.Ingredients), "recipe should have the same number of ingredients")

			for i, ingredient := range wanted.Ingredients {
				require.NotZero(t, got.Ingredients[i].ID, "ingredient should have an ID")
				require.Equal(t, ingredient.Name, got.Ingredients[i].Name, "ingredient should have the same name")
				require.InEpsilon(t, ingredient.Quantity, got.Ingredients[i].Quantity, 0.001, "ingredient should have the same quantity")
				require.Equal(t, ingredient.Unit, got.Ingredients[i].Unit, "ingredient should have the same unit")
			}

			gotInDB, err := repo.GetByID(t.Context(), wanted.ID.String())
			require.NoError(t, err, "error should be nil")

			require.Equal(t, *got, *gotInDB, "recipe should be created in the db")
		})
		t.Run("returns a ConstrainError when the ID is duplicated", func(t *testing.T) {
			recipe := recipeFixtures[0]
			recipe.Slug = gofakeit.Word()
			_, err := repo.Create(t.Context(), recipe)

			var errDuplicatedKey *common.DuplicateError

			require.ErrorAs(t, err, &errDuplicatedKey, "should fail with ErrDuplicateKey")

			_, err = repo.GetBySlug(t.Context(), recipe.Slug)

			var errNotFound *common.NotFoundError

			require.ErrorAs(t, err, &errNotFound, "should not have created the recipe")
		})
		t.Run("returns a DuplicateError when unique fields are duplicated", func(t *testing.T) {
			recipe := recipeFixtures[0]
			recipe.ID = uuid.New()
			_, err := repo.Create(t.Context(), recipe)

			var errDuplicatedKey *common.DuplicateError

			require.ErrorAs(t, err, &errDuplicatedKey, "should fail with ErrDuplicateKey")

			_, err = repo.GetByID(t.Context(), recipe.ID.String())

			var errNotFound *common.NotFoundError

			require.ErrorAs(t, err, &errNotFound, "should not have created the recipe")
		})
		t.Run("returns a ConstrainError if any of the ingredients does not exist", func(t *testing.T) {
			recipe := recipes.Recipe{}
			gofakeit.Struct(&recipe)

			recipe.Ingredients[0].ID = uuid.New()
			_, err := repo.Create(t.Context(), recipe)

			var errConstrain *common.ConstrainError

			require.ErrorAs(t, err, &errConstrain, "should fail with ConstrainError")
		})
		t.Run("returns a DuplicateError if any of the ingredients is duplicated", func(t *testing.T) {
			recipe := recipes.Recipe{}
			gofakeit.Struct(&recipe)

			recipe.Ingredients = append(recipe.Ingredients, recipe.Ingredients[0])
			_, err := repo.Create(t.Context(), recipe)

			var errConstrain *common.ConstrainError

			require.ErrorAs(t, err, &errConstrain, "should fail with ConstrainError")
		})
	})
	t.Run("Update", func(t *testing.T) {
		t.Run("update the Recipe and returns the updated version with ingredients", func(t *testing.T) {
			base := recipeFixtures[1]
			wanted := new(recipes.Recipe)
			gofakeit.Struct(wanted)

			wanted.ID = base.ID

			for j, i := range base.Ingredients {
				wanted.Ingredients[j] = recipes.RecipeIngredient{
					ID:       i.ID,
					Name:     i.Name,
					Kind:     i.Kind,
					Quantity: gofakeit.Float64Range(0.1, 200),
					Unit:     recipes.Unit(recipes.Units[gofakeit.Number(0, len(recipes.Units)-1)]),
				}
			}

			got, err := repo.Update(t.Context(), *wanted)
			require.NoError(t, err, "error should be nil")

			require.Equal(t, wanted.ID, got.ID, "recipe should have the same ID")
			require.WithinDuration(t, wanted.CreatedAt, got.CreatedAt, time.Second, "recipe should have a CreatedAt")
			require.WithinDuration(t, got.UpdatedAt, time.Now(), time.Second, "recipe should have a UpdatedAt")
			require.Equal(t, wanted.Title, got.Title, "recipe should have the same title")
			require.Equal(t, *wanted.Headline, *got.Headline, "recipe should have the same headline")
			require.Equal(t, *wanted.Description, *got.Description, "recipe should have the same description")
			require.Equal(t, *wanted.URL, *got.URL, "recipe should have the same URL")
			require.Equal(t, wanted.Tags, got.Tags, "recipe should have the same tags")
			require.Equal(t, *wanted.Steps, *got.Steps, "recipe should have the same steps")
			require.Equal(t, wanted.Servings, got.Servings, "recipe should have the same servings")
			require.Equal(t, *wanted.PrepTime, *got.PrepTime, "recipe should have the same prep time")
			require.Equal(t, wanted.Slug, got.Slug, "recipe should have the same slug")
			require.Len(t, got.Ingredients, len(wanted.Ingredients), "recipe should have the same number of ingredients")

			for i, ingredient := range wanted.Ingredients {
				require.NotZero(t, got.Ingredients[i].ID, "ingredient should have an ID")
				require.Equal(t, ingredient.Name, got.Ingredients[i].Name, "ingredient should have the same name")
				require.InEpsilon(t, ingredient.Quantity, got.Ingredients[i].Quantity, 0.001, "ingredient should have the same quantity")
				require.Equal(t, ingredient.Unit, got.Ingredients[i].Unit, "ingredient should have the same unit")
			}

			gotInDB, err := repo.GetByID(t.Context(), wanted.ID.String())
			require.NoError(t, err, "error should be nil")

			require.Equal(t, *got, *gotInDB, "recipe should be updated in the db")
		})
		t.Run("returns a DuplicateError when unique fields are duplicated", func(t *testing.T) {
			recipe := new(recipes.Recipe)
			gofakeit.Struct(recipe)
			recipe.ID = recipeFixtures[1].ID
			recipe.Slug = recipeFixtures[0].Slug

			_, err := repo.Update(t.Context(), *recipe)

			var errDuplicate *common.DuplicateError

			require.ErrorAs(t, err, &errDuplicate, "error should be DuplicateError")
		})
		t.Run("returns a ConstrainError if any ingredient does not exist", func(t *testing.T) {
			recipe := new(recipes.Recipe)
			gofakeit.Struct(recipe)
			recipe.ID = recipeFixtures[1].ID

			recipe.Ingredients[0].ID = uuid.New()
			_, err := repo.Update(t.Context(), *recipe)

			var errConstrain *common.ConstrainError

			require.ErrorAs(t, err, &errConstrain, "should fail with ConstrainError")
		})
		t.Run("returns ErrNotFound if recipe does not exist", func(t *testing.T) {
			recipe := new(recipes.Recipe)
			gofakeit.Struct(recipe)

			_, err := repo.Update(t.Context(), *recipe)

			var errNotFound *common.NotFoundError

			require.ErrorAs(t, err, &errNotFound, "error should be ErrNotFound")
		})
	})
	t.Run("Delete", func(t *testing.T) {
		t.Run("deletes a recipe", func(t *testing.T) {
			err := repo.Delete(t.Context(), recipeFixtures[9].ID.String())

			require.NoError(t, err, "error should be nil")

			_, err = repo.GetByID(t.Context(), recipeFixtures[9].ID.String())

			var errNotFound *common.NotFoundError

			require.ErrorAs(t, err, &errNotFound, "should not find the recipe after deleting")
		})
		t.Run("returns no error if recipe does not exists", func(t *testing.T) {
			err := repo.Delete(t.Context(), uuid.NewString())

			require.NoError(t, err, "error should be nil")
		})
	})
	t.Run("GetUniqueSlug", func(t *testing.T) {
		t.Run("returs default slug if empty", func(t *testing.T) {
			title := gofakeit.Adjective() + " " + gofakeit.Adjective() + " " + gofakeit.Dinner()
			wanted := slug.Make(title)

			got, err := repo.GetUniqueSlug(t.Context(), title)
			require.NoError(t, err, "error should be nil")
			require.Equal(t, wanted, got, "should return the same slug")
		})
		t.Run("returns next available slug", func(t *testing.T) {
			title := gofakeit.Adjective() + " " + gofakeit.Adjective() + " " + gofakeit.Dinner()
			slug := slug.Make(title)

			testRecipes := []sql.CreateRecipeParams{
				{ID: uuid.New().String(), Title: title, Slug: slug},
				{ID: uuid.New().String(), Title: title, Slug: slug + "-1"},
				{ID: uuid.New().String(), Title: title, Slug: slug + "-2"},
				{ID: uuid.New().String(), Title: title, Slug: slug + "-3"},
				{ID: uuid.New().String(), Title: title, Slug: slug + "-5"},
				{ID: uuid.New().String(), Title: title, Slug: slug + "-vegan"},
			}
			queries := sql.New()

			for _, r := range testRecipes {
				_, err := queries.CreateRecipe(t.Context(),
					db, r)

				require.NoError(t, err, "error inserting test data")
			}

			wanted := slug + "-4"
			got, err := repo.GetUniqueSlug(t.Context(), title)
			require.NoError(t, err, "error should be nil")
			require.Equal(t, wanted, got, "should return the next available slug")
		})
	})
}
