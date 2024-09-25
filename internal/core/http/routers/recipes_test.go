package routers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/AlejandroHerr/cook-book-go/internal/common/api"
	"github.com/AlejandroHerr/cook-book-go/internal/common/infra/db"
	"github.com/AlejandroHerr/cook-book-go/internal/common/logging"
	"github.com/AlejandroHerr/cook-book-go/internal/core/dtos"
	"github.com/AlejandroHerr/cook-book-go/internal/core/http/routers"
	"github.com/AlejandroHerr/cook-book-go/internal/core/model"
	"github.com/AlejandroHerr/cook-book-go/internal/core/repo"
	"github.com/AlejandroHerr/cook-book-go/internal/core/testutil"
	"github.com/AlejandroHerr/cook-book-go/internal/core/usecases"
	ctu "github.com/AlejandroHerr/cook-book-go/internal/testutil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/stretchr/testify/require"
)

type testServer struct {
	server          *httptest.Server
	cleanup         func()
	recipesRepo     *repo.PgRecipesRepo
	ingredientsRepo *repo.PgIngredientsRepo
}

func setupTestServer(t *testing.T) *testServer {
	t.Helper()

	// Initialize repositories
	recipesRepo := repo.NewPgRecipesRepository(pgPool)
	ingredientsRepo := repo.NewPgIngredientsRepo(pgPool)

	// Initialize transaction manager
	transactionManager := db.NewPgxTransactionManager(pgPool)

	// Initialize VoidLogger
	logger := logging.NewVoidLogger()

	// Initialize services
	useCases := usecases.NewRecipesUseCases(transactionManager, recipesRepo, ingredientsRepo, logger)

	// Initialize handlers
	recipesRouter := routers.NewRecipesRouter(useCases)

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
				require.NoError(t, err)
				defer resp.Body.Close()

				require.Equal(t, http.StatusOK, resp.StatusCode, "should return status ok")

				var fetched routers.GetRecipesResponse
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err, "shuld be a GetRecipesResponse")

				dbRecipes, err := ts.recipesRepo.GetAll(context.Background())
				require.NoError(t, err)

				expected := routers.MakeGetRecipesResponse(dbRecipes)

				require.Equal(t, *expected, fetched, "should be the recipes in the db")
			})
		})

		t.Run("POST /", func(t *testing.T) {
			t.Run("creates a new recipe", func(t *testing.T) {
				var recipeDTO dtos.CreateUpdateRecipeDTO

				ctu.MustMakeStructFixture(&recipeDTO)

				existingIngredient := fixtures.Ingredients[0]
				newIngredientName := "New Ingredient"

				recipeDTO.Ingredients = []dtos.CreateRecipeIngredient{
					{
						Quantity: 1.0,
						Unit:     model.Gram,
						Name:     existingIngredient.Name,
					},
					{
						Quantity: 5.5,
						Unit:     model.Liter,
						Name:     newIngredientName,
					},
				}

				fmt.Println("recipeDTO", recipeDTO)

				jsonBody, err := json.Marshal(recipeDTO)
				require.NoError(t, err, "error marshaling recipeDTO")

				resp, err := http.Post(
					ts.server.URL+"/recipes/",
					"application/json",
					bytes.NewBuffer(jsonBody),
				)

				require.NoError(t, err, "should not return an error")

				defer resp.Body.Close()

				require.Equal(t, http.StatusCreated, resp.StatusCode, "should return status created")

				var fetched routers.CreateRecipeResponse
				err = json.NewDecoder(resp.Body).Decode(&fetched)

				require.NoError(t, err, "response should be the created recipe")
				require.Equal(t, recipeDTO.Title, fetched.Title, "title should be equal")
				require.Equal(t, recipeDTO.Headline, *fetched.Headline, "headline should be equal")
				require.Equal(t, recipeDTO.Description, *fetched.Description, "description should be equal")
				require.Equal(t, recipeDTO.Steps, *fetched.Steps, "steps should be equal")
				require.Equal(t, recipeDTO.Servings, *fetched.Servings, "servings should be equal")
				require.Equal(t, recipeDTO.URL, *fetched.URL, "url should be equal")
				require.Equal(t, recipeDTO.Tags, fetched.Tags, "tags should be equal")
				require.Equal(t, len(recipeDTO.Ingredients), len(fetched.Ingredients), "ingredients should have equal length")
				require.Equal(t, model.RecipeIngredient{
					ID:       existingIngredient.ID,
					Quantity: recipeDTO.Ingredients[0].Quantity,
					Unit:     recipeDTO.Ingredients[0].Unit,
					Name:     recipeDTO.Ingredients[0].Name,
					Kind:     existingIngredient.Kind,
				}, fetched.Ingredients[0], "existing ingredient should be equal")
				require.Equal(t, recipeDTO.Ingredients[1].Name, fetched.Ingredients[1].Name, "new ingredient name should be equal")
				require.Equal(t, recipeDTO.Ingredients[1].Quantity, fetched.Ingredients[1].Quantity, "new ingredient quantity should be equal")
				require.Equal(t, recipeDTO.Ingredients[1].Unit, fetched.Ingredients[1].Unit, "new ingredient unit should be equal")
				require.NotEqual(t, uuid.Nil, fetched.Ingredients[1].ID, "new ingredient id should not be nil")
				require.Nil(t, fetched.Ingredients[1].Kind, "new ingredient kind should be nil")
				require.True(t, fetched.CreatedAt.Before(time.Now()), "created at should be before now")

				dbRecipe, err := ts.recipesRepo.GetByID(context.Background(), fetched.ID.String())
				require.NoError(t, err, "new recipe should be in the db")

				expected := model.Recipe{
					ID:          fetched.ID,
					Title:       fetched.Title,
					Headline:    fetched.Headline,
					CreatedAt:   fetched.CreatedAt,
					UpdatedAt:   fetched.UpdatedAt,
					Description: fetched.Description,
					Steps:       fetched.Steps,
					Servings:    fetched.Servings,
					URL:         fetched.URL,
					Tags:        fetched.Tags,
					Ingredients: fetched.Ingredients,
				}
				require.Equal(t, expected, *dbRecipe, "should be the recipe in the db")
			})
			t.Run("return a Bad Request Status if data is invalid", func(t *testing.T) {
				var recipeDTO dtos.CreateUpdateRecipeDTO

				ctu.MustMakeStructFixture(&recipeDTO)

				recipeDTO.Title = ""

				jsonBody, err := json.Marshal(recipeDTO)
				require.NoError(t, err, "error marshaling recipeDTO")

				resp, err := http.Post(
					ts.server.URL+"/recipes/",
					"application/json",
					bytes.NewBuffer(jsonBody),
				)
				require.NoError(t, err, "should not return an error")

				defer resp.Body.Close()

				require.Equal(t, http.StatusBadRequest, resp.StatusCode, "should return status confilt")

				var fetched api.ErrValidationResponse
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err, "response should be an ErrResponse")

				require.Equal(t, fetched.ErrorText, "Invalid payload", "error text should be for an invalid request")
				require.Len(t, fetched.Details, 1, "should contain validation error details")

				dbRecipe, err := ts.recipesRepo.GetBySlug(context.Background(), slug.Make(recipeDTO.Title))
				require.Error(t, err, "should return an error")
				require.Nil(t, dbRecipe, "should not create a recipe")
			})
			t.Run("return a Conflict Status if recipe title already exists", func(t *testing.T) {
				var recipeDTO dtos.CreateUpdateRecipeDTO

				ctu.MustMakeStructFixture(&recipeDTO)

				recipeDTO.Title = fixtures.Recipes[0].Title

				jsonBody, err := json.Marshal(recipeDTO)
				require.NoError(t, err, "error marshaling recipeDTO")

				resp, err := http.Post(
					ts.server.URL+"/recipes/",
					"application/json",
					bytes.NewBuffer(jsonBody),
				)
				require.NoError(t, err, "should not return an error")
				defer resp.Body.Close()

				require.Equal(t, http.StatusConflict, resp.StatusCode, "should return status confilt")

				var fetched api.ErrResponse
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err, "response should be an ErrResponse")
			})
		})

		t.Run("GET /{slug}", func(t *testing.T) {
			t.Run("returns a recipe by the slug", func(t *testing.T) {
				resp, err := http.Get(
					ts.server.URL + "/recipes/" + fixtures.Recipes[0].Slug(),
				)
				require.NoError(t, err)
				defer resp.Body.Close()

				require.Equal(t, http.StatusOK, resp.StatusCode)

				var fetched model.Recipe
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err)

				testutil.RequireRecipeEquals(t, *fixtures.Recipes[0], fetched, false, false, "should be the recipe in the db")
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
			t.Run("returns the recpe by id", func(t *testing.T) {
				resp, err := http.Get(
					ts.server.URL + "/recipes/" + fixtures.Recipes[0].ID.String(),
				)
				require.NoError(t, err)
				defer resp.Body.Close()

				require.Equal(t, http.StatusOK, resp.StatusCode)

				var fetched model.Recipe
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err)

				sort.Slice(fixtures.Recipes[0].Ingredients, func(i, j int) bool {
					return fixtures.Recipes[0].Ingredients[i].Name < fixtures.Recipes[0].Ingredients[j].Name
				})
				sort.Slice(fetched.Ingredients, func(i, j int) bool {
					return fetched.Ingredients[i].Name < fetched.Ingredients[j].Name
				})
				require.Equal(t, *fixtures.Recipes[0], fetched, "should be the recipe in the db")
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
		t.Run("PUT /", func(t *testing.T) {
			t.Run("updatets an existing recipe", func(t *testing.T) {
				recipeID := fixtures.Recipes[3].ID

				var recipeDTO dtos.CreateUpdateRecipeDTO

				ctu.MustMakeStructFixture(&recipeDTO)

				recipeDTO.Ingredients = []dtos.CreateRecipeIngredient{
					{
						Quantity: 1.7,
						Unit:     model.Tablespoon,
						Name:     fixtures.Ingredients[0].Name,
					},
					{
						Quantity: 5.5,
						Unit:     model.Liter,
						Name:     "New Ingredient for Updating",
					},
				}

				jsonBody, err := json.Marshal(recipeDTO)
				require.NoError(t, err, "error marshaling recipeDTO")

				req, err := http.NewRequest(
					http.MethodPut,
					ts.server.URL+"/recipes/"+recipeID.String(),
					bytes.NewBuffer(jsonBody),
				)
				require.NoError(t, err, "error creating request")

				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err, "error making request")
				defer resp.Body.Close()

				require.Equal(t, http.StatusOK, resp.StatusCode, "should return status ok")

				var fetched routers.CreateRecipeResponse
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err, "response should be a CreateRecipeResponse")

				expected := model.Recipe{ //nolint:exhaustruct
					ID:          recipeID,
					Title:       recipeDTO.Title,
					Headline:    &recipeDTO.Headline,
					Description: &recipeDTO.Description,
					Steps:       &recipeDTO.Steps,
					Servings:    &recipeDTO.Servings,
					URL:         &recipeDTO.URL,
					Tags:        recipeDTO.Tags,
					Ingredients: []model.RecipeIngredient{
						{
							ID:       fixtures.Ingredients[0].ID,
							Quantity: recipeDTO.Ingredients[0].Quantity,
							Unit:     recipeDTO.Ingredients[0].Unit,
							Name:     recipeDTO.Ingredients[0].Name,
							Kind:     fixtures.Ingredients[0].Kind,
						},
						{
							ID:       uuid.Nil,
							Quantity: recipeDTO.Ingredients[1].Quantity,
							Unit:     recipeDTO.Ingredients[1].Unit,
							Name:     recipeDTO.Ingredients[1].Name,
							Kind:     nil,
						},
					},
				}

				require.Equal(t, expected.ID, fetched.ID, "fetched ID should be equal")
				require.Equal(t, expected.Title, fetched.Title, "fetched title should be equal")
				require.Equal(t, *expected.Headline, *fetched.Headline, "fetched headline should be equal")
				require.Equal(t, *expected.Description, *fetched.Description, "fetched description should be equal")
				require.Equal(t, *expected.Steps, *fetched.Steps, "fetched steps should be equal")
				require.Equal(t, *expected.Servings, *fetched.Servings, "fetched servings should be equal")
				require.Equal(t, *expected.URL, *fetched.URL, "fetched url should be equal")
				require.Equal(t, expected.Tags, fetched.Tags, "fetched tags should be equal")
				require.Equal(t, len(expected.Ingredients), len(fetched.Ingredients), "fetched ingredients should have equal length")
				require.Equal(t, expected.Ingredients[0], fetched.Ingredients[0], "fetched existing ingredient ID should be equal")
				require.Equal(t, expected.Ingredients[1].Name, fetched.Ingredients[1].Name, "fetched new ingredient name should be equal")
				require.Equal(t, expected.Ingredients[1].Quantity, fetched.Ingredients[1].Quantity, "fetched new ingredient quantity should be equal")
				require.Equal(t, expected.Ingredients[1].Unit, fetched.Ingredients[1].Unit, "fetched new ingredient unit should be equal")
				require.NotEqual(t, uuid.Nil, fetched.Ingredients[1].ID, "fetched new ingredient id should not be nil")
				require.Nil(t, fetched.Ingredients[1].Kind, "fetched new ingredient kind should be nil")
				require.True(t, fetched.CreatedAt.Before(time.Now()), "fetched created at should be before now")

				dbRecipe, err := ts.recipesRepo.GetByID(context.Background(), recipeID.String())
				require.NoError(t, err, "new recipe should be in the db")

				require.Equal(t, model.Recipe{
					ID:          recipeID,
					Title:       recipeDTO.Title,
					Headline:    &recipeDTO.Headline,
					CreatedAt:   fetched.CreatedAt,
					UpdatedAt:   fetched.UpdatedAt,
					Description: &recipeDTO.Description,
					Steps:       &recipeDTO.Steps,
					Servings:    &recipeDTO.Servings,
					URL:         &recipeDTO.URL,
					Tags:        recipeDTO.Tags,
					Ingredients: fetched.Ingredients,
				}, *dbRecipe, "should be the recipe in the db")
			})
			t.Run("return a Bad Request Status if data is invalid", func(t *testing.T) {
				recipeID := fixtures.Recipes[4].ID

				var recipeDTO dtos.CreateUpdateRecipeDTO

				ctu.MustMakeStructFixture(&recipeDTO)

				recipeDTO.Title = ""

				jsonBody, err := json.Marshal(recipeDTO)
				require.NoError(t, err, "error marshaling recipeDTO")

				req, err := http.NewRequest(
					http.MethodPut,
					ts.server.URL+"/recipes/"+recipeID.String(),
					bytes.NewBuffer(jsonBody),
				)
				require.NoError(t, err, "error creating request")

				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err, "error making request")
				defer resp.Body.Close()

				require.Equal(t, http.StatusBadRequest, resp.StatusCode, "should return status confilt")

				var fetched api.ErrValidationResponse
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err, "response should be an ErrResponse")

				require.Equal(t, fetched.ErrorText, "Invalid payload", "error text should be for an invalid request")
				require.Len(t, fetched.Details, 1, "should contain validation error details")

				dbRecipe, err := ts.recipesRepo.GetByID(context.Background(), recipeID.String())
				require.NoError(t, err)
				testutil.RequireRecipeEquals(t, *fixtures.Recipes[4], *dbRecipe, false, false, "should not update the recipe in the db")
			})
			t.Run("return an error if recipe title already exists", func(t *testing.T) {
				recipeID := fixtures.Recipes[5].ID

				var recipeDTO dtos.CreateUpdateRecipeDTO

				ctu.MustMakeStructFixture(&recipeDTO)

				recipeDTO.Title = fixtures.Recipes[0].Title

				jsonBody, err := json.Marshal(recipeDTO)
				require.NoError(t, err, "error marshaling recipeDTO")

				req, err := http.NewRequest(
					http.MethodPut,
					ts.server.URL+"/recipes/"+recipeID.String(),
					bytes.NewBuffer(jsonBody),
				)
				require.NoError(t, err, "error creating request")

				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err, "error making request")
				defer resp.Body.Close()

				require.Equal(t, http.StatusConflict, resp.StatusCode, "should return status confilt")

				var fetched api.ErrResponse
				err = json.NewDecoder(resp.Body).Decode(&fetched)
				require.NoError(t, err, "response should be an ErrResponse")

				dbRecipe, err := ts.recipesRepo.GetByID(context.Background(), recipeID.String())
				require.NoError(t, err)
				testutil.RequireRecipeEquals(t, *fixtures.Recipes[5], *dbRecipe, false, false, "should not update the recipe in the db")
			})
		})
		t.Run("DELETE /{id}", func(t *testing.T) {
			t.Run("deletes the recpe by id", func(t *testing.T) {
				req, err := http.NewRequest(http.MethodDelete, ts.server.URL+"/recipes/"+fixtures.Recipes[2].ID.String(), nil)
				require.NoError(t, err)

				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()

				require.Equal(t, http.StatusNoContent, resp.StatusCode)

				_, err = ts.recipesRepo.GetByID(context.Background(), fixtures.Recipes[2].ID.String())
				require.Error(t, err, "should return an error")
			})
			t.Run("returns a Not Found Error when the recipe does not exist", func(t *testing.T) {
				req, err := http.NewRequest(http.MethodDelete, ts.server.URL+"/recipes/"+uuid.New().String(), nil)
				require.NoError(t, err)

				resp, err := http.DefaultClient.Do(req)
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
	})
}
