//nolint:wrapcheck,exhaustruct
package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AlejandroHerr/cook-book-go/internal/common"
	"github.com/AlejandroHerr/cook-book-go/internal/common/logging"
	"github.com/AlejandroHerr/cook-book-go/internal/core/dtos"
	"github.com/AlejandroHerr/cook-book-go/internal/core/model"
	"github.com/AlejandroHerr/cook-book-go/internal/core/usecases"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type testServices struct {
	mockTxm             *MockTransactionManager
	mockRecipesRepo     *MockRecipesRepo
	mockIngredientsRepo *MockIngredientsRepo
	recipeUsecases      *usecases.RecipeUseCases
}

type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) Begin(ctx context.Context) (common.Transaction, error) {
	args := m.Called(ctx)
	return args.Get(0).(common.Transaction), args.Error(1)
}

type MockTransaction struct {
	mock.Mock
}

func (m *MockTransaction) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTransaction) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTransaction) Transaction() interface{} {
	return nil
}

type MockRecipesRepo struct {
	mock.Mock
}

func (m *MockRecipesRepo) GetAll(ctx context.Context) ([]*model.Recipe, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.Recipe), args.Error(1)
}

func (m *MockRecipesRepo) Create(ctx context.Context, recipe model.Recipe) (*model.Recipe, error) {
	args := m.Called(ctx, recipe)
	return args.Get(0).(*model.Recipe), args.Error(1)
}

func (m *MockRecipesRepo) GetByID(ctx context.Context, recipeID string) (*model.Recipe, error) {
	args := m.Called(ctx, recipeID)
	return args.Get(0).(*model.Recipe), args.Error(1)
}

func (m *MockRecipesRepo) GetBySlug(ctx context.Context, recipeSlug string) (*model.Recipe, error) {
	args := m.Called(ctx, recipeSlug)
	return args.Get(0).(*model.Recipe), args.Error(1)
}

func (m *MockRecipesRepo) Update(ctx context.Context, recipe model.Recipe) (*model.Recipe, error) {
	args := m.Called(ctx, recipe)
	return args.Get(0).(*model.Recipe), args.Error(1)
}

func (m *MockRecipesRepo) Delete(ctx context.Context, recipeID string) error {
	args := m.Called(ctx, recipeID)
	return args.Error(0)
}

type MockIngredientsRepo struct {
	mock.Mock
}

func (m *MockIngredientsRepo) UpsertMany(ctx context.Context, names []string) ([]model.Ingredient, error) {
	args := m.Called(ctx, names)
	return args.Get(0).([]model.Ingredient), args.Error(1)
}

func newTestServices() *testServices {
	mockTxm := &MockTransactionManager{}
	mockRecipesRepo := &MockRecipesRepo{}
	mockIngredientsRepo := &MockIngredientsRepo{}

	recipeUsecases := usecases.NewRecipesUseCases(mockTxm, mockRecipesRepo, mockIngredientsRepo, logging.NewVoidLogger())

	return &testServices{
		mockTxm:             mockTxm,
		mockRecipesRepo:     mockRecipesRepo,
		mockIngredientsRepo: mockIngredientsRepo,
		recipeUsecases:      recipeUsecases,
	}
}

func TestRecipesUseCases(t *testing.T) {
	t.Run("GetAll", func(t *testing.T) {
		t.Parallel()
		t.Run("returns the recipes from the repo", func(t *testing.T) {
			services := newTestServices()
			recipes := []*model.Recipe{{ID: uuid.New()}}
			services.mockRecipesRepo.On("GetAll", mock.Anything).Return(recipes, nil)

			result, err := services.recipeUsecases.GetAll(context.Background())

			require.NoError(t, err)
			require.Equal(t, recipes, result)

			services.mockRecipesRepo.AssertNumberOfCalls(t, "GetAll", 1)

			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
		})
		t.Run("when the recipes repo fails it returns an error", func(t *testing.T) {
			services := newTestServices()
			repoErr := errors.New("repo error")
			services.mockRecipesRepo.On("GetAll", mock.Anything).Return([]*model.Recipe{}, repoErr)

			_, err := services.recipeUsecases.GetAll(context.Background())

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

			mockTransaction := &MockTransaction{}
			mockTransaction.On("Commit").Return(nil)
			mockTransaction.On("Rollback").Return(nil)
			services.mockTxm.On("Begin", mock.Anything, mock.Anything).Return(mockTransaction, nil)

			ingredientNames := []string{"ingredient1", "ingredient2"}
			recipeIngredients := []dtos.CreateRecipeIngredient{{Name: ingredientNames[0]}, {Name: ingredientNames[1]}}
			dto := &dtos.CreateUpdateRecipeDTO{Ingredients: recipeIngredients}
			created := &model.Recipe{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now()} //nolint:exhaustruct

			upsertedIngredients := []model.Ingredient{{ID: uuid.New(), Name: ingredientNames[0]}, {ID: uuid.New(), Name: ingredientNames[1]}}

			services.mockIngredientsRepo.On("UpsertMany", mock.Anything, mock.Anything).Return(upsertedIngredients, nil)
			services.mockRecipesRepo.On("Create", mock.Anything, mock.Anything).Return(created, nil)

			got, err := services.recipeUsecases.Create(ctx, dto)

			require.NoError(t, err)
			require.Equal(t, created, got)

			services.mockIngredientsRepo.AssertNumberOfCalls(t, "UpsertMany", 1)
			services.mockIngredientsRepo.AssertCalled(t, "UpsertMany", mock.Anything, ingredientNames)

			services.mockRecipesRepo.AssertNumberOfCalls(t, "Create", 1)
			services.mockRecipesRepo.AssertCalled(t, "Create", mock.Anything, mock.AnythingOfType("model.Recipe"))

			mockTransaction.AssertNumberOfCalls(t, "Commit", 1)
			mockTransaction.AssertNumberOfCalls(t, "Rollback", 1)

			mock.AssertExpectationsForObjects(t, services.mockIngredientsRepo)
			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
			mock.AssertExpectationsForObjects(t, mockTransaction)
		})
		t.Run("rolls back and returns an error if ingredients operations fail", func(t *testing.T) {
			ctx := context.Background()
			services := newTestServices()

			mockTransaction := &MockTransaction{}
			mockTransaction.On("Rollback").Return(nil)
			services.mockTxm.On("Begin", mock.Anything, mock.Anything).Return(mockTransaction, nil)

			dto := &dtos.CreateUpdateRecipeDTO{Ingredients: []dtos.CreateRecipeIngredient{}}

			repoErr := errors.New("repo error")

			services.mockIngredientsRepo.On("UpsertMany", mock.Anything, mock.Anything).Return([]model.Ingredient{}, repoErr)

			created, err := services.recipeUsecases.Create(ctx, dto)

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

			mockTransaction := &MockTransaction{}
			mockTransaction.On("Rollback").Return(nil)
			services.mockTxm.On("Begin", mock.Anything, mock.Anything).Return(mockTransaction, nil)

			dto := &dtos.CreateUpdateRecipeDTO{Ingredients: []dtos.CreateRecipeIngredient{}}

			repoErr := errors.New("repo error")

			services.mockIngredientsRepo.On("UpsertMany", mock.Anything, mock.Anything).Return([]model.Ingredient{}, nil)
			services.mockRecipesRepo.On("Create", mock.Anything, mock.Anything).Return(new(model.Recipe), repoErr)

			created, err := services.recipeUsecases.Create(ctx, dto)

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
					recipe := &model.Recipe{ID: uuid.New()}
					services.mockRecipesRepo.On(tc.method, mock.Anything, mock.Anything).Return(recipe, nil)

					result, err := services.recipeUsecases.Get(ctx, tc.idOrSlug)

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
					services.mockRecipesRepo.On(tc.method, mock.Anything, mock.Anything).Return(&model.Recipe{}, repoErr)

					_, err := services.recipeUsecases.Get(context.Background(), tc.idOrSlug)

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

			mockTransaction := &MockTransaction{}
			mockTransaction.On("Commit").Return(nil)
			mockTransaction.On("Rollback").Return(nil)
			services.mockTxm.On("Begin", mock.Anything, mock.Anything).Return(mockTransaction, nil)

			ingredientNames := []string{"ingredient1", "ingredient2"}
			recipeIngredients := []dtos.CreateRecipeIngredient{{Name: ingredientNames[0]}, {Name: ingredientNames[1]}}
			dto := &dtos.CreateUpdateRecipeDTO{Ingredients: recipeIngredients}
			expected := &model.Recipe{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now()} //nolint:exhaustruct

			upsertedIngredients := []model.Ingredient{{ID: uuid.New(), Name: ingredientNames[0]}, {ID: uuid.New(), Name: ingredientNames[1]}}

			services.mockIngredientsRepo.On("UpsertMany", mock.Anything, mock.Anything).Return(upsertedIngredients, nil)
			services.mockRecipesRepo.On("Update", mock.Anything, mock.Anything).Return(expected, nil)

			got, err := services.recipeUsecases.Update(ctx, uuid.New(), dto)

			require.NoError(t, err)
			require.NotNil(t, got)

			services.mockIngredientsRepo.AssertNumberOfCalls(t, "UpsertMany", 1)
			services.mockIngredientsRepo.AssertCalled(t, "UpsertMany", mock.Anything, ingredientNames)

			services.mockRecipesRepo.AssertNumberOfCalls(t, "Update", 1)
			services.mockRecipesRepo.AssertCalled(t, "Update", mock.Anything, mock.AnythingOfType("model.Recipe"))

			mockTransaction.AssertNumberOfCalls(t, "Commit", 1)
			mockTransaction.AssertNumberOfCalls(t, "Rollback", 1)

			mock.AssertExpectationsForObjects(t, services.mockIngredientsRepo)
			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
			mock.AssertExpectationsForObjects(t, mockTransaction)
		})
		t.Run("rolls back and returns an error if ingredients operations fail", func(t *testing.T) {
			ctx := context.Background()
			services := newTestServices()

			mockTransaction := &MockTransaction{}
			mockTransaction.On("Rollback").Return(nil)
			services.mockTxm.On("Begin", mock.Anything, mock.Anything).Return(mockTransaction, nil)

			dto := &dtos.CreateUpdateRecipeDTO{Ingredients: []dtos.CreateRecipeIngredient{}}

			repoErr := errors.New("repo error")

			services.mockIngredientsRepo.On("UpsertMany", mock.Anything, mock.Anything).Return([]model.Ingredient{}, repoErr)

			created, err := services.recipeUsecases.Update(ctx, uuid.New(), dto)

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

			mockTransaction := &MockTransaction{}
			mockTransaction.On("Rollback").Return(nil)
			services.mockTxm.On("Begin", mock.Anything, mock.Anything).Return(mockTransaction, nil)

			dto := &dtos.CreateUpdateRecipeDTO{Ingredients: []dtos.CreateRecipeIngredient{}}

			repoErr := errors.New("repo error")

			services.mockIngredientsRepo.On("UpsertMany", mock.Anything, mock.Anything).Return([]model.Ingredient{}, nil)
			services.mockRecipesRepo.On("Update", mock.Anything, mock.Anything).Return(new(model.Recipe), repoErr)

			created, err := services.recipeUsecases.Update(ctx, uuid.New(), dto)

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
			recipe := &model.Recipe{ID: uuid.New()}

			services.mockRecipesRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)

			err := services.recipeUsecases.Delete(ctx, recipe.ID.String())

			require.NoError(t, err)

			services.mockRecipesRepo.AssertNumberOfCalls(t, "Delete", 1)
			services.mockRecipesRepo.AssertCalled(t, "Delete", ctx, recipe.ID.String())

			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
		})
		t.Run("returns an error if it fails", func(t *testing.T) {
			ctx := context.Background()
			services := newTestServices()
			recipe := &model.Recipe{ID: uuid.New()}
			repoErr := errors.New("repo error")
			services.mockRecipesRepo.On("Delete", mock.Anything, mock.Anything).Return(repoErr)

			err := services.recipeUsecases.Delete(ctx, recipe.ID.String())

			require.Error(t, err)
			require.ErrorIs(t, err, repoErr)

			services.mockRecipesRepo.AssertNumberOfCalls(t, "Delete", 1)
			services.mockRecipesRepo.AssertCalled(t, "Delete", ctx, recipe.ID.String())

			mock.AssertExpectationsForObjects(t, services.mockRecipesRepo)
		})
	})
}
