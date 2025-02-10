package recipes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/AlejandroHerr/cookbook/internal/common"
	"github.com/AlejandroHerr/cookbook/internal/common/api"
	"github.com/AlejandroHerr/cookbook/internal/common/infra/db"
	"github.com/AlejandroHerr/cookbook/internal/common/logging"
	"github.com/AlejandroHerr/cookbook/internal/common/testutil"
	"github.com/AlejandroHerr/cookbook/internal/recipes"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/stretchr/testify/require"
)

type testServer struct {
	server          *httptest.Server
	cleanup         func()
	recipesRepo     *recipes.PgRecipesRepo
	ingredientsRepo *recipes.PgIngredientsRepo
}

func setupTestServer(t *testing.T) *testServer {
	t.Helper()

	// Initialize repositories
	recipesRepo := recipes.MakePgRecipesRepository(pgPool)
	ingredientsRepo := recipes.MakePgIngredientsRepo(pgPool)

	// Initialize transaction manager
	transactionManager := db.MakePgxTransactionManager(pgPool)

	// Initialize VoidLogger
	logger := logging.NewVoidLogger()

	// Initialize services
	useCases := recipes.MakeUseCases(transactionManager, recipesRepo, ingredientsRepo, logger)

	// Initialize handlers
	recipesRouter := recipes.MakeRouter(useCases)

	// Initialize router
	r := chi.NewRouter()
	r.Mount("/recipes", recipesRouter)

	// Create test server
	// return hello worl from test server
	server := httptest.NewServer(r)

	cleanup := func() {
		server.Close()
	}

	return &testServer{
		server:          server,
		cleanup:         cleanup,
		recipesRepo:     recipesRepo,
		ingredientsRepo: ingredientsRepo,
	}
}

func TestRecipesRouter(t *testing.T) {
	ts := setupTestServer(t)
	t.Cleanup(ts.cleanup)

	t.Run("RecipesRouter", func(t *testing.T) {
		t.Parallel()
		t.Run("GET /", func(t *testing.T) {
			t.Run("return a list of all the recipes without ingredients", func(t *testing.T) {
				resp, err := http.Get(
					ts.server.URL + "/recipes/",
				)
				require.NoError(t, err, "request should not fail")
				defer resp.Body.Close()

				require.Equal(t, http.StatusOK, resp.StatusCode, "should return status ok")

				var fetched recipes.GetRecipesResponse
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err, "shuld be a GetRecipesResponse")

				dbRecipes, err := ts.recipesRepo.GetAll(context.Background())
				require.NoError(t, err)

				require.Equal(t, len(dbRecipes), len(fetched.Recipes), "should return the same number of recipes")

				for i := range fetched.Recipes {
					got := recipes.Recipe{
						ID:          fetched.Recipes[i].ID,
						Title:       fetched.Recipes[i].Title,
						Headline:    fetched.Recipes[i].Headline,
						Description: fetched.Recipes[i].Description,
						Steps:       fetched.Recipes[i].Steps,
						Servings:    fetched.Recipes[i].Servings,
						URL:         fetched.Recipes[i].URL,
						Tags:        fetched.Recipes[i].Tags,
						CreatedAt:   fetched.Recipes[i].CreatedAt,
						UpdatedAt:   fetched.Recipes[i].UpdatedAt,
						Ingredients: nil,
					}

					require.Equal(t, *dbRecipes[i], got, "should be the recipe "+strconv.Itoa(i)+" in the db")
				}
			})
		})

		t.Run("POST /", func(t *testing.T) {
			t.Run("creates a new recipe", func(t *testing.T) {
				var recipeDTO recipes.CreateUpdateRecipeDTO

				testutil.MustMakeStructFixture(&recipeDTO)

				jsonBody, err := json.Marshal(recipeDTO)
				require.NoError(t, err, "error marshaling recipeDTO")

				resp, err := http.Post(
					ts.server.URL+"/recipes/",
					"application/json",
					bytes.NewBuffer(jsonBody),
				)
				require.NoError(t, err, "request should not fail")
				defer resp.Body.Close()

				require.Equal(t, http.StatusCreated, resp.StatusCode, "should return status created")

				var fetched recipes.CreateRecipeResponse
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err, "response should be a CreateRecipeResponse")

				require.NotEqual(t, uuid.Nil, fetched.ID, "id should not be nil")
				require.Equal(t, recipeDTO.Title, fetched.Title, "title should be equal")
				require.Equal(t, recipeDTO.Headline, *fetched.Headline, "headline should be equal")
				require.Equal(t, recipeDTO.Description, *fetched.Description, "description should be equal")
				require.Equal(t, recipeDTO.Steps, *fetched.Steps, "steps should be equal")
				require.Equal(t, recipeDTO.Servings, *fetched.Servings, "servings should be equal")
				require.Equal(t, recipeDTO.URL, *fetched.URL, "url should be equal")
				require.Equal(t, recipeDTO.Tags, fetched.Tags, "tags should be equal")
				require.True(t, fetched.CreatedAt.Before(time.Now()), "created at should be before now")
				require.True(t, fetched.UpdatedAt.Before(time.Now()), "updated at should be before now")
				require.Equal(t, len(recipeDTO.Ingredients), len(fetched.Ingredients), "ingredients should have equal length")

				for i := range fetched.Ingredients {
					require.NotEqual(t, uuid.Nil, fetched.Ingredients[i].ID, "recipeIngredient id should not be nil")
					require.Equal(t, recipeDTO.Ingredients[i].Quantity, fetched.Ingredients[i].Quantity, "recipeIngredient quantity should be equal")
					require.Equal(t, recipeDTO.Ingredients[i].Unit, fetched.Ingredients[i].Unit, "recipeIngredient unit should be equal")
					require.Equal(t, recipeDTO.Ingredients[i].Name, fetched.Ingredients[i].Name, "recipeIngredient name should be equal")
					require.Nil(t, fetched.Ingredients[i].Kind, "recipeIngredient kind should be nil")
				}

				dbRecipe, err := ts.recipesRepo.GetBySlug(context.Background(), fetched.Slug)
				require.NoError(t, err, "new recipe should be in the db")

				RequireRecipeEqual(t, *fetched.Recipe, *dbRecipe, "fetched recipe should be equal to the one in the db")
			})
			t.Run("return a Bad Request Status if data is invalid", func(t *testing.T) {
				var recipeDTO recipes.CreateUpdateRecipeDTO

				testutil.MustMakeStructFixture(&recipeDTO)

				recipeDTO.URL = "wrong-updated-url"

				jsonBody, err := json.Marshal(recipeDTO)
				require.NoError(t, err, "error marshaling recipeDTO")

				resp, err := http.Post(
					ts.server.URL+"/recipes/",
					"application/json",
					bytes.NewBuffer(jsonBody),
				)
				require.NoError(t, err, "request should not fail")
				defer resp.Body.Close()

				require.Equal(t, http.StatusBadRequest, resp.StatusCode, "should return a Bad Request Status")

				var fetched api.ErrValidationResponse
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err, "response should be an ErrResponse")

				_, err = ts.recipesRepo.GetBySlug(context.Background(), slug.Make(recipeDTO.Title))

				var errNotFound *common.ErrNotFound

				require.ErrorAs(t, err, &errNotFound, "should not create a recipe")
			})
			t.Run("return a Conflict Status if recipe title already exists", func(t *testing.T) {
				var recipeDTO recipes.CreateUpdateRecipeDTO

				testutil.MustMakeStructFixture(&recipeDTO)

				recipeDTO.Title = fixtures[0].Title

				jsonBody, err := json.Marshal(recipeDTO)
				require.NoError(t, err, "error marshaling recipeDTO")

				resp, err := http.Post(
					ts.server.URL+"/recipes/",
					"application/json",
					bytes.NewBuffer(jsonBody),
				)
				require.NoError(t, err, "request should not fail")
				defer resp.Body.Close()

				require.Equal(t, http.StatusConflict, resp.StatusCode, "should return a Conflict Status")

				var fetched api.ErrResponse
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err, "response should be an ErrResponse")
			})
		})
		t.Run("GET /{slug}", func(t *testing.T) {
			t.Run("returns a recipe by the slug", func(t *testing.T) {
				recipeToFind := fixtures[0]
				resp, err := http.Get(
					ts.server.URL + "/recipes/" + recipeToFind.Slug(),
				)
				require.NoError(t, err, "request should not fail")

				defer resp.Body.Close()

				require.Equal(t, http.StatusOK, resp.StatusCode)

				var fetched recipes.GetRecipeResponse
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err, "response should be a GetRecipeResponse")

				dbRecipe, err := ts.recipesRepo.GetBySlug(context.Background(), recipeToFind.Slug())
				require.NoError(t, err, "should not return an error")

				require.Equal(t, *dbRecipe, *fetched.Recipe, "should be the recipe in the db")
			})
			t.Run("returns a Not Found Error when the recipe does not exist", func(t *testing.T) {
				resp, err := http.Get(
					ts.server.URL + "/recipes/" + "fake-slug",
				)
				require.NoError(t, err)
				defer resp.Body.Close()

				require.Equal(t, http.StatusNotFound, resp.StatusCode)

				var fetched api.ErrResponse
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err, "response should be an api.ErrResponse")

				require.Equal(t, http.StatusText(http.StatusNotFound), fetched.StatusText, "status text should inform about not found")
				require.Equal(t, "recipe not found", fetched.ErrorText, "error text should inform about not found")
			})
		})
		t.Run("GET /{id}", func(t *testing.T) {
			t.Run("returns a recipe by the ID", func(t *testing.T) {
				recipeToFind := fixtures[0]
				resp, err := http.Get(
					ts.server.URL + "/recipes/" + recipeToFind.ID.String(),
				)
				require.NoError(t, err, "request should not fail")

				defer resp.Body.Close()

				require.Equal(t, http.StatusOK, resp.StatusCode)

				var fetched recipes.GetRecipeResponse
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err, "response should be a GetRecipeResponse")

				dbRecipe, err := ts.recipesRepo.GetByID(context.Background(), recipeToFind.ID.String())
				require.NoError(t, err, "should not return an error")

				require.Equal(t, *dbRecipe, *fetched.Recipe, "should be the recipe in the db")
			})
			t.Run("returns a Not Found Error when the recipe does not exist", func(t *testing.T) {
				resp, err := http.Get(
					ts.server.URL + "/recipes/" + uuid.New().String(),
				)
				require.NoError(t, err)
				defer resp.Body.Close()

				require.Equal(t, http.StatusNotFound, resp.StatusCode)

				var fetched api.ErrResponse
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err, "response should be an api.ErrResponse")

				require.Equal(t, http.StatusText(http.StatusNotFound), fetched.StatusText, "status text should inform about not found")
				require.Equal(t, "recipe not found", fetched.ErrorText, "error text should inform about not found")
			})
		})
		t.Run("PUT /{id}", func(t *testing.T) {
			t.Run("updates a recipe", func(t *testing.T) {
				recipeToUpdate := fixtures[10]

				var recipeDTO recipes.CreateUpdateRecipeDTO

				testutil.MustMakeStructFixture(&recipeDTO)

				jsonBody, err := json.Marshal(recipeDTO)
				require.NoError(t, err, "error marshaling recipeDTO")

				req, err := http.NewRequest(http.MethodPut, ts.server.URL+"/recipes/"+recipeToUpdate.ID.String(), bytes.NewBuffer(jsonBody))
				require.NoError(t, err, "error creating request")

				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err, "request should not fail")
				defer resp.Body.Close()

				require.Equal(t, http.StatusOK, resp.StatusCode, "should return status ok")

				var fetched recipes.CreateRecipeResponse
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err, "response should be a CreateRecipeResponse")

				require.Equal(t, recipeToUpdate.ID, fetched.ID, "should be the same id")
				require.Equal(t, recipeDTO.Title, fetched.Title, "title should be equal")
				require.Equal(t, recipeDTO.Headline, *fetched.Headline, "headline should be equal")
				require.Equal(t, recipeDTO.Description, *fetched.Description, "description should be equal")
				require.Equal(t, recipeDTO.Steps, *fetched.Steps, "steps should be equal")
				require.Equal(t, recipeDTO.Servings, *fetched.Servings, "servings should be equal")
				require.Equal(t, recipeDTO.URL, *fetched.URL, "url should be equal")
				require.Equal(t, recipeDTO.Tags, fetched.Tags, "tags should be equal")
				require.Equal(t, recipeToUpdate.CreatedAt, fetched.CreatedAt, "created at should be before now")
				require.True(t, fetched.UpdatedAt.Before(time.Now()), "updated at should be before now")
				require.Equal(t, len(recipeDTO.Ingredients), len(fetched.Ingredients), "ingredients should have equal length")

				for i := range fetched.Ingredients {
					require.NotEqual(t, uuid.Nil, fetched.Ingredients[i].ID, "recipeIngredient id should not be nil")
					require.Equal(t, recipeDTO.Ingredients[i].Quantity, fetched.Ingredients[i].Quantity, "recipeIngredient quantity should be equal")
					require.Equal(t, recipeDTO.Ingredients[i].Unit, fetched.Ingredients[i].Unit, "recipeIngredient unit should be equal")
					require.Equal(t, recipeDTO.Ingredients[i].Name, fetched.Ingredients[i].Name, "recipeIngredient name should be equal")
					require.Nil(t, fetched.Ingredients[i].Kind, "recipeIngredient kind should be nil")
				}

				dbRecipe, err := ts.recipesRepo.GetBySlug(context.Background(), fetched.Slug)
				require.NoError(t, err, "updated recipe should be in the db")

				RequireRecipeEqual(t, *fetched.Recipe, *dbRecipe, "fetched recipe should be equal to the one in the db")
			})
			t.Run("return a Bad Request Status if data is invalid", func(t *testing.T) {
				recipeToUpdate := fixtures[10]

				var recipeDTO recipes.CreateUpdateRecipeDTO

				testutil.MustMakeStructFixture(&recipeDTO)

				recipeDTO.URL = "wrong-url"

				jsonBody, err := json.Marshal(recipeDTO)
				require.NoError(t, err, "error marshaling recipeDTO")

				req, err := http.NewRequest(http.MethodPut, ts.server.URL+"/recipes/"+recipeToUpdate.ID.String(), bytes.NewBuffer(jsonBody))
				require.NoError(t, err, "error creating request")

				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err, "request should not fail")
				defer resp.Body.Close()

				require.Equal(t, http.StatusBadRequest, resp.StatusCode, "should return a Bad Request Status")

				var fetched api.ErrValidationResponse
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err, "response should be an ErrResponse")

				_, err = ts.recipesRepo.GetBySlug(context.Background(), slug.Make(recipeDTO.Title))

				var errNotFound *common.ErrNotFound

				require.ErrorAs(t, err, &errNotFound, "should not create a recipe")
			})
			t.Run("return a Conflict Status if recipe title already exists", func(t *testing.T) {
				recipeToUpdate := fixtures[10]

				var recipeDTO recipes.CreateUpdateRecipeDTO

				testutil.MustMakeStructFixture(&recipeDTO)

				recipeDTO.Title = fixtures[0].Title

				jsonBody, err := json.Marshal(recipeDTO)
				require.NoError(t, err, "error marshaling recipeDTO")

				req, err := http.NewRequest(http.MethodPut, ts.server.URL+"/recipes/"+recipeToUpdate.ID.String(), bytes.NewBuffer(jsonBody))
				require.NoError(t, err, "error creating request")

				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err, "request should not fail")
				defer resp.Body.Close()

				require.Equal(t, http.StatusConflict, resp.StatusCode, "should return a Conflict Status")

				var fetched api.ErrResponse
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err, "response should be an ErrResponse")
			})
		})
		t.Run("DELETE /{id}", func(t *testing.T) {
			t.Run("deletes the recpe by id", func(t *testing.T) {
				recipeToDeleteID := fixtures[2].ID

				req, err := http.NewRequest(http.MethodDelete, ts.server.URL+"/recipes/"+recipeToDeleteID.String(), nil)
				require.NoError(t, err)

				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err, "request should not fail")
				defer resp.Body.Close()

				require.Equal(t, http.StatusNoContent, resp.StatusCode)

				_, err = ts.recipesRepo.GetByID(context.Background(), recipeToDeleteID.String())

				var errNotFound *common.ErrNotFound

				require.ErrorAs(t, err, &errNotFound, "should not find the recipe after deleting")
			})
			t.Run("returns a Not Found Error when the recipe does not exist", func(t *testing.T) {
				req, err := http.NewRequest(http.MethodDelete, ts.server.URL+"/recipes/"+uuid.New().String(), nil)
				require.NoError(t, err)

				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err, "request should not fail")

				defer resp.Body.Close()

				require.Equal(t, http.StatusNotFound, resp.StatusCode)

				var fetched api.ErrResponse
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err, "response should be an api.ErrResponse")

				require.Equal(t, http.StatusText(http.StatusNotFound), fetched.StatusText, "status text should inform about not found")
				require.Equal(t, "recipe not found", fetched.ErrorText, "error text should inform about not found")
			})
		})
	})
}
