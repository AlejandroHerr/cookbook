//nolint:exhaustruct
package recipes_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AlejandroHerr/cookbook/internal/common"
	"github.com/AlejandroHerr/cookbook/internal/common/logging"
	"github.com/AlejandroHerr/cookbook/internal/recipes"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type testServices struct {
	mockTxm             *common.MockTransactionManager
	mockRecipesRepo     *recipes.MockRecipesRepo
	mockIngredientsRepo *recipes.MockIngredientsRepo
	useCases            *recipes.UseCases
}

func newTestServices() *testServices {
	mockTxm := new(common.MockTransactionManager)
	mockRecipesRepo := new(recipes.MockRecipesRepo)
	mockIngredientsRepo := new(recipes.MockIngredientsRepo)

	useCases := recipes.MakeUseCases(mockTxm, mockRecipesRepo, mockIngredientsRepo, logging.NewVoidLogger())

	return &testServices{
		mockTxm:             mockTxm,
		mockRecipesRepo:     mockRecipesRepo,
		mockIngredientsRepo: mockIngredientsRepo,
		useCases:            useCases,
	}
}

func TestRecipesUseCases(t *testing.T) {
	t.Run("GetAll", func(t *testing.T) {
		t.Parallel()
		t.Run("returns the recipes from the repo", func(t *testing.T) {
			services := newTestServices()
			recipes := []*recipes.Recipe{{ID: uuid.New()}}
			services.mockRecipesRepo.On("GetAll", mock.Anything).Return(recipes, nil)

			result, err := services.useCases.GetAll(context.Background())

			require.NoError(t, err)
			require.Equal(t, recipes, result)

			services.mockRecipesRepo.AssertNumberOfCalls(t, "GetAll", 1)

			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
		})
		t.Run("when the recipes repo fails it returns an error", func(t *testing.T) {
			services := newTestServices()
			repoErr := errors.New("repo error")
			services.mockRecipesRepo.On("GetAll", mock.Anything).Return([]*recipes.Recipe{}, repoErr)

			_, err := services.useCases.GetAll(context.Background())

			require.Error(t, err)
			require.ErrorIs(t, err, repoErr)

			services.mockRecipesRepo.AssertNumberOfCalls(t, "GetAll", 1)

			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
		})
	})
	t.Run("Create", func(t *testing.T) {
		t.Parallel()
		t.Run("creates inside a transaction and returns the created recipe", func(t *testing.T) {
			ctx := context.Background()
			services := newTestServices()

			mockTransaction := new(common.MockTransaction)
			mockTransaction.On("Commit").Return(nil)
			mockTransaction.On("Rollback").Return(nil)
			services.mockTxm.On("Begin", mock.Anything, mock.Anything).Return(mockTransaction, nil)

			dto := &recipes.CreateUpdateRecipeDTO{Ingredients: []recipes.CreateRecipeIngredientDTO{{Name: "ingredient1"}, {Name: "ingredient2"}}}
			created := &recipes.Recipe{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now()}

			services.mockIngredientsRepo.On("UpsertMany", mock.Anything, mock.Anything).Return([]recipes.RecipeIngredient{}, nil)
			services.mockRecipesRepo.On("Create", mock.Anything, mock.Anything).Return(created, nil)

			got, err := services.useCases.Create(ctx, dto)

			require.NoError(t, err)
			require.Equal(t, created, got)

			services.mockIngredientsRepo.AssertNumberOfCalls(t, "UpsertMany", 1)
			services.mockIngredientsRepo.AssertCalled(t, "UpsertMany", mock.Anything, dto.Ingredients)

			services.mockRecipesRepo.AssertNumberOfCalls(t, "Create", 1)
			services.mockRecipesRepo.AssertCalled(t, "Create", mock.Anything, mock.AnythingOfType("recipes.Recipe"))

			mockTransaction.AssertNumberOfCalls(t, "Commit", 1)
			mockTransaction.AssertNumberOfCalls(t, "Rollback", 1)

			mock.AssertExpectationsForObjects(t, services.mockIngredientsRepo)
			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
			mock.AssertExpectationsForObjects(t, mockTransaction)
		})
		t.Run("rolls back and returns an error if ingredients operations fail", func(t *testing.T) {
			ctx := context.Background()
			services := newTestServices()

			mockTransaction := new(common.MockTransaction)
			mockTransaction.On("Rollback").Return(nil)
			services.mockTxm.On("Begin", mock.Anything, mock.Anything).Return(mockTransaction, nil)

			dto := &recipes.CreateUpdateRecipeDTO{Ingredients: []recipes.CreateRecipeIngredientDTO{}}

			repoErr := errors.New("repo error")

			services.mockIngredientsRepo.On("UpsertMany", mock.Anything, mock.Anything).Return([]recipes.RecipeIngredient{}, repoErr)

			created, err := services.useCases.Create(ctx, dto)

			require.Error(t, err)
			require.ErrorIs(t, err, repoErr)
			require.Nil(t, created)

			services.mockIngredientsRepo.AssertNumberOfCalls(t, "UpsertMany", 1)
			services.mockRecipesRepo.AssertNumberOfCalls(t, "Create", 0)

			mockTransaction.AssertNumberOfCalls(t, "Commit", 0)
			mockTransaction.AssertNumberOfCalls(t, "Rollback", 1)

			mock.AssertExpectationsForObjects(t, services.mockIngredientsRepo)
			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
			mock.AssertExpectationsForObjects(t, mockTransaction)
		})
		t.Run("rolls back and returns an error if recipes operations fail", func(t *testing.T) {
			ctx := context.Background()
			services := newTestServices()

			mockTransaction := new(common.MockTransaction)
			mockTransaction.On("Rollback").Return(nil)
			services.mockTxm.On("Begin", mock.Anything, mock.Anything).Return(mockTransaction, nil)

			dto := &recipes.CreateUpdateRecipeDTO{Ingredients: []recipes.CreateRecipeIngredientDTO{}}

			repoErr := errors.New("repo error")

			services.mockIngredientsRepo.On("UpsertMany", mock.Anything, mock.Anything).Return([]recipes.RecipeIngredient{}, nil)
			services.mockRecipesRepo.On("Create", mock.Anything, mock.Anything).Return(new(recipes.Recipe), repoErr)

			created, err := services.useCases.Create(ctx, dto)

			require.Error(t, err)
			require.ErrorIs(t, err, repoErr)
			require.Nil(t, created)

			services.mockIngredientsRepo.AssertNumberOfCalls(t, "UpsertMany", 1)
			services.mockRecipesRepo.AssertNumberOfCalls(t, "Create", 1)

			mockTransaction.AssertNumberOfCalls(t, "Commit", 0)
			mockTransaction.AssertNumberOfCalls(t, "Rollback", 1)

			mock.AssertExpectationsForObjects(t, services.mockIngredientsRepo)
			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
			mock.AssertExpectationsForObjects(t, mockTransaction)
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
					ctx := context.Background()
					services := newTestServices()
					recipe := &recipes.Recipe{ID: uuid.New()}
					services.mockRecipesRepo.On(tc.method, mock.Anything, mock.Anything).Return(recipe, nil)

					result, err := services.useCases.Get(ctx, tc.idOrSlug)

					require.NoError(t, err)
					require.Equal(t, recipe, result)

					services.mockRecipesRepo.AssertNumberOfCalls(t, tc.method, 1)
					services.mockRecipesRepo.AssertCalled(t, tc.method, ctx, tc.idOrSlug)

					mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
				})
				t.Run("when the recipes repo fails it returns an error", func(t *testing.T) {
					ctx := context.Background()
					services := newTestServices()
					repoErr := errors.New("repo error")
					services.mockRecipesRepo.On(tc.method, mock.Anything, mock.Anything).Return(&recipes.Recipe{}, repoErr)

					_, err := services.useCases.Get(context.Background(), tc.idOrSlug)

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
		t.Run("updates inside a transaction and returns the created recipe", func(t *testing.T) {
			ctx := context.Background()
			services := newTestServices()

			mockTransaction := new(common.MockTransaction)
			mockTransaction.On("Commit").Return(nil)
			mockTransaction.On("Rollback").Return(nil)
			services.mockTxm.On("Begin", mock.Anything, mock.Anything).Return(mockTransaction, nil)

			id := uuid.New()
			dto := &recipes.CreateUpdateRecipeDTO{Ingredients: []recipes.CreateRecipeIngredientDTO{{Name: "ingredient1"}, {Name: "ingredient2"}}}
			created := &recipes.Recipe{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now()}

			services.mockIngredientsRepo.On("UpsertMany", mock.Anything, mock.Anything).Return([]recipes.RecipeIngredient{}, nil)
			services.mockRecipesRepo.On("Update", mock.Anything, mock.Anything).Return(created, nil)

			got, err := services.useCases.Update(ctx, id, dto)

			require.NoError(t, err)
			require.NotNil(t, got)

			services.mockIngredientsRepo.AssertNumberOfCalls(t, "UpsertMany", 1)
			services.mockIngredientsRepo.AssertCalled(t, "UpsertMany", mock.Anything, dto.Ingredients)

			services.mockRecipesRepo.AssertNumberOfCalls(t, "Update", 1)
			services.mockRecipesRepo.AssertCalled(t, "Update", mock.Anything, mock.AnythingOfType("recipes.Recipe"))

			mockTransaction.AssertNumberOfCalls(t, "Commit", 1)
			mockTransaction.AssertNumberOfCalls(t, "Rollback", 1)

			mock.AssertExpectationsForObjects(t, services.mockIngredientsRepo)
			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
			mock.AssertExpectationsForObjects(t, mockTransaction)
		})
		t.Run("rolls back and returns an error if ingredients operations fail", func(t *testing.T) {
			ctx := context.Background()
			services := newTestServices()

			mockTransaction := new(common.MockTransaction)
			mockTransaction.On("Rollback").Return(nil)
			services.mockTxm.On("Begin", mock.Anything, mock.Anything).Return(mockTransaction, nil)

			dto := &recipes.CreateUpdateRecipeDTO{Ingredients: []recipes.CreateRecipeIngredientDTO{}}

			repoErr := errors.New("repo error")

			services.mockIngredientsRepo.On("UpsertMany", mock.Anything, mock.Anything).Return([]recipes.RecipeIngredient{}, repoErr)

			created, err := services.useCases.Update(ctx, uuid.New(), dto)

			require.Error(t, err)
			require.ErrorIs(t, err, repoErr)
			require.Nil(t, created)

			services.mockIngredientsRepo.AssertNumberOfCalls(t, "UpsertMany", 1)
			services.mockRecipesRepo.AssertNumberOfCalls(t, "Update", 0)

			mockTransaction.AssertNumberOfCalls(t, "Commit", 0)
			mockTransaction.AssertNumberOfCalls(t, "Rollback", 1)

			mock.AssertExpectationsForObjects(t, services.mockIngredientsRepo)
			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
			mock.AssertExpectationsForObjects(t, mockTransaction)
		})
		t.Run("rolls back and returns an error if recipes operations fail", func(t *testing.T) {
			ctx := context.Background()
			services := newTestServices()

			mockTransaction := new(common.MockTransaction)
			mockTransaction.On("Rollback").Return(nil)
			services.mockTxm.On("Begin", mock.Anything, mock.Anything).Return(mockTransaction, nil)

			dto := &recipes.CreateUpdateRecipeDTO{Ingredients: []recipes.CreateRecipeIngredientDTO{}}

			repoErr := errors.New("repo error")

			services.mockIngredientsRepo.On("UpsertMany", mock.Anything, mock.Anything).Return([]recipes.RecipeIngredient{}, nil)
			services.mockRecipesRepo.On("Update", mock.Anything, mock.Anything).Return(new(recipes.Recipe), repoErr)

			created, err := services.useCases.Update(ctx, uuid.New(), dto)

			require.Error(t, err)
			require.ErrorIs(t, err, repoErr)
			require.Nil(t, created)

			services.mockIngredientsRepo.AssertNumberOfCalls(t, "UpsertMany", 1)
			services.mockRecipesRepo.AssertNumberOfCalls(t, "Update", 1)

			mockTransaction.AssertNumberOfCalls(t, "Commit", 0)
			mockTransaction.AssertNumberOfCalls(t, "Rollback", 1)

			mock.AssertExpectationsForObjects(t, services.mockIngredientsRepo)
			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
			mock.AssertExpectationsForObjects(t, mockTransaction)
		})
	})
	t.Run("Delete", func(t *testing.T) {
		t.Parallel()
		t.Run("return no error if delete succeeds", func(t *testing.T) {
			ctx := context.Background()
			services := newTestServices()
			recipe := &recipes.Recipe{ID: uuid.New()}

			services.mockRecipesRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)

			err := services.useCases.Delete(ctx, recipe.ID.String())

			require.NoError(t, err)

			services.mockRecipesRepo.AssertNumberOfCalls(t, "Delete", 1)
			services.mockRecipesRepo.AssertCalled(t, "Delete", ctx, recipe.ID.String())

			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
		})
		t.Run("returns an error if it fails", func(t *testing.T) {
			ctx := context.Background()
			services := newTestServices()
			recipe := &recipes.Recipe{ID: uuid.New()}
			repoErr := errors.New("repo error")
			services.mockRecipesRepo.On("Delete", mock.Anything, mock.Anything).Return(repoErr)

			err := services.useCases.Delete(ctx, recipe.ID.String())

			require.Error(t, err)
			require.ErrorIs(t, err, repoErr)

			services.mockRecipesRepo.AssertNumberOfCalls(t, "Delete", 1)
			services.mockRecipesRepo.AssertCalled(t, "Delete", ctx, recipe.ID.String())

			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
		})
	})
}
