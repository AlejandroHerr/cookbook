package recipes

import (
	"context"
	"fmt"

	"github.com/AlejandroHerr/cookbook/internal/common/infra/db"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgRecipesRepo struct {
	pool *pgxpool.Pool
}

var _ RecipesRepo = (*PgRecipesRepo)(nil)

func MakePgRecipesRepository(pool *pgxpool.Pool) *PgRecipesRepo {
	return &PgRecipesRepo{
		pool: pool,
	}
}

func (repo PgRecipesRepo) GetAll(ctx context.Context) ([]*Recipe, error) {
	query := `
    SELECT
      id, title, headline, description, steps, servings, url, tags, created_at, updated_at
    FROM
      recipes
  `

	rows, err := repo.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("quering recipes: %w", db.HandlePgError(err))
	}
	defer rows.Close()

	recipes := make([]*Recipe, 0)

	for rows.Next() {
		recipe := new(Recipe)

		if err := rows.Scan(
			&recipe.ID,
			&recipe.Title,
			&recipe.Headline,
			&recipe.Description,
			&recipe.Steps,
			&recipe.Servings,
			&recipe.URL,
			&recipe.Tags,
			&recipe.CreatedAt,
			&recipe.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning recipe row: %w", err)
		}

		recipes = append(
			recipes,
			recipe,
		)
	}

	return recipes, nil
}

func (repo PgRecipesRepo) GetByID(ctx context.Context, recipeID string) (*Recipe, error) {
	return repo.get(ctx, "id", recipeID)
}

func (repo PgRecipesRepo) GetBySlug(ctx context.Context, slug string) (*Recipe, error) {
	return repo.get(ctx, "slug", slug)
}

func (repo PgRecipesRepo) get(ctx context.Context, field string, value string) (*Recipe, error) {
	query := `
    SELECT 
      id, title, headline, description, steps, servings, url, tags, created_at, updated_at
    FROM
      recipes      
    WHERE ` + field + ` = $1
  `
	row := repo.pool.QueryRow(ctx, query, value)

	recipe := new(Recipe)

	if err := row.Scan(
		&recipe.ID,
		&recipe.Title,
		&recipe.Headline,
		&recipe.Description,
		&recipe.Steps,
		&recipe.Servings,
		&recipe.URL,
		&recipe.Tags,
		&recipe.CreatedAt,
		&recipe.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("scaning recipe %s=%s: %w", field, value, db.HandlePgError(err))
	}

	ingredients, err := repo.getRecipeIngredients(ctx, recipe.ID)
	if err != nil {
		return nil, fmt.Errorf("getting recipe_ingredients: %w", err)
	}

	recipe.Ingredients = ingredients

	return recipe, nil
}

func (repo PgRecipesRepo) getRecipeIngredients(ctx context.Context, recipeID uuid.UUID) ([]RecipeIngredient, error) {
	query := `
    SELECT 
      i.id, i.name, i.kind, ri.unit, ri.quantity
    FROM 
      recipe_ingredients ri 
    LEFT JOIN 
      ingredients i ON i.id = ri.ingredient_id 
    WHERE 
      ri.recipe_id = $1`

	rows, err := repo.pool.Query(ctx, query, recipeID)
	if err != nil {
		return nil, fmt.Errorf("querying recipe_ingredients: %w", err)
	}
	defer rows.Close()

	ingredients := make([]RecipeIngredient, 0)

	for rows.Next() {
		ri := RecipeIngredient{} //nolint: exhaustruct

		err := rows.Scan(&ri.ID, &ri.Name, &ri.Kind, &ri.Unit, &ri.Quantity)
		if err != nil {
			return nil, fmt.Errorf("scanning recipe_ingredients row: %w", err)
		}

		ingredients = append(ingredients, ri)
	}

	return ingredients, nil
}

func (repo PgRecipesRepo) Create(ctx context.Context, recipe Recipe) (*Recipe, error) {
	executor := db.GetBatcherExecutorQuerier(ctx, repo.pool)

	sql := `
    INSERT INTO
      recipes (id, title, headline, description, steps, servings, url, tags, slug)
    VALUES 
      (@id, @title, @headline,@description, @steps,@servings,@url, @tags, @slug)
    RETURNING
      id, title, headline, description, steps, servings, url, tags, created_at, updated_at
  `
	values := pgx.NamedArgs{
		"id":          recipe.ID,
		"title":       recipe.Title,
		"headline":    recipe.Headline,
		"description": recipe.Description,
		"steps":       recipe.Steps,
		"servings":    recipe.Servings,
		"url":         recipe.URL,
		"tags":        recipe.Tags,
		"slug":        recipe.Slug(),
	}

	row := executor.QueryRow(ctx, sql, values)

	createdRecipe := new(Recipe)
	if err := row.Scan(
		&createdRecipe.ID,
		&createdRecipe.Title,
		&createdRecipe.Headline,
		&createdRecipe.Description,
		&createdRecipe.Steps,
		&createdRecipe.Servings,
		&createdRecipe.URL,
		&createdRecipe.Tags,
		&createdRecipe.CreatedAt,
		&createdRecipe.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("executing insert recipe query: %w", db.HandlePgError(err))
	}

	err := repo.insertRecipeIngredients(ctx, executor, recipe.ID, recipe.Ingredients)
	if err != nil {
		return nil, fmt.Errorf("inserting recipe ingredients: %w", db.HandlePgError(err))
	}

	createdRecipe.Ingredients = recipe.Ingredients

	return createdRecipe, nil
}

func (repo PgRecipesRepo) Update(ctx context.Context, recipe Recipe) (*Recipe, error) {
	executor := db.GetBatcherExecutorQuerier(ctx, repo.pool)

	sql := `
    UPDATE 
      recipes 
    SET
      title = @title,
      headline = @headline,
      description = @description,
      steps = @steps,
      servings = @servings,
      url = @url,
      tags = @tags,
      slug = @slug
    WHERE 
      id = @id
    RETURNING
      id, title, headline, description, steps, servings, url, tags, created_at, updated_at

  `
	values := pgx.NamedArgs{
		"id":          recipe.ID,
		"title":       recipe.Title,
		"headline":    recipe.Headline,
		"description": recipe.Description,
		"steps":       recipe.Steps,
		"servings":    recipe.Servings,
		"url":         recipe.URL,
		"tags":        recipe.Tags,
		"slug":        recipe.Slug(),
	}

	row := executor.QueryRow(ctx, sql, values)

	updatedRecipe := new(Recipe)

	if err := row.Scan(
		&updatedRecipe.ID,
		&updatedRecipe.Title,
		&updatedRecipe.Headline,
		&updatedRecipe.Description,
		&updatedRecipe.Steps,
		&updatedRecipe.Servings,
		&updatedRecipe.URL,
		&updatedRecipe.Tags,
		&updatedRecipe.CreatedAt,
		&updatedRecipe.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("executing update recipe query: %w", db.HandlePgError(err))
	}

	_, err := executor.Exec(
		ctx,
		"DELETE FROM recipe_ingredients WHERE recipe_id = $1",
		recipe.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("deleting recipe ingredients: %w", db.HandlePgError(err))
	}

	err = repo.insertRecipeIngredients(ctx, executor, recipe.ID, recipe.Ingredients)
	if err != nil {
		return nil, fmt.Errorf("inserting recipe ingredients: %w", db.HandlePgError(err))
	}

	updatedRecipe.Ingredients = recipe.Ingredients

	return updatedRecipe, nil
}

func (repo PgRecipesRepo) Delete(ctx context.Context, recipeID string) error {
	executor := db.GetBatcherExecutorQuerier(ctx, repo.pool)

	sql := "DELETE FROM recipes WHERE id = $1;"

	_, err := executor.Exec(ctx, sql, recipeID)
	if err != nil {
		return fmt.Errorf("executing delete recipe query: %w", db.HandlePgError(err))
	}

	return nil
}

func (repo PgRecipesRepo) insertRecipeIngredients(ctx context.Context, executor db.BatcherExecutorQuerier, recipeID uuid.UUID, ingredients []RecipeIngredient) error {
	batch := &pgx.Batch{} //nolint: exhaustruct
	recipeIngredientsQuery := `
    INSERT INTO
      recipe_ingredients (recipe_id, ingredient_id, unit, quantity) 
    VALUES 
      (@recipe_id, @ingredient_id, @unit, @quantity)
  `

	for _, ingredient := range ingredients {
		batch.Queue(recipeIngredientsQuery, pgx.NamedArgs{
			"recipe_id":     recipeID,
			"ingredient_id": ingredient.ID,
			"unit":          ingredient.Unit,
			"quantity":      ingredient.Quantity,
		})
	}

	batchResult := executor.SendBatch(ctx, batch)
	defer batchResult.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := batchResult.Exec()
		if err != nil {
			return err //nolint: wrapcheck
		}
	}

	return nil
}
