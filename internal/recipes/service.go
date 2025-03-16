package recipes

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/AlejandroHerr/cookbook/internal/common"
	"github.com/google/uuid"
)

type RecipesRepo interface {
	List(ctx context.Context) ([]Recipe, error)
	Create(ctx context.Context, recipe Recipe) (*Recipe, error)
	GetByID(ctx context.Context, recipeID string) (*Recipe, error)
	GetBySlug(ctx context.Context, recipeSlug string) (*Recipe, error)
	Update(ctx context.Context, recipe Recipe) (*Recipe, error)
	Delete(ctx context.Context, recipeID string) error
	GetUniqueSlug(ctx context.Context, title string) (string, error)
}

type IngredientsRepo interface {
	UpsertMany(ctx context.Context, names []CreateRecipeIngredientDTO) ([]RecipeIngredient, error)
}

type Service struct {
	recipesRepo        RecipesRepo
	ingredientsRepo    IngredientsRepo
	transactionManager common.TransactionManager
	logger             *slog.Logger
}

func NewService(
	transactionManager common.TransactionManager,
	recipesRepo RecipesRepo,
	ingredientsRepo IngredientsRepo,
	logger *slog.Logger,
) *Service {
	return &Service{
		transactionManager: transactionManager,
		recipesRepo:        recipesRepo,
		ingredientsRepo:    ingredientsRepo,
		logger:             logger,
	}
}

func (s Service) List(ctx context.Context) ([]Recipe, error) {
	recipes, err := s.recipesRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("List: %w", err)
	}

	return recipes, nil
}

func (s Service) Create(ctx context.Context, dto *CreateUpdateRecipeDTO) (*Recipe, error) {
	ctxWithUow, uow, err := s.transactionManager.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("transactionManager begin: %w", err)
	}

	defer func() {
		err := uow.Rollback(ctx)
		if err != nil {
			s.logger.WarnContext(ctx, "rolling back transaction", slog.Any("error", err))
		}
	}()

	recipeIngredients, err := s.ingredientsRepo.UpsertMany(ctx, dto.Ingredients)
	if err != nil {
		return nil, fmt.Errorf("ingredientsRepo UpsertMany: %w", err)
	}

	servings := uint(1)
	if dto.Servings != 0 {
		servings = dto.Servings
	}

	slug, err := s.recipesRepo.GetUniqueSlug(ctxWithUow, dto.Title)
	if err != nil {
		return nil, fmt.Errorf("recipesRepo GetUniqueSlug: %w", err)
	}

	recipe := Recipe{
		ID:          uuid.New(),
		Title:       dto.Title,
		Slug:        slug,
		Headline:    &dto.Headline,
		Description: &dto.Description,
		Steps:       &dto.Steps,
		PrepTime:    &dto.PrepTime,
		Servings:    servings,
		URL:         &dto.URL,
		Tags:        dto.Tags,
		Ingredients: recipeIngredients,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	created, err := s.recipesRepo.Create(ctxWithUow, recipe)
	if err != nil {
		return nil, fmt.Errorf("recipesRepo Create: %w", err)
	}

	err = uow.Commit(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "committing transaction", slog.Any("error", err))
		return nil, fmt.Errorf("uow commit: %w", err)
	}

	return created, nil
}

func (s Service) Get(ctx context.Context, recipeIDOrSlug string) (*Recipe, error) {
	idErr := uuid.Validate(recipeIDOrSlug)
	if idErr != nil {
		recipe, err := s.recipesRepo.GetBySlug(ctx, recipeIDOrSlug)
		if err != nil {
			return nil, fmt.Errorf("recipesRepo GetBySlug: %w", err)
		}

		return recipe, nil
	}

	recipe, err := s.recipesRepo.GetByID(ctx, recipeIDOrSlug)
	if err != nil {
		return nil, fmt.Errorf("recipesRepo GetById: %w", err)
	}

	return recipe, nil
}

func (s Service) Update(ctx context.Context, recipe *Recipe, dto *CreateUpdateRecipeDTO) (*Recipe, error) {
	ctxWithUow, uow, err := s.transactionManager.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("transactionManager begin: %w", err)
	}

	defer func() {
		err := uow.Rollback(ctx)
		if err != nil {
			s.logger.WarnContext(ctx, "rolling back transaction", slog.Any("error", err))
		}
	}()

	recipeIngredients, err := s.ingredientsRepo.UpsertMany(ctxWithUow, dto.Ingredients)
	if err != nil {
		return nil, fmt.Errorf("ingredientsRepo UpsertMany: %w", err)
	}

	servings := uint(1)
	if dto.Servings != 0 {
		servings = dto.Servings
	}

	slug := recipe.Slug
	if dto.Title != recipe.Title {
		slug, err = s.recipesRepo.GetUniqueSlug(ctxWithUow, dto.Title)
		if err != nil {
			return nil, fmt.Errorf("recipesRepo GetUniqueSlug: %w", err)
		}
	}

	updatedRecipe := Recipe{
		ID:          recipe.ID,
		Title:       dto.Title,
		Slug:        slug,
		Headline:    &dto.Headline,
		Description: &dto.Description,
		Steps:       &dto.Steps,
		PrepTime:    &dto.PrepTime,
		Servings:    servings,
		URL:         &dto.URL,
		Tags:        dto.Tags,
		Ingredients: recipeIngredients,
		CreatedAt:   recipe.CreatedAt,
		UpdatedAt:   time.Now(),
	}

	updated, err := s.recipesRepo.Update(ctxWithUow, updatedRecipe)
	if err != nil {
		return nil, fmt.Errorf("recipesRepo Create: %w", err)
	}

	err = uow.Commit(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "committing transaction", slog.Any("error", err))
		return nil, fmt.Errorf("uow commit: %w", err)
	}

	return updated, nil
}

func (s Service) Delete(ctx context.Context, recipeID string) error {
	err := s.recipesRepo.Delete(ctx, recipeID)
	if err != nil {
		return fmt.Errorf("recipesRepo Delete: %w", err)
	}

	return nil
}
