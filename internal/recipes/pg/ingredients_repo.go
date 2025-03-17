package pg

import (
	"context"
	"fmt"

	"github.com/AlejandroHerr/cookbook/internal/common"
	pgdb "github.com/AlejandroHerr/cookbook/internal/common/pg"
	"github.com/AlejandroHerr/cookbook/internal/recipes"
	"github.com/AlejandroHerr/cookbook/internal/sql"
	"github.com/google/uuid"
)

var _ recipes.IngredientsRepo = (*IngredientsRepo)(nil)

type IngredientsRepo struct {
	dbtx    sql.DBTX
	queries *sql.Queries
}

func NewIngredientsRepo(dbtx pgdb.DBTX) *IngredientsRepo {
	return &IngredientsRepo{
		dbtx:    dbtx,
		queries: sql.New(),
	}
}

func (r IngredientsRepo) UpsertMany(ctx context.Context, ingredients []recipes.CreateRecipeIngredientDTO) ([]recipes.RecipeIngredient, error) {
	dbtx := pgdb.GetDBTX(ctx, r.dbtx)

	recipeIngredients := make([]recipes.RecipeIngredient, len(ingredients))

	for idx, ingredient := range ingredients {
		i, err := r.queries.UpsertIngredient(ctx, dbtx, ingredient.Name)
		if err != nil {
			return nil, fmt.Errorf("query UpsertIngredient for name %s: %w", ingredient.Name, &common.UnexpectedError{Err: err})
		}

		parsedID, err := uuid.Parse(i.ID)
		if err != nil {
			return nil, fmt.Errorf("parsing ingredient ID %si for name %s: %w", i.ID, i.Name, &common.UnexpectedError{Err: err})
		}

		recipeIngredients[idx] = recipes.RecipeIngredient{
			ID:       parsedID,
			Name:     i.Name,
			Kind:     i.Kind,
			Quantity: ingredient.Quantity,
			Unit:     ingredient.Unit,
		}
	}

	return recipeIngredients, nil
}
