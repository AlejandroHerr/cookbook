package recipes

import (
	"context"
	"fmt"

	"github.com/AlejandroHerr/cookbook/internal/common/infra/db"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
)

type PgRecipesRepo struct {
	pgxDB db.PGXDB
}

var _ RecipesRepo = (*PgRecipesRepo)(nil)

func MakePgRecipesRepository(pgxDB db.PGXDB) *PgRecipesRepo {
	return &PgRecipesRepo{
		pgxDB: pgxDB,
	}
}

func (repo PgRecipesRepo) GetAll(ctx context.Context) ([]Recipe, error) {
	query := `
    SELECT
      id, title, slug, headline, description, steps, prep_time, servings, url, tags, created_at, updated_at
    FROM
      recipes
  `

	rows, err := repo.pgxDB.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("quering recipes: %w", db.HandlePgError(err))
	}
	defer rows.Close()

	recipes := make([]Recipe, 0)

	for rows.Next() {
		recipe := Recipe{} //nolint: exhaustruct

		if err = rows.Scan(
			&recipe.ID,
			&recipe.Title,
			&recipe.Slug,
			&recipe.Headline,
			&recipe.Description,
			&recipe.Steps,
			&recipe.PrepTime,
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
      id, title, slug, headline, description, steps, prep_time, servings, url, tags, created_at, updated_at
    FROM
      recipes      
    WHERE ` + field + ` = $1
  `
	row := repo.pgxDB.QueryRow(ctx, query, value)

	recipe := new(Recipe)

	if err := row.Scan(
		&recipe.ID,
		&recipe.Title,
		&recipe.Slug,
		&recipe.Headline,
		&recipe.Description,
		&recipe.Steps,
		&recipe.PrepTime,
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

	rows, err := repo.pgxDB.Query(ctx, query, recipeID)
	if err != nil {
		return nil, fmt.Errorf("querying recipe_ingredients: %w", err)
	}
	defer rows.Close()

	ingredients := make([]RecipeIngredient, 0)

	for rows.Next() {
		ri := RecipeIngredient{} //nolint: exhaustruct

		err = rows.Scan(&ri.ID, &ri.Name, &ri.Kind, &ri.Unit, &ri.Quantity)
		if err != nil {
			return nil, fmt.Errorf("scanning recipe_ingredients row: %w", err)
		}

		ingredients = append(ingredients, ri)
	}

	return ingredients, nil
}

func (repo PgRecipesRepo) Create(ctx context.Context, recipe Recipe) (*Recipe, error) {
	pgxDB := db.GetPGXDB(ctx, repo.pgxDB)

	sql := `
    INSERT INTO
      recipes (id, title, slug, headline, description, steps, prep_time, servings, url, tags)
    VALUES 
      (@id, @title, @slug, @headline, @description, @steps, @prep_time, @servings, @url, @tags)
    RETURNING
      id, title, slug, headline, description, steps, prep_time, servings, url, tags, created_at, updated_at
  `
	values := pgx.NamedArgs{
		"id":          recipe.ID,
		"title":       recipe.Title,
		"slug":        recipe.Slug,
		"headline":    recipe.Headline,
		"description": recipe.Description,
		"steps":       recipe.Steps,
		"prep_time":   recipe.PrepTime,
		"servings":    recipe.Servings,
		"url":         recipe.URL,
		"tags":        recipe.Tags,
	}

	row := pgxDB.QueryRow(ctx, sql, values)

	createdRecipe := new(Recipe)
	if err := row.Scan(
		&createdRecipe.ID,
		&createdRecipe.Title,
		&createdRecipe.Slug,
		&createdRecipe.Headline,
		&createdRecipe.Description,
		&createdRecipe.Steps,
		&createdRecipe.PrepTime,
		&createdRecipe.Servings,
		&createdRecipe.URL,
		&createdRecipe.Tags,
		&createdRecipe.CreatedAt,
		&createdRecipe.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("executing insert recipe query: %w", db.HandlePgError(err))
	}

	err := repo.insertRecipeIngredients(ctx, pgxDB, recipe.ID, recipe.Ingredients)
	if err != nil {
		return nil, fmt.Errorf("inserting recipe ingredients: %w", db.HandlePgError(err))
	}

	createdRecipe.Ingredients = recipe.Ingredients

	return createdRecipe, nil
}

func (repo PgRecipesRepo) Update(ctx context.Context, recipe Recipe) (*Recipe, error) {
	pgxDB := db.GetPGXDB(ctx, repo.pgxDB)

	sql := `
    UPDATE 
      recipes 
    SET
      title = @title,
      slug = @slug,
      headline = @headline,
      description = @description,
      steps = @steps,
      prep_time = @prep_time,
      servings = @servings,
      url = @url,
      tags = @tags
    WHERE 
      id = @id
    RETURNING
      id, title, slug, headline, description, steps, prep_time, servings, url, tags, created_at, updated_at

  `
	values := pgx.NamedArgs{
		"id":          recipe.ID,
		"title":       recipe.Title,
		"headline":    recipe.Headline,
		"description": recipe.Description,
		"steps":       recipe.Steps,
		"prep_time":   recipe.PrepTime,
		"servings":    recipe.Servings,
		"url":         recipe.URL,
		"tags":        recipe.Tags,
		"slug":        recipe.Slug,
	}

	row := pgxDB.QueryRow(ctx, sql, values)

	updatedRecipe := new(Recipe)

	if err := row.Scan(
		&updatedRecipe.ID,
		&updatedRecipe.Title,
		&updatedRecipe.Slug,
		&updatedRecipe.Headline,
		&updatedRecipe.Description,
		&updatedRecipe.Steps,
		&updatedRecipe.PrepTime,
		&updatedRecipe.Servings,
		&updatedRecipe.URL,
		&updatedRecipe.Tags,
		&updatedRecipe.CreatedAt,
		&updatedRecipe.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("executing update recipe query: %w", db.HandlePgError(err))
	}

	_, err := pgxDB.Exec(
		ctx,
		"DELETE FROM recipe_ingredients WHERE recipe_id = $1",
		recipe.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("deleting recipe ingredients: %w", db.HandlePgError(err))
	}

	err = repo.insertRecipeIngredients(ctx, pgxDB, recipe.ID, recipe.Ingredients)
	if err != nil {
		return nil, fmt.Errorf("inserting recipe ingredients: %w", db.HandlePgError(err))
	}

	updatedRecipe.Ingredients = recipe.Ingredients

	return updatedRecipe, nil
}

func (repo PgRecipesRepo) Delete(ctx context.Context, recipeID string) error {
	pgxDB := db.GetPGXDB(ctx, repo.pgxDB)

	sql := "DELETE FROM recipes WHERE id = $1;"

	_, err := pgxDB.Exec(ctx, sql, recipeID)
	if err != nil {
		return fmt.Errorf("executing delete recipe query: %w", db.HandlePgError(err))
	}

	return nil
}

func (repo PgRecipesRepo) insertRecipeIngredients(ctx context.Context, pgxDB db.PGXDB, recipeID uuid.UUID, ingredients []RecipeIngredient) error { //nolint:lll
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

	batchResult := pgxDB.SendBatch(ctx, batch)
	defer batchResult.Close()

	for i := range batch.Len() {
		_, err := batchResult.Exec()
		if err != nil {
			return fmt.Errorf("error executing batched query at index %d: %w", i, err)
		}
	}

	return nil
}

func (repo PgRecipesRepo) GetUniqueSlug(ctx context.Context, title string) (string, error) {
	pgxDB := db.GetPGXDB(ctx, repo.pgxDB)

	slug := slug.Make(title)

	query := `
    SELECT slug FROM recipes 
    WHERE slug = $1 
      OR slug ~ ($1 || '-[0-9]+$');
  `

	rows, err := pgxDB.Query(ctx, query, slug)
	if err != nil {
		return "", fmt.Errorf("querying existing slugs: %w", err)
	}

	existingSlugs := make(map[string]bool)

	for rows.Next() {
		var existingSlug string
		if err = rows.Scan(&existingSlug); err != nil {
			return "", fmt.Errorf("scanning existing slug: %w", err)
		}

		existingSlugs[existingSlug] = true
	}

	if !existingSlugs[slug] {
		return slug, nil
	}

	for i := 1; i < 1000; i++ {
		newSlug := fmt.Sprintf("%s-%d", slug, i)
		if !existingSlugs[newSlug] {
			return newSlug, nil
		}
	}

	return "", fmt.Errorf("could not find a unique slug for %s", title)
}
