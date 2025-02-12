package recipes

import (
	"context"

	"github.com/stretchr/testify/mock"
)

var (
	_ RecipesRepo     = (*MockRecipesRepo)(nil)
	_ IngredientsRepo = (*MockIngredientsRepo)(nil)
)

type MockRecipesRepo struct {
	mock.Mock
}

func (m *MockRecipesRepo) GetAll(ctx context.Context) ([]Recipe, error) {
	args := m.Called(ctx)
	return args.Get(0).([]Recipe), args.Error(1)
}

func (m *MockRecipesRepo) Create(ctx context.Context, recipe Recipe) (*Recipe, error) {
	args := m.Called(ctx, recipe)
	return args.Get(0).(*Recipe), args.Error(1)
}

func (m *MockRecipesRepo) GetByID(ctx context.Context, recipeID string) (*Recipe, error) {
	args := m.Called(ctx, recipeID)
	return args.Get(0).(*Recipe), args.Error(1)
}

func (m *MockRecipesRepo) GetBySlug(ctx context.Context, recipeSlug string) (*Recipe, error) {
	args := m.Called(ctx, recipeSlug)
	return args.Get(0).(*Recipe), args.Error(1)
}

func (m *MockRecipesRepo) Update(ctx context.Context, recipe Recipe) (*Recipe, error) {
	args := m.Called(ctx, recipe)
	return args.Get(0).(*Recipe), args.Error(1)
}

func (m *MockRecipesRepo) Delete(ctx context.Context, recipeID string) error {
	args := m.Called(ctx, recipeID)
	return args.Error(0)
}

type MockIngredientsRepo struct {
	mock.Mock
}

func (m *MockIngredientsRepo) UpsertMany(ctx context.Context, ingredients []CreateRecipeIngredientDTO) ([]RecipeIngredient, error) {
	args := m.Called(ctx, ingredients)
	return args.Get(0).([]RecipeIngredient), args.Error(1)
}
