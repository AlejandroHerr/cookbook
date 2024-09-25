package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/AlejandroHerr/cook-book-go/internal/common"
	"github.com/AlejandroHerr/cook-book-go/internal/common/logging"
	"github.com/AlejandroHerr/cook-book-go/internal/core/dtos"
	"github.com/AlejandroHerr/cook-book-go/internal/core/model"
	"github.com/google/uuid"
)

type RecipesRepo interface {
	GetAll(ctx context.Context) ([]*model.Recipe, error)
	Create(ctx context.Context, recipe model.Recipe) (*model.Recipe, error)
	GetByID(ctx context.Context, recipeID string) (*model.Recipe, error)
	GetBySlug(ctx context.Context, recipeSlug string) (*model.Recipe, error)
	Update(ctx context.Context, recipe model.Recipe) (*model.Recipe, error)
	Delete(ctx context.Context, recipeID string) error
}

type IngredientsRepo interface {
	UpsertMany(ctx context.Context, names []string) ([]model.Ingredient, error)
}

type RecipeUseCases struct {
	recipesRepo        RecipesRepo
	ingredientsRepo    IngredientsRepo
	transactionManager common.TransactionManager
	logger             logging.Logger
}

func NewRecipesUseCases(
	transactionManager common.TransactionManager,
	recipesRepo RecipesRepo,
	ingredientsRepo IngredientsRepo,
	logger logging.Logger,
) *RecipeUseCases {
	return &RecipeUseCases{
		transactionManager: transactionManager,
		recipesRepo:        recipesRepo,
		ingredientsRepo:    ingredientsRepo,
		logger:             logger,
	}
}

func (u RecipeUseCases) GetAll(ctx context.Context) ([]*model.Recipe, error) {
	recipes, err := u.recipesRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting all recipes: %w", err)
	}

	return recipes, nil
}

func (u RecipeUseCases) Create(ctx context.Context, dto *dtos.CreateUpdateRecipeDTO) (*model.Recipe, error) {
	transaction, err := u.transactionManager.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("transaction manager begin: %w", err)
	}

	defer func() {
		err = transaction.Rollback()
		u.logger.Errorw("error rolling back transaction", "error", err)
	}()

	ctxWithTransaction := context.WithValue(ctx, common.TransactionContextKey{}, transaction)

	ingredientsNames := make([]string, len(dto.Ingredients))

	for i, v := range dto.Ingredients {
		ingredientsNames[i] = v.Name
	}

	ingredients, err := u.ingredientsRepo.UpsertMany(ctxWithTransaction, ingredientsNames)
	if err != nil {
		return nil, fmt.Errorf("upsert ingredients: %w", err)
	}

	recipeIngredients := make([]model.RecipeIngredient, len(dto.Ingredients))

	for i, v := range dto.Ingredients {
		if v.Name != ingredients[i].Name {
			return nil, errors.New("ingredients are not in order")
		}

		recipeIngredients[i] = model.RecipeIngredient{
			ID:       ingredients[i].ID,
			Name:     ingredients[i].Name,
			Kind:     ingredients[i].Kind,
			Unit:     v.Unit,
			Quantity: v.Quantity,
		}
	}

	recipe := model.NewRecipe(
		uuid.New(),
		dto.Title,
		&dto.Headline,
		&dto.Description,
		&dto.Steps,
		&dto.Servings,
		&dto.URL,
		dto.Tags,
		recipeIngredients,
	)

	created, err := u.recipesRepo.Create(ctxWithTransaction, *recipe)
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

func (u RecipeUseCases) Get(ctx context.Context, recipeIDOrSlug string) (*model.Recipe, error) {
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

func (u RecipeUseCases) Update(ctx context.Context, id uuid.UUID, dto *dtos.CreateUpdateRecipeDTO) (*model.Recipe, error) {
	transaction, err := u.transactionManager.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("transaction manager begin: %w", err)
	}

	defer func() {
		err = transaction.Rollback()
		u.logger.Errorw("error rolling back transaction", "error", err)
	}()

	ctxWithTransaction := context.WithValue(ctx, common.TransactionContextKey{}, transaction)

	ingredientsNames := make([]string, len(dto.Ingredients))

	for i, v := range dto.Ingredients {
		ingredientsNames[i] = v.Name
	}

	ingredients, err := u.ingredientsRepo.UpsertMany(ctxWithTransaction, ingredientsNames)
	if err != nil {
		return nil, fmt.Errorf("upsert ingredients: %w", err)
	}

	recipeIngredients := make([]model.RecipeIngredient, len(dto.Ingredients))

	for i, v := range dto.Ingredients {
		if v.Name != ingredients[i].Name {
			return nil, errors.New("ingredients are not in order")
		}

		recipeIngredients[i] = model.RecipeIngredient{
			ID:       ingredients[i].ID,
			Name:     ingredients[i].Name,
			Kind:     ingredients[i].Kind,
			Unit:     v.Unit,
			Quantity: v.Quantity,
		}
	}

	recipe := model.NewRecipe(
		id,
		dto.Title,
		&dto.Headline,
		&dto.Description,
		&dto.Steps,
		&dto.Servings,
		&dto.URL,
		dto.Tags,
		recipeIngredients,
	)

	updated, err := u.recipesRepo.Update(ctxWithTransaction, *recipe)
	if err != nil {
		return nil, fmt.Errorf("recipesRepo update: %w", err)
	}

	commitErr := transaction.Commit()
	if commitErr != nil {
		u.logger.Errorw("error committing transaction", "error", commitErr)
		return nil, fmt.Errorf("commit transaction: %w", commitErr)
	}

	return updated, nil
}

func (u RecipeUseCases) Delete(ctx context.Context, recipeID string) error {
	err := u.recipesRepo.Delete(ctx, recipeID)
	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}

	return nil
}
