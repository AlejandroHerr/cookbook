//nolint:exhaustruct
package recipes_test

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/AlejandroHerr/cookbook/internal/common/logger"
	commonmocks "github.com/AlejandroHerr/cookbook/internal/common/mocks"
	"github.com/AlejandroHerr/cookbook/internal/recipes"
	"github.com/AlejandroHerr/cookbook/internal/recipes/mocks"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type testServices struct {
	mockTxm             *commonmocks.TransactionManager
	mockRecipesRepo     *mocks.RecipesRepo
	mockIngredientsRepo *mocks.IngredientsRepo
	service             *recipes.Service
}

func newTestServices() *testServices {
	mockTxm := new(commonmocks.TransactionManager)
	mockRecipesRepo := new(mocks.RecipesRepo)
	mockIngredientsRepo := new(mocks.IngredientsRepo)

	service := recipes.NewService(mockTxm, mockRecipesRepo, mockIngredientsRepo, logger.NewTestLogger())

	return &testServices{
		mockTxm:             mockTxm,
		mockRecipesRepo:     mockRecipesRepo,
		mockIngredientsRepo: mockIngredientsRepo,
		service:             service,
	}
}

func TestRecipesService(t *testing.T) {
	t.Run("List", func(t *testing.T) {
		t.Parallel()
		t.Run("returns the recipes from the repo", func(t *testing.T) {
			services := newTestServices()
			recipes := []recipes.Recipe{{ID: uuid.New()}}
			services.mockRecipesRepo.On("List", mock.Anything).Return(recipes, nil)

			result, err := services.service.List(t.Context())

			require.NoError(t, err)
			require.Equal(t, recipes, result)
			services.mockIngredientsRepo.AssertExpectations(t)
		})
		t.Run("when the recipes repo fails it returns an error", func(t *testing.T) {
			services := newTestServices()
			repoErr := errors.New("repo error")
			services.mockRecipesRepo.On("List", mock.Anything).Return([]recipes.Recipe{}, repoErr)

			_, err := services.service.List(t.Context())

			require.Error(t, err)
			require.ErrorIs(t, err, repoErr)

			services.mockRecipesRepo.AssertNumberOfCalls(t, "List", 1)

			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
		})
	})
	t.Run("Create", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			dto  *recipes.CreateUpdateRecipeDTO
			slug string
		}{{
			dto: func() *recipes.CreateUpdateRecipeDTO {
				dto := &recipes.CreateUpdateRecipeDTO{}
				gofakeit.Struct(&dto)
				return dto
			}(),
			slug: slug.Make(gofakeit.Sentence(10)),
		}, {
			dto: func() *recipes.CreateUpdateRecipeDTO {
				dto := &recipes.CreateUpdateRecipeDTO{}
				gofakeit.Struct(&dto)
				dto.Servings = 0

				return dto
			}(),
			slug: slug.Make(gofakeit.Sentence(10)),
		}}

		for _, tc := range testCases {
			t.Run("creates inside a transaction and returns the created recipe", func(t *testing.T) {
				ctx := t.Context()
				services := newTestServices()

				mockUow := &commonmocks.UnitOfWork{}
				mockUow.On("Commit", mock.Anything).Return(nil)
				mockUow.On("Rollback", mock.Anything).Return(nil)
				services.mockTxm.On("Begin", mock.Anything, mock.Anything).Return(ctx, mockUow, nil)

				recipeIngredients := []recipes.RecipeIngredient{}

				expected := &recipes.Recipe{
					ID:          uuid.New(),
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					Title:       tc.dto.Title,
					Slug:        tc.slug,
					Headline:    &tc.dto.Headline,
					Description: &tc.dto.Description,
					Steps:       &tc.dto.Steps,
					PrepTime:    &tc.dto.PrepTime,
					Servings:    tc.dto.Servings,
					URL:         &tc.dto.URL,
					Tags:        tc.dto.Tags,
					Ingredients: recipeIngredients,
				}

				if expected.Servings < 1 {
					expected.Servings = 1
				}

				services.mockIngredientsRepo.On("UpsertMany", mock.Anything, mock.Anything).Return(recipeIngredients, nil)
				services.mockRecipesRepo.On("GetUniqueSlug", mock.Anything, mock.Anything).Return(tc.slug, nil)
				services.mockRecipesRepo.On("Create", mock.Anything, mock.Anything).Return(expected, nil)

				got, err := services.service.Create(ctx, tc.dto)

				require.NoError(t, err)
				require.Equal(t, expected, got)

				services.mockIngredientsRepo.AssertNumberOfCalls(t, "UpsertMany", 1)
				services.mockIngredientsRepo.AssertCalled(t, "UpsertMany", mock.Anything, tc.dto.Ingredients)

				services.mockRecipesRepo.AssertNumberOfCalls(t, "GetUniqueSlug", 1)
				services.mockRecipesRepo.AssertCalled(t, "GetUniqueSlug", mock.Anything, tc.dto.Title)

				services.mockRecipesRepo.AssertNumberOfCalls(t, "Create", 1)
				services.mockRecipesRepo.AssertCalled(t, "Create", mock.Anything, mock.MatchedBy(func(input any) bool {
					r := input.(recipes.Recipe)

					return r.ID.String() != uuid.Nil.String() &&
						r.Title == expected.Title &&
						r.Slug == tc.slug &&
						r.Headline == expected.Headline &&
						r.Description == expected.Description &&
						r.Steps == expected.Steps &&
						r.PrepTime == expected.PrepTime &&
						r.Servings == expected.Servings &&
						r.URL == expected.URL &&
						slices.Equal(r.Tags, expected.Tags) &&
						slices.Equal(r.Ingredients, expected.Ingredients)
				}))

				mockUow.AssertNumberOfCalls(t, "Commit", 1)
				mockUow.AssertCalled(t, "Commit", ctx)
				mockUow.AssertNumberOfCalls(t, "Rollback", 1)
				mockUow.AssertCalled(t, "Rollback", ctx)

				mock.AssertExpectationsForObjects(t, services.mockIngredientsRepo)
				mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
				mock.AssertExpectationsForObjects(t, mockUow)
			})
		}

		t.Run("rolls back and returns an error if ingredients operations fail", func(t *testing.T) {
			ctx := t.Context()
			services := newTestServices()

			mockUow := &commonmocks.UnitOfWork{}
			mockUow.On("Rollback", mock.Anything).Return(nil)
			services.mockTxm.On("Begin", mock.Anything, mock.Anything).Return(ctx, mockUow, nil)

			dto := &recipes.CreateUpdateRecipeDTO{Ingredients: []recipes.CreateRecipeIngredientDTO{}}

			repoErr := errors.New("repo error")

			services.mockIngredientsRepo.On("UpsertMany", mock.Anything, mock.Anything).Return([]recipes.RecipeIngredient{}, repoErr)

			created, err := services.service.Create(ctx, dto)

			require.Error(t, err)
			require.ErrorIs(t, err, repoErr)
			require.Nil(t, created)

			services.mockIngredientsRepo.AssertNumberOfCalls(t, "UpsertMany", 1)
			services.mockRecipesRepo.AssertNumberOfCalls(t, "Create", 0)

			mockUow.AssertNumberOfCalls(t, "Commit", 0)
			mockUow.AssertNumberOfCalls(t, "Rollback", 1)
			mockUow.AssertCalled(t, "Rollback", ctx)

			mock.AssertExpectationsForObjects(t, services.mockIngredientsRepo)
			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
			mock.AssertExpectationsForObjects(t, mockUow)
		})
		t.Run("rolls back and returns an error if recipes operations fail", func(t *testing.T) {
			ctx := t.Context()
			services := newTestServices()

			mockUow := &commonmocks.UnitOfWork{}
			mockUow.On("Rollback", mock.Anything).Return(nil)
			services.mockTxm.On("Begin", mock.Anything, mock.Anything).Return(ctx, mockUow, nil)

			dto := &recipes.CreateUpdateRecipeDTO{Ingredients: []recipes.CreateRecipeIngredientDTO{}}

			repoErr := errors.New("repo error")

			services.mockIngredientsRepo.On("UpsertMany", mock.Anything, mock.Anything).Return([]recipes.RecipeIngredient{}, nil)
			services.mockRecipesRepo.On("GetUniqueSlug", mock.Anything, mock.Anything).Return("some-slug", nil)
			services.mockRecipesRepo.On("Create", mock.Anything, mock.Anything).Return(new(recipes.Recipe), repoErr)

			created, err := services.service.Create(ctx, dto)

			require.Error(t, err)
			require.ErrorIs(t, err, repoErr)
			require.Nil(t, created)

			services.mockIngredientsRepo.AssertNumberOfCalls(t, "UpsertMany", 1)
			services.mockRecipesRepo.AssertNumberOfCalls(t, "GetUniqueSlug", 1)
			services.mockRecipesRepo.AssertNumberOfCalls(t, "Create", 1)

			mockUow.AssertNumberOfCalls(t, "Commit", 0)
			mockUow.AssertNumberOfCalls(t, "Rollback", 1)
			mockUow.AssertCalled(t, "Rollback", ctx)

			mock.AssertExpectationsForObjects(t, services.mockIngredientsRepo)
			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
			mock.AssertExpectationsForObjects(t, mockUow)
		})
	})
	t.Run("Get", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name     string
			idOrSlug string
			method   string
		}{
			{name: "by slug", idOrSlug: "some-slug", method: "GetBySlug"},
			{name: "by id", idOrSlug: uuid.NewString(), method: "GetByID"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Run("returns the recipe from the repo", func(t *testing.T) {
					ctx := t.Context()
					services := newTestServices()
					recipe := &recipes.Recipe{ID: uuid.New()}
					services.mockRecipesRepo.On(tc.method, mock.Anything, mock.Anything).Return(recipe, nil)

					result, err := services.service.Get(ctx, tc.idOrSlug)

					require.NoError(t, err)
					require.Equal(t, recipe, result)

					services.mockRecipesRepo.AssertNumberOfCalls(t, tc.method, 1)
					services.mockRecipesRepo.AssertCalled(t, tc.method, ctx, tc.idOrSlug)

					mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
				})
				t.Run("when the recipes repo fails it returns an error", func(t *testing.T) {
					ctx := t.Context()
					services := newTestServices()
					repoErr := errors.New("repo error")
					services.mockRecipesRepo.On(tc.method, mock.Anything, mock.Anything).Return(&recipes.Recipe{}, repoErr)

					_, err := services.service.Get(t.Context(), tc.idOrSlug)

					require.Error(t, err)
					require.ErrorIs(t, err, repoErr)

					services.mockRecipesRepo.AssertNumberOfCalls(t, tc.method, 1)
					services.mockRecipesRepo.AssertCalled(t, tc.method, ctx, tc.idOrSlug)

					mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
				})
			})
		}
	})
	t.Run("Update", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			dto       *recipes.CreateUpdateRecipeDTO
			recipe    *recipes.Recipe
			slug      string
			keepTitle bool
		}{
			{
				dto: func() *recipes.CreateUpdateRecipeDTO {
					dto := &recipes.CreateUpdateRecipeDTO{}
					gofakeit.Struct(&dto)
					return dto
				}(),
				recipe: func() *recipes.Recipe {
					dto := &recipes.Recipe{}
					gofakeit.Struct(&dto)
					return dto
				}(),
				slug:      slug.Make(gofakeit.Sentence(10)),
				keepTitle: true,
			},
			{
				dto: func() *recipes.CreateUpdateRecipeDTO {
					dto := &recipes.CreateUpdateRecipeDTO{}
					gofakeit.Struct(&dto)
					return dto
				}(),
				recipe: func() *recipes.Recipe {
					dto := &recipes.Recipe{}
					gofakeit.Struct(&dto)
					return dto
				}(),
				slug:      slug.Make(gofakeit.Sentence(10)),
				keepTitle: false,
			},
			{
				dto: func() *recipes.CreateUpdateRecipeDTO {
					dto := &recipes.CreateUpdateRecipeDTO{}
					gofakeit.Struct(&dto)
					dto.Servings = 0

					return dto
				}(),
				recipe: func() *recipes.Recipe {
					dto := &recipes.Recipe{}
					gofakeit.Struct(&dto)
					return dto
				}(),
				slug:      slug.Make(gofakeit.Sentence(10)),
				keepTitle: true,
			},
		}

		for _, tc := range testCases {
			t.Run("updates inside a transaction and returns the created recipe", func(t *testing.T) {
				ctx := t.Context()
				services := newTestServices()

				mockUow := &commonmocks.UnitOfWork{}
				mockUow.On("Commit", mock.Anything).Return(nil)
				mockUow.On("Rollback", mock.Anything).Return(nil)
				services.mockTxm.On("Begin", mock.Anything, mock.Anything).Return(ctx, mockUow, nil)

				recipeIngredients := []recipes.RecipeIngredient{}

				expected := &recipes.Recipe{
					ID:          tc.recipe.ID,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					Title:       tc.dto.Title,
					Slug:        tc.recipe.Slug,
					Headline:    &tc.dto.Headline,
					Description: &tc.dto.Description,
					Steps:       &tc.dto.Steps,
					PrepTime:    &tc.dto.PrepTime,
					Servings:    tc.dto.Servings,
					URL:         &tc.dto.URL,
					Tags:        tc.dto.Tags,
					Ingredients: recipeIngredients,
				}

				if tc.keepTitle {
					tc.recipe.Title = tc.dto.Title
				} else {
					expected.Slug = tc.slug
				}

				if expected.Servings < 1 {
					expected.Servings = 1
				}

				services.mockIngredientsRepo.On("UpsertMany", mock.Anything, mock.Anything).Return(recipeIngredients, nil)

				if !tc.keepTitle {
					services.mockRecipesRepo.On("GetUniqueSlug", mock.Anything, mock.Anything).Return(tc.slug, nil)
				}

				services.mockRecipesRepo.On("Update", mock.Anything, mock.Anything).Return(expected, nil)

				got, err := services.service.Update(ctx, tc.recipe, tc.dto)

				require.NoError(t, err)
				require.Equal(t, expected, got)

				services.mockIngredientsRepo.AssertNumberOfCalls(t, "UpsertMany", 1)
				services.mockIngredientsRepo.AssertCalled(t, "UpsertMany", mock.Anything, tc.dto.Ingredients)

				if !tc.keepTitle {
					services.mockRecipesRepo.AssertNumberOfCalls(t, "GetUniqueSlug", 1)
					services.mockRecipesRepo.AssertCalled(t, "GetUniqueSlug", mock.Anything, tc.dto.Title)
				}

				services.mockRecipesRepo.AssertNumberOfCalls(t, "Update", 1)
				services.mockRecipesRepo.AssertCalled(t, "Update", mock.Anything, mock.MatchedBy(func(input any) bool {
					r := input.(recipes.Recipe)

					return r.ID.String() != uuid.Nil.String() &&
						r.Title == expected.Title &&
						r.Slug == expected.Slug &&
						r.Headline == expected.Headline &&
						r.Description == expected.Description &&
						r.Steps == expected.Steps &&
						r.PrepTime == expected.PrepTime &&
						r.Servings == expected.Servings &&
						r.URL == expected.URL &&
						slices.Equal(r.Tags, expected.Tags) &&
						slices.Equal(r.Ingredients, expected.Ingredients)
				}))

				mockUow.AssertNumberOfCalls(t, "Commit", 1)
				mockUow.AssertNumberOfCalls(t, "Rollback", 1)

				mock.AssertExpectationsForObjects(t, services.mockIngredientsRepo)
				mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
				mock.AssertExpectationsForObjects(t, mockUow)
			})
		}

		t.Run("rolls back and returns an error if ingredients operations fail", func(t *testing.T) {
			ctx := t.Context()
			services := newTestServices()

			mockUow := &commonmocks.UnitOfWork{}
			mockUow.On("Rollback", mock.Anything).Return(nil)
			services.mockTxm.On("Begin", mock.Anything, mock.Anything).Return(ctx, mockUow, nil)

			dto := &recipes.CreateUpdateRecipeDTO{Ingredients: []recipes.CreateRecipeIngredientDTO{}}

			repoErr := errors.New("repo error")

			services.mockIngredientsRepo.On("UpsertMany", mock.Anything, mock.Anything).Return([]recipes.RecipeIngredient{}, repoErr)

			created, err := services.service.Update(ctx, &recipes.Recipe{
				ID: uuid.New(),
			}, dto)

			require.Error(t, err)
			require.ErrorIs(t, err, repoErr)
			require.Nil(t, created)

			services.mockIngredientsRepo.AssertNumberOfCalls(t, "UpsertMany", 1)
			services.mockRecipesRepo.AssertNumberOfCalls(t, "Update", 0)

			mockUow.AssertNumberOfCalls(t, "Commit", 0)
			mockUow.AssertNumberOfCalls(t, "Rollback", 1)
			mockUow.AssertCalled(t, "Rollback", ctx)

			mock.AssertExpectationsForObjects(t, services.mockIngredientsRepo)
			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
			mock.AssertExpectationsForObjects(t, mockUow)
		})
		t.Run("rolls back and returns an error if recipes operations fail", func(t *testing.T) {
			ctx := t.Context()
			services := newTestServices()

			mockUow := &commonmocks.UnitOfWork{}
			mockUow.On("Rollback", mock.Anything).Return(nil)
			services.mockTxm.On("Begin", mock.Anything, mock.Anything).Return(ctx, mockUow, nil)

			dto := &recipes.CreateUpdateRecipeDTO{Ingredients: []recipes.CreateRecipeIngredientDTO{}}

			repoErr := errors.New("repo error")

			services.mockIngredientsRepo.On("UpsertMany", mock.Anything, mock.Anything).Return([]recipes.RecipeIngredient{}, nil)
			services.mockRecipesRepo.On("Update", mock.Anything, mock.Anything).Return(new(recipes.Recipe), repoErr)

			created, err := services.service.Update(ctx, &recipes.Recipe{
				ID: uuid.New(),
			}, dto)

			require.Error(t, err)
			require.ErrorIs(t, err, repoErr)
			require.Nil(t, created)

			services.mockIngredientsRepo.AssertNumberOfCalls(t, "UpsertMany", 1)
			services.mockRecipesRepo.AssertNumberOfCalls(t, "Update", 1)

			mockUow.AssertNumberOfCalls(t, "Commit", 0)
			mockUow.AssertNumberOfCalls(t, "Rollback", 1)
			mockUow.AssertCalled(t, "Rollback", ctx)

			mock.AssertExpectationsForObjects(t, services.mockIngredientsRepo)
			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
			mock.AssertExpectationsForObjects(t, mockUow)
		})
	})
	t.Run("Delete", func(t *testing.T) {
		t.Parallel()
		t.Run("return no error if delete succeeds", func(t *testing.T) {
			ctx := t.Context()
			services := newTestServices()
			recipe := &recipes.Recipe{ID: uuid.New()}

			services.mockRecipesRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)

			err := services.service.Delete(ctx, recipe.ID.String())

			require.NoError(t, err)

			services.mockRecipesRepo.AssertNumberOfCalls(t, "Delete", 1)
			services.mockRecipesRepo.AssertCalled(t, "Delete", ctx, recipe.ID.String())

			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
		})
		t.Run("returns an error if it fails", func(t *testing.T) {
			ctx := t.Context()
			services := newTestServices()
			recipe := &recipes.Recipe{ID: uuid.New()}
			repoErr := errors.New("repo error")
			services.mockRecipesRepo.On("Delete", mock.Anything, mock.Anything).Return(repoErr)

			err := services.service.Delete(ctx, recipe.ID.String())

			require.Error(t, err)
			require.ErrorIs(t, err, repoErr)

			services.mockRecipesRepo.AssertNumberOfCalls(t, "Delete", 1)
			services.mockRecipesRepo.AssertCalled(t, "Delete", ctx, recipe.ID.String())

			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
		})
	})
}
