package recipes

import (
	"context"
	"fmt"

	"github.com/AlejandroHerr/cook-book-go/internal/common/infra/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgIngredientsRepo struct {
	pool *pgxpool.Pool
}

var _ IngredientsRepo = (*PgIngredientsRepo)(nil)

func NewPgIngredientsRepo(pool *pgxpool.Pool) *PgIngredientsRepo {
	return &PgIngredientsRepo{
		pool: pool,
	}
}

func (repo PgIngredientsRepo) UpsertMany(ctx context.Context, ingredients []CreateRecipeIngredientDTO) ([]RecipeIngredient, error) {
	client := db.GetBatcherExecutorQuerier(ctx, repo.pool)

	query := `
    INSERT INTO ingredients (name)
    VALUES ($1)
    ON CONFLICT (name) DO UPDATE 
    SET name = EXCLUDED.name
    RETURNING id, name, kind
  `

	recipeIngredients := make([]RecipeIngredient, len(ingredients))

	for i, ingredient := range ingredients {
		row := client.QueryRow(ctx, query, ingredient.Name)

		err := row.Scan(&recipeIngredients[i].ID, &recipeIngredients[i].Name, &recipeIngredients[i].Kind)
		if err != nil {
			return nil, fmt.Errorf("scanning ingredient name=%s: %w", ingredient.Name, db.HandlePgError(err))
		}

		recipeIngredients[i].Quantity = ingredient.Quantity
		recipeIngredients[i].Unit = ingredient.Unit
	}

	return recipeIngredients, nil
}
