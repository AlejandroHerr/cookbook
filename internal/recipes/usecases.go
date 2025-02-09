package recipes

import (
	"context"
	"fmt"

	"github.com/AlejandroHerr/cookbook/internal/common"
	"github.com/AlejandroHerr/cookbook/internal/common/logging"
	"github.com/google/uuid"
)

type RecipesRepo interface {
	GetAll(ctx context.Context) ([]*Recipe, error)
	Create(ctx context.Context, recipe Recipe) (*Recipe, error)
	GetByID(ctx context.Context, recipeID string) (*Recipe, error)
	GetBySlug(ctx context.Context, recipeSlug string) (*Recipe, error)
	Update(ctx context.Context, recipe Recipe) (*Recipe, error)
	Delete(ctx context.Context, recipeID string) error
}

type IngredientsRepo interface {
	UpsertMany(ctx context.Context, names []CreateRecipeIngredientDTO) ([]RecipeIngredient, error)
}

type UseCases struct {
	recipesRepo        RecipesRepo
	ingredientsRepo    IngredientsRepo
	transactionManager common.TransactionManager
	logger             logging.Logger
}

func NewUseCases(
	transactionManager common.TransactionManager,
	recipesRepo RecipesRepo,
	ingredientsRepo IngredientsRepo,
	logger logging.Logger,
) *UseCases {
	return &UseCases{
		transactionManager: transactionManager,
		recipesRepo:        recipesRepo,
		ingredientsRepo:    ingredientsRepo,
		logger:             logger,
	}
}

func (u UseCases) GetAll(ctx context.Context) ([]*Recipe, error) {
	recipes, err := u.recipesRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting all recipes: %w", err)
	}

	return recipes, nil
}

func (u UseCases) Create(ctx context.Context, dto *CreateUpdateRecipeDTO) (*Recipe, error) {
	transaction, err := u.transactionManager.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("transaction manager begin: %w", err)
	}

	defer func() {
		err = transaction.Rollback()
		u.logger.Errorw("error rolling back transaction", "error", err)
	}()

	ctxWithTransaction := context.WithValue(ctx, common.TransactionContextKey{}, transaction)

	recipeIngredients, err := u.ingredientsRepo.UpsertMany(ctxWithTransaction, dto.Ingredients)
	if err != nil {
		return nil, fmt.Errorf("upsert ingredients: %w", err)
	}

	recipe := Recipe{
		ID:          uuid.New(),
		Title:       dto.Title,
		Headline:    &dto.Headline,
		Description: &dto.Description,
		Steps:       &dto.Steps,
		Servings:    &dto.Servings,
		URL:         &dto.URL,
		Tags:        dto.Tags,
		Ingredients: recipeIngredients,
	}

	created, err := u.recipesRepo.Create(ctxWithTransaction, recipe)
	if err != nil {
		return nil, fmt.Errorf("recipesRepo create: %w", err)
	}

	commitErr := transaction.Commit()
	if commitErr != nil {
		u.logger.Errorw("error committing transaction", "error", commitErr)
		return nil, fmt.Errorf("commit transaction: %w", commitErr)
	}

	return created, nil
}

func (u UseCases) Get(ctx context.Context, recipeIDOrSlug string) (*Recipe, error) {
	idErr := uuid.Validate(recipeIDOrSlug)
	if idErr != nil {
		recipe, err := u.recipesRepo.GetBySlug(ctx, recipeIDOrSlug)
		if err != nil {
			return nil, fmt.Errorf("GetBySlug: %w", err)
		}

		return recipe, nil
	}

	recipe, err := u.recipesRepo.GetByID(ctx, recipeIDOrSlug)
	if err != nil {
		return nil, fmt.Errorf("GetById: %w", err)
	}

	return recipe, nil
}

func (u UseCases) Update(ctx context.Context, id uuid.UUID, dto *CreateUpdateRecipeDTO) (*Recipe, error) {
	transaction, err := u.transactionManager.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("transaction manager begin: %w", err)
	}

	defer func() {
		err = transaction.Rollback()
		u.logger.Errorw("error rolling back transaction", "error", err)
	}()

	ctxWithTransaction := context.WithValue(ctx, common.TransactionContextKey{}, transaction)

	recipeIngredients, err := u.ingredientsRepo.UpsertMany(ctxWithTransaction, dto.Ingredients)
	if err != nil {
		return nil, fmt.Errorf("upsert ingredients: %w", err)
	}

	recipe := Recipe{
		ID:          id,
		Title:       dto.Title,
		Headline:    &dto.Headline,
		Description: &dto.Description,
		Steps:       &dto.Steps,
		Servings:    &dto.Servings,
		URL:         &dto.URL,
		Tags:        dto.Tags,
		Ingredients: recipeIngredients,
	}

	updated, err := u.recipesRepo.Update(ctxWithTransaction, recipe)
	if err != nil {
		return nil, fmt.Errorf("recipes update: %w", err)
	}

	commitErr := transaction.Commit()
	if commitErr != nil {
		u.logger.Errorw("error committing transaction", "error", commitErr)
		return nil, fmt.Errorf("commit transaction: %w", commitErr)
	}

	return updated, nil
}

func (u UseCases) Delete(ctx context.Context, recipeID string) error {
	err := u.recipesRepo.Delete(ctx, recipeID)
	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}

	return nil
}
