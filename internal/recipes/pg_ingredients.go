package recipes

import (
	"context"
	"fmt"

	"github.com/AlejandroHerr/cookbook/internal/common/infra/db"
)

type PgIngredientsRepo struct {
	pgxDB db.PGXDB
}

var _ IngredientsRepo = (*PgIngredientsRepo)(nil)

func MakePgIngredientsRepo(pgxDB db.PGXDB) *PgIngredientsRepo {
	return &PgIngredientsRepo{
		pgxDB: pgxDB,
	}
}

func (repo PgIngredientsRepo) UpsertMany(ctx context.Context, ingredients []CreateRecipeIngredientDTO) ([]RecipeIngredient, error) { //nolint:lll
	pgxDB := db.GetPGXDB(ctx, repo.pgxDB)

	query := `
    INSERT INTO ingredients (name)
    VALUES ($1)
    ON CONFLICT (name) DO UPDATE 
    SET name = EXCLUDED.name
    RETURNING id, name, kind
  `

	recipeIngredients := make([]RecipeIngredient, len(ingredients))

	for i, ingredient := range ingredients {
		row := pgxDB.QueryRow(ctx, query, ingredient.Name)

		err := row.Scan(&recipeIngredients[i].ID, &recipeIngredients[i].Name, &recipeIngredients[i].Kind)
		if err != nil {
			return nil, fmt.Errorf("scanning ingredient name=%s: %w", ingredient.Name, db.HandlePgError(err))
		}

		recipeIngredients[i].Quantity = ingredient.Quantity
		recipeIngredients[i].Unit = ingredient.Unit
	}

	return recipeIngredients, nil
}
