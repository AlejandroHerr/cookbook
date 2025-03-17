package pg

import (
	"context"
	"fmt"

	"github.com/AlejandroHerr/cookbook/internal/common"
	pgdb "github.com/AlejandroHerr/cookbook/internal/common/pg"
	"github.com/AlejandroHerr/cookbook/internal/common/utils"
	"github.com/AlejandroHerr/cookbook/internal/recipes"
	"github.com/AlejandroHerr/cookbook/internal/sql"
)

var _ recipes.RecipesRepo = (*RecipesRepo)(nil)

type RecipesRepo struct {
	dbtx    sql.DBTX
	queries *sql.Queries
}

func NewRecipesRepo(dbtx sql.DBTX) *RecipesRepo {
	return &RecipesRepo{
		dbtx:    dbtx,
		queries: sql.New(),
	}
}

func (r RecipesRepo) List(ctx context.Context) ([]recipes.Recipe, error) {
	dbtx := pgdb.GetDBTX(ctx, r.dbtx)

	list, err := r.queries.ListRecipes(ctx, dbtx)
	if err != nil {
		return nil, fmt.Errorf("query GetRecipes: %w", err)
	}

	recipesList := make([]recipes.Recipe, len(list))

	for i, r := range list {
		recipe, err := r.ToModel()
		if err != nil {
			return nil, fmt.Errorf("error converting recipe to model: %w", err)
		}

		recipesList[i] = recipe
	}

	return recipesList, nil
}

func (r RecipesRepo) GetByID(ctx context.Context, id string) (*recipes.Recipe, error) {
	recipe, err := r.getBy(ctx, "id", id)
	if err != nil {
		return nil, fmt.Errorf("getBy: %w", err)
	}

	return recipe, nil
}

func (r RecipesRepo) GetBySlug(ctx context.Context, slug string) (*recipes.Recipe, error) {
	recipe, err := r.getBy(ctx, "slug", slug)
	if err != nil {
		return nil, fmt.Errorf("getBy: %w", err)
	}

	return recipe, nil
}

func (r RecipesRepo) Create(ctx context.Context, recipe recipes.Recipe) (*recipes.Recipe, error) {
	dbtx := pgdb.GetDBTX(ctx, r.dbtx)

	servings, err := utils.UintToInt32Safe(recipe.Servings)
	if err != nil {
		return nil, fmt.Errorf("converting servings from uint to int32: %w", err)
	}

	prepTime, err := utils.UintPtrToInt32PtrSafe(recipe.PrepTime)
	if err != nil {
		return nil, fmt.Errorf("converting prepTime from uint to int32: %w", err)
	}

	created, err := r.queries.CreateRecipe(ctx, dbtx, sql.CreateRecipeParams{
		ID:          recipe.ID.String(),
		Title:       recipe.Title,
		Slug:        recipe.Slug,
		Headline:    recipe.Headline,
		Description: recipe.Description,
		Steps:       recipe.Steps,
		Servings:    &servings,
		PrepTime:    prepTime,
		Url:         recipe.URL,
		Tags:        recipe.Tags,
	})
	if err != nil {
		return nil, fmt.Errorf("query CreateRecipe: %w", pgdb.HandleExecError(err))
	}

	for _, ri := range recipe.Ingredients {
		_, err = r.queries.CreateRecipeIngredient(ctx, dbtx, sql.CreateRecipeIngredientParams{
			RecipeID:     recipe.ID.String(),
			IngredientID: ri.ID.String(),
			Unit:         ri.Unit.String(),
			Quantity:     ri.Quantity,
		})
		if err != nil {
			return nil, fmt.Errorf("query CreateRecipeIngredient for %s: %w", ri.Name, pgdb.HandleExecError(err))
		}
	}

	createdRecipe, err := created.ToModel()
	if err != nil {
		return nil, fmt.Errorf("converting recipe to model: %w", err)
	}

	createdRecipe.Ingredients = recipe.Ingredients

	return &createdRecipe, nil
}

func (r RecipesRepo) Update(ctx context.Context, recipe recipes.Recipe) (*recipes.Recipe, error) {
	dbtx := pgdb.GetDBTX(ctx, r.dbtx)

	servings, err := utils.UintToInt32Safe(recipe.Servings)
	if err != nil {
		return nil, fmt.Errorf("converting servings from uint to int32: %w", err)
	}

	prepTime, err := utils.UintPtrToInt32PtrSafe(recipe.PrepTime)
	if err != nil {
		return nil, fmt.Errorf("converting prepTime from uint to int32: %w", err)
	}

	updated, err := r.queries.UpdateRecipe(ctx, dbtx, sql.UpdateRecipeParams{
		ID:          recipe.ID.String(),
		Title:       recipe.Title,
		Slug:        recipe.Slug,
		Headline:    recipe.Headline,
		Description: recipe.Description,
		Steps:       recipe.Steps,
		Servings:    &servings,
		PrepTime:    prepTime,
		Url:         recipe.URL,
		Tags:        recipe.Tags,
	})
	if err != nil {
		return nil, fmt.Errorf("query UpdateRecipe: %w", pgdb.HandleExecError(err))
	}

	_, err = r.queries.DeleteRecipeIngredients(ctx, dbtx, recipe.ID.String())
	if err != nil {
		return nil, fmt.Errorf("query DeleteRecipeIngredients: %w", pgdb.HandleExecError(err))
	}

	for _, ri := range recipe.Ingredients {
		_, err = r.queries.CreateRecipeIngredient(ctx, dbtx, sql.CreateRecipeIngredientParams{
			RecipeID:     recipe.ID.String(),
			IngredientID: ri.ID.String(),
			Unit:         ri.Unit.String(),
			Quantity:     ri.Quantity,
		})
		if err != nil {
			return nil, fmt.Errorf("query CreateRecipeIngredient for %s: %w", ri.Name, pgdb.HandleExecError(err))
		}
	}

	updatedRecipe, err := updated.ToModel()
	if err != nil {
		return nil, fmt.Errorf("converting recipe to model: %w", err)
	}

	updatedRecipe.Ingredients = recipe.Ingredients

	return &updatedRecipe, nil
}

func (r RecipesRepo) Delete(ctx context.Context, recipeID string) error {
	dbtx := pgdb.GetDBTX(ctx, r.dbtx)

	_, err := r.queries.DeleteRecipe(ctx, dbtx, recipeID)
	if err != nil {
		return fmt.Errorf("executing delete recipe query: %w", pgdb.HandleExecError(err))
	}

	return nil
}

func (r RecipesRepo) getBy(ctx context.Context, field string, value string) (*recipes.Recipe, error) {
	dbtx := pgdb.GetDBTX(ctx, r.dbtx)

	var recipe sql.Recipe

	var err error

	switch field {
	case "id":
		recipe, err = r.queries.GetRecipeByID(ctx, dbtx, value)
	case "slug":
		recipe, err = r.queries.GetRecipeBySlug(ctx, dbtx, value)
	default:
		return nil, fmt.Errorf("get by %s not implemented", field)
	}

	if err != nil {
		return nil, fmt.Errorf("query getBy for key %s value %s: %w", field, value, pgdb.HandleScanError(err))
	}

	ingredients, err := r.queries.GetRecipeIngredients(ctx, dbtx, recipe.ID)
	if err != nil {
		return nil, fmt.Errorf("query GetRecipeIngredients for recipe id %s: %w", recipe.ID, &common.UnexpectedError{Err: err})
	}

	model, err := recipe.ToModel()
	if err != nil {
		return nil, fmt.Errorf("converting recipe to model: %w", err)
	}

	recipeIngredients := make([]recipes.RecipeIngredient, len(ingredients))

	for i, ingredient := range ingredients {
		recipeIngredient, err := ingredient.ToModel()
		if err != nil {
			return nil, fmt.Errorf("error converting recipe ingredient to model: %w", err)
		}

		recipeIngredients[i] = recipeIngredient
	}

	model.Ingredients = recipeIngredients

	return &model, nil
}
