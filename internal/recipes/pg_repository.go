package recipes

import (
	"context"
	"errors"
	"fmt"

	"github.com/AlejandroHerr/cook-book-go/internal/core"
	"github.com/AlejandroHerr/cook-book-go/internal/ingredients"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgRepository struct {
	pool *pgxpool.Pool
}

func NewPgRepository(pool *pgxpool.Pool) *PgRepository {
	return &PgRepository{
		pool: pool,
	}
}

func (r PgRepository) Find(id uuid.UUID) (*Recipe, error) {
	return r.findBy("id", id.String())
}

func (r PgRepository) FindByName(name string) (*Recipe, error) {
	return r.findBy("name", name)
}

func (r PgRepository) findBy(field string, value string) (*Recipe, error) {
	row := r.pool.QueryRow(context.TODO(), "SELECT * from recipes WHERE "+field+" = $1", value)
	recipeDB := struct {
		ID          uuid.UUID
		Name        string
		Description *string
		URL         *string
	}{}

	err := row.Scan(&recipeDB.ID, &recipeDB.Name, &recipeDB.Description, &recipeDB.URL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &core.NotFoundError{Field: field, Value: value, Err: err}
		}

		return nil, fmt.Errorf("error scanning recipe: %w", err)
	}

	ingredientsList, err := r.loadIngredients(recipeDB.ID)
	if err != nil {
		return nil, fmt.Errorf("error loading ingredients: %w", err)
	}

	recipe := New(recipeDB.ID, recipeDB.Name, recipeDB.Description, recipeDB.URL, ingredientsList)

	return recipe, nil
}

func (r PgRepository) loadIngredients(recipeID uuid.UUID) ([]ingredients.Ingredient, error) {
	query := `
    select i.id, i.name, i.kind 
    from recipe_ingredients ri
    inner JOIN ingredients i on ri.ingredient_id = i.id 
    where ri.recipe_id = $1
  `

	rows, err := r.pool.Query(
		context.TODO(),
		query,
		recipeID,
	)
	if err != nil {
		return nil, fmt.Errorf("error loading ingredients: %w", err)
	}

	ingredientsList := make([]ingredients.Ingredient, 0)

	for rows.Next() {
		ingredientDB := struct {
			ID   uuid.UUID
			Name string
			Kind *string
		}{}

		err = rows.Scan(&ingredientDB.ID, &ingredientDB.Name, &ingredientDB.Kind)
		if err != nil {
			return nil, fmt.Errorf("error scanning ingredient: %w", err)
		}

		ingredient := ingredients.New(ingredientDB.ID, ingredientDB.Name, ingredientDB.Kind)
		ingredientsList = append(ingredientsList, *ingredient)
	}

	return ingredientsList, nil
}

func (r PgRepository) Save(recipe *Recipe) error {
	values := pgx.NamedArgs{
		"id":          recipe.ID(),
		"name":        recipe.Name(),
		"description": recipe.Description(),
		"url":         recipe.URL(),
	}

	_, err := r.pool.Exec(
		context.TODO(),
		"INSERT INTO recipes (id, name, description, url) VALUES (@id,@name,@description,@url)",
		values,
	)
	if err != nil {
		return core.HandleError(err)
	}

	for _, ingredient := range recipe.Ingredients() {
		_, err := r.pool.Exec(
			context.TODO(),
			"INSERT INTO recipe_ingredients (recipe_id, ingredient_id) VALUES ($1, $2)",
			recipe.ID(),
			ingredient.ID(),
		)
		if err != nil {
			return core.HandleError(err)
		}
	}

	return nil
}
