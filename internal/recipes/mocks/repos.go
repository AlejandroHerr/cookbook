//nolint:errcheck,wrapcheck
package mocks

import (
	"context"

	"github.com/AlejandroHerr/cookbook/internal/recipes"
	"github.com/stretchr/testify/mock"
)

var (
	_ recipes.RecipesRepo     = (*RecipesRepo)(nil)
	_ recipes.IngredientsRepo = (*IngredientsRepo)(nil)
)

type RecipesRepo struct {
	mock.Mock
}

func (m *RecipesRepo) List(ctx context.Context) ([]recipes.Recipe, error) {
	args := m.Called(ctx)
	return args.Get(0).([]recipes.Recipe), args.Error(1)
}

func (m *RecipesRepo) Create(ctx context.Context, recipe recipes.Recipe) (*recipes.Recipe, error) {
	args := m.Called(ctx, recipe)
	return args.Get(0).(*recipes.Recipe), args.Error(1)
}

func (m *RecipesRepo) GetByID(ctx context.Context, recipeID string) (*recipes.Recipe, error) {
	args := m.Called(ctx, recipeID)
	return args.Get(0).(*recipes.Recipe), args.Error(1)
}

func (m *RecipesRepo) GetBySlug(ctx context.Context, recipeSlug string) (*recipes.Recipe, error) {
	args := m.Called(ctx, recipeSlug)
	return args.Get(0).(*recipes.Recipe), args.Error(1)
}

func (m *RecipesRepo) Update(ctx context.Context, recipe recipes.Recipe) (*recipes.Recipe, error) {
	args := m.Called(ctx, recipe)
	return args.Get(0).(*recipes.Recipe), args.Error(1)
}

func (m *RecipesRepo) Delete(ctx context.Context, recipeID string) error {
	args := m.Called(ctx, recipeID)
	return args.Error(0)
}

func (m *RecipesRepo) GetUniqueSlug(ctx context.Context, title string) (string, error) {
	args := m.Called(ctx, title)
	return args.String(0), args.Error(1)
}

type IngredientsRepo struct {
	mock.Mock
}

func (m *IngredientsRepo) UpsertMany(ctx context.Context, ingredients []recipes.CreateRecipeIngredientDTO) ([]recipes.RecipeIngredient, error) { //nolint:lll
	args := m.Called(ctx, ingredients)
	return args.Get(0).([]recipes.RecipeIngredient), args.Error(1)
}
