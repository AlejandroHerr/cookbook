//nolint:paralleltest
package e2e_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AlejandroHerr/cookbook/internal/common"
	"github.com/AlejandroHerr/cookbook/internal/common/api"
	"github.com/AlejandroHerr/cookbook/internal/common/logger"
	"github.com/AlejandroHerr/cookbook/internal/common/pg"
	pgtestutil "github.com/AlejandroHerr/cookbook/internal/common/pg/testutil"
	"github.com/AlejandroHerr/cookbook/internal/recipes"
	pgrecipes "github.com/AlejandroHerr/cookbook/internal/recipes/pg"
	"github.com/AlejandroHerr/cookbook/internal/recipes/pg/fixtures"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type recipesTestSuite struct {
	db          *pgxpool.Pool
	server      *httptest.Server
	fixtures    []recipes.Recipe
	recipesRepo *pgrecipes.RecipesRepo
}

func TestRecipes_E2E(t *testing.T) {
	suite := setup(t)

	t.Run("GET /", func(t *testing.T) {
		t.Run("returns a list will al the recipes without ingredients", func(t *testing.T) {
			resp, err := http.Get(suite.server.URL)
			require.NoError(t, err, "failed to make GET request")

			defer resp.Body.Close()

			require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status code")

			var got recipes.RecipesListResponse
			err = json.NewDecoder(resp.Body).Decode(&got)
			require.NoError(t, err, "response should be a list of recipes")

			require.Len(t, got.Recipes, len(suite.fixtures), "should return all recipes")

			expected, err := suite.recipesRepo.List(t.Context())
			require.NoError(t, err, "error should be nil")

			require.Equal(t, expected, got.Recipes, "should return the same recipes")
		})
	})

	t.Run("POST /", func(t *testing.T) {
		t.Run("creates and returns a new Recipe", func(t *testing.T) {
			var dto recipes.CreateUpdateRecipeDTO

			gofakeit.Struct(&dto)

			jsonBody, err := json.Marshal(dto)
			require.NoError(t, err, "failed to marshal createUpdateRecipeDTO")

			resp, err := http.Post(suite.server.URL, "application/json", bytes.NewBuffer(jsonBody))
			require.NoError(t, err, "failed to make POST request")

			defer resp.Body.Close()

			require.Equal(t, http.StatusCreated, resp.StatusCode, "unexpected status code")

			var got recipes.Recipe
			err = json.NewDecoder(resp.Body).Decode(&got)
			require.NoError(t, err, "response should be a recipe")

			require.NotZero(t, got.ID, "recipe should have an ID")
			require.WithinDuration(t, got.CreatedAt, time.Now(), time.Second, "recipe should have a CreatedAt")
			require.WithinDuration(t, got.UpdatedAt, time.Now(), time.Second, "recipe should have a UpdatedAt")
			require.Equal(t, dto.Title, got.Title, "recipe should have the same title")
			require.Equal(t, dto.Headline, *got.Headline, "recipe should have the same headline")
			require.Equal(t, dto.Description, *got.Description, "recipe should have the same description")
			require.Equal(t, dto.URL, *got.URL, "recipe should have the same URL")
			require.Equal(t, dto.Tags, got.Tags, "recipe should have the same tags")
			require.Equal(t, dto.Steps, *got.Steps, "recipe should have the same steps")
			require.Equal(t, dto.Servings, got.Servings, "recipe should have the same servings")
			require.Equal(t, dto.PrepTime, *got.PrepTime, "recipe should have the same prep time")
			require.Equal(t, slug.Make(dto.Title), got.Slug, "recipe should have the same slug")
			require.Len(t, got.Ingredients, len(dto.Ingredients), "recipe should have the same number of ingredients")

			for i, ingredient := range dto.Ingredients {
				require.NotZero(t, got.Ingredients[i].ID, "ingredient should have an ID")
				require.Equal(t, ingredient.Name, got.Ingredients[i].Name, "ingredient should have the same name")
				require.InEpsilon(t, ingredient.Quantity, got.Ingredients[i].Quantity, 0.001, "ingredient should have the same quantity")
				require.Equal(t, ingredient.Unit, got.Ingredients[i].Unit, "ingredient should have the same unit")
			}

			gotInDB, err := suite.recipesRepo.GetByID(t.Context(), got.ID.String())
			require.NoError(t, err, "failed to get recipe by ID")

			require.Equal(t, got, *gotInDB, "recipe should be in the db")
		})
		t.Run("returns bad request when data is invalid", func(t *testing.T) {
			var createUpdateRecipeDTO recipes.CreateUpdateRecipeDTO

			gofakeit.Struct(&createUpdateRecipeDTO)

			createUpdateRecipeDTO.Title = ""

			jsonBody, err := json.Marshal(createUpdateRecipeDTO)
			require.NoError(t, err, "failed to marshal createUpdateRecipeDTO")

			resp, err := http.Post(suite.server.URL, "application/json", bytes.NewBuffer(jsonBody))
			require.NoError(t, err, "failed to make POST request")

			defer resp.Body.Close()

			require.Equal(t, http.StatusBadRequest, resp.StatusCode, "unexpected status code")

			var got api.ErrorResponse
			err = json.NewDecoder(resp.Body).Decode(&got)
			require.NoError(t, err, "response should be an error response")
		})
	})

	t.Run("GET /{IDSlug}", func(t *testing.T) {
		getTestCase := []struct {
			descr    string
			param    string
			expected recipes.Recipe
		}{{
			descr:    "GET /{id}",
			param:    suite.fixtures[1].ID.String(),
			expected: suite.fixtures[1],
		}, {
			descr:    "GET /{slug}",
			param:    suite.fixtures[2].Slug,
			expected: suite.fixtures[2],
		}}

		for _, tc := range getTestCase {
			resp, err := http.Get(suite.server.URL + "/" + tc.param)
			require.NoError(t, err, "failed to make GET request")

			defer resp.Body.Close()

			require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status code")

			var got recipes.Recipe
			err = json.NewDecoder(resp.Body).Decode(&got)
			require.NoError(t, err, "response should be a recipe")

			require.Equal(t, tc.expected, got, "recipe should be equal for "+tc.param)
		}

		t.Run("returns not found when the recipe id does not exist", func(t *testing.T) {
			resp, err := http.Get(suite.server.URL + "/" + uuid.NewString())
			require.NoError(t, err, "failed to make GET request")

			defer resp.Body.Close()

			require.Equal(t, http.StatusNotFound, resp.StatusCode, "unexpected status code")

			var got api.ErrorResponse
			err = json.NewDecoder(resp.Body).Decode(&got)
			require.NoError(t, err, "response should be an error response")
		})
		t.Run("returns not found when the recipe slug does not exist", func(t *testing.T) {
			resp, err := http.Get(suite.server.URL + "/" + slug.Make(gofakeit.Sentence(5)))
			require.NoError(t, err, "failed to make GET request")

			defer resp.Body.Close()

			require.Equal(t, http.StatusNotFound, resp.StatusCode, "unexpected status code")

			var got api.ErrorResponse
			err = json.NewDecoder(resp.Body).Decode(&got)
			require.NoError(t, err, "response should be an error response")
		})
	})

	t.Run("PUT /{ID}", func(t *testing.T) {
		t.Run("deletes the recipe by ID", func(t *testing.T) {
			recipeToUpdate := suite.fixtures[3]

			var dto recipes.CreateUpdateRecipeDTO

			gofakeit.Struct(&dto)

			jsonBody, err := json.Marshal(dto)
			require.NoError(t, err, "failed to marshal createUpdateRecipeDTO")

			req, err := http.NewRequest(http.MethodPut, suite.server.URL+"/"+recipeToUpdate.ID.String(), bytes.NewBuffer(jsonBody))
			require.NoError(t, err, "failed to create PUT request")

			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err, "failed to make PUT request")

			defer resp.Body.Close()

			require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status code")

			var got recipes.Recipe
			err = json.NewDecoder(resp.Body).Decode(&got)
			require.NoError(t, err, "response should be a recipe")

			require.Equal(t, recipeToUpdate.ID, got.ID, "recipe should have the same ID")
			require.Equal(t, recipeToUpdate.CreatedAt, got.CreatedAt, "recipe should have a CreatedAt")
			require.WithinDuration(t, got.UpdatedAt, time.Now(), time.Second, "recipe should have a UpdatedAt")
			require.Equal(t, dto.Title, got.Title, "recipe should have the same title")
			require.Equal(t, dto.Headline, *got.Headline, "recipe should have the same headline")
			require.Equal(t, dto.Description, *got.Description, "recipe should have the same description")
			require.Equal(t, dto.URL, *got.URL, "recipe should have the same URL")
			require.Equal(t, dto.Tags, got.Tags, "recipe should have the same tags")
			require.Equal(t, dto.Steps, *got.Steps, "recipe should have the same steps")
			require.Equal(t, dto.Servings, got.Servings, "recipe should have the same servings")
			require.Equal(t, dto.PrepTime, *got.PrepTime, "recipe should have the same prep time")
			require.Equal(t, slug.Make(dto.Title), got.Slug, "recipe should have the same slug")
			require.Len(t, got.Ingredients, len(dto.Ingredients), "recipe should have the same ingredients")

			for i, ingredient := range dto.Ingredients {
				require.NotZero(t, got.Ingredients[i].ID, "ingredient should have an ID")
				require.Equal(t, ingredient.Name, got.Ingredients[i].Name, "ingredient should have the same name")
				require.InEpsilon(t, ingredient.Quantity, got.Ingredients[i].Quantity, 0.001, "ingredient should have the same quantity")
				require.Equal(t, ingredient.Unit, got.Ingredients[i].Unit, "ingredient should have the same unit")
			}

			gotInDB, err := suite.recipesRepo.GetByID(t.Context(), got.ID.String())
			require.NoError(t, err, "failed to get recipe by ID")

			require.Equal(t, got, *gotInDB, "recipe should be in the db")
		})

		t.Run("returns not found if the recipe does not exist", func(t *testing.T) {
			var createUpdateRecipeDTO recipes.CreateUpdateRecipeDTO

			gofakeit.Struct(&createUpdateRecipeDTO)

			jsonBody, err := json.Marshal(createUpdateRecipeDTO)
			require.NoError(t, err, "failed to marshal createUpdateRecipeDTO")

			req, err := http.NewRequest(http.MethodPut, suite.server.URL+"/"+uuid.NewString(), bytes.NewBuffer(jsonBody))
			require.NoError(t, err, "failed to create PUT request")

			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err, "failed to make PUT request")

			defer resp.Body.Close()

			require.Equal(t, http.StatusNotFound, resp.StatusCode, "unexpected status code")

			var got api.ErrorResponse
			err = json.NewDecoder(resp.Body).Decode(&got)
			require.NoError(t, err, "response should be an error response")
		})
		t.Run("returns bad request when data is invalid", func(t *testing.T) {
			recipeToUpdate := suite.fixtures[4]

			var createUpdateRecipeDTO recipes.CreateUpdateRecipeDTO

			createUpdateRecipeDTO.Title = ""

			jsonBody, err := json.Marshal(createUpdateRecipeDTO)
			require.NoError(t, err, "failed to marshal createUpdateRecipeDTO")

			req, err := http.NewRequest(http.MethodPut, suite.server.URL+"/"+recipeToUpdate.ID.String(), bytes.NewBuffer(jsonBody))
			require.NoError(t, err, "failed to create PUT request")

			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err, "failed to make PUT request")

			defer resp.Body.Close()

			require.Equal(t, http.StatusBadRequest, resp.StatusCode, "unexpected status code")

			var got api.ErrorResponse
			err = json.NewDecoder(resp.Body).Decode(&got)
			require.NoError(t, err, "response should be an error response")

			gotInDB, err := suite.recipesRepo.GetByID(t.Context(), recipeToUpdate.ID.String())
			require.NoError(t, err, "failed to get recipe by ID")

			require.Equal(t, recipeToUpdate, *gotInDB, "recipe should be in the db")
		})
	})
	t.Run("DELETE /{ID}", func(t *testing.T) {
		t.Run("deletes the recipe by ID", func(t *testing.T) {
			recipeToDelete := suite.fixtures[5]

			req, err := http.NewRequest(http.MethodDelete, suite.server.URL+"/"+recipeToDelete.ID.String(), nil)
			require.NoError(t, err, "failed to create DELETE request")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err, "failed to make DELETE request")

			defer resp.Body.Close()

			require.Equal(t, http.StatusNoContent, resp.StatusCode, "unexpected status code")

			_, err = suite.recipesRepo.GetByID(t.Context(), recipeToDelete.ID.String())

			var errNotFound *common.NotFoundError

			require.ErrorAs(t, err, &errNotFound, "recipe should not exist")
		})

		t.Run("returns not found if the recipe does not exist", func(t *testing.T) {
			req, err := http.NewRequest(http.MethodDelete, suite.server.URL+"/"+uuid.NewString(), nil)
			require.NoError(t, err, "failed to create DELETE request")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err, "failed to make DELETE request")

			defer resp.Body.Close()

			require.Equal(t, http.StatusNotFound, resp.StatusCode, "unexpected status code")

			var got api.ErrorResponse
			err = json.NewDecoder(resp.Body).Decode(&got)
			require.NoError(t, err, "response should be an error response")
		})
	})
}

func setup(t *testing.T) *recipesTestSuite {
	t.Helper()

	db := pgtestutil.MustConnect(t)
	recipeFixtures := fixtures.MustMakeFixtures(t, 100, db)

	logger := logger.NewTestLogger()

	transactionManager := pg.NewTransactionManager(db)

	ingredientsRepo := pgrecipes.NewIngredientsRepo(db)
	recipesRepo := pgrecipes.NewRecipesRepo(db)

	recipesService := recipes.NewService(transactionManager, recipesRepo, ingredientsRepo, logger)

	recipesRouter := recipes.NewRouter(recipesService, logger)

	server := httptest.NewServer(recipesRouter)

	t.Cleanup(server.Close)

	return &recipesTestSuite{
		db:          db,
		server:      server,
		fixtures:    recipeFixtures,
		recipesRepo: recipesRepo,
	}
}
