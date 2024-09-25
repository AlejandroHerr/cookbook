package repo

import (
	"context"
	"fmt"

	"github.com/AlejandroHerr/cook-book-go/internal/common/infra/db"
	"github.com/AlejandroHerr/cook-book-go/internal/core/model"
	"github.com/AlejandroHerr/cook-book-go/internal/core/usecases"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgIngredientsRepo struct {
	pool *pgxpool.Pool
}

var _ usecases.IngredientsRepo = (*PgIngredientsRepo)(nil)

func NewPgIngredientsRepo(pool *pgxpool.Pool) *PgIngredientsRepo {
	return &PgIngredientsRepo{
		pool: pool,
	}
}

func (repo PgIngredientsRepo) UpsertMany(ctx context.Context, names []string) ([]model.Ingredient, error) {
	client := db.GetBatcherExecutorQuerier(ctx, repo.pool)

	query := `
    INSERT INTO ingredients (name)
    VALUES ($1)
    ON CONFLICT (name) DO UPDATE 
    SET name = EXCLUDED.name
    RETURNING id, name, kind
  `

	ingredients := make([]model.Ingredient, len(names))

	for i, name := range names {
		row := client.QueryRow(ctx, query, name)

		err := row.Scan(&ingredients[i].ID, &ingredients[i].Name, &ingredients[i].Kind)
		if err != nil {
			return nil, fmt.Errorf("scanning ingredient name=%s: %w", name, db.HandlePgError(err))
		}
	}

	return ingredients, nil
}
