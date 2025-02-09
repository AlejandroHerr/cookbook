package suggestions

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgSuggestionsRepo struct {
	pool *pgxpool.Pool
}

var _ Repo = (*PgSuggestionsRepo)(nil)

func NewPgSuggestionsRepository(pool *pgxpool.Pool) *PgSuggestionsRepo {
	return &PgSuggestionsRepo{
		pool: pool,
	}
}

func (repo PgSuggestionsRepo) FindAllTags(ctx context.Context) ([]Option, error) {
	query := `
    SELECT DISTINCT 
      tag AS unique_tag
	  FROM 
      recipes,
	  LATERAL UNNEST(tags) AS tag
    ORDER BY
      tag ASC;
  `

	return repo.findOptions(ctx, query)
}

func (repo PgSuggestionsRepo) FindMatchingTags(ctx context.Context, search string) ([]Option, error) {
	query := `
    SELECT
      tag 
    FROM (
      SELECT DISTINCT 
        tag, similarity(tag, $1) AS score
      FROM 
        recipes,
      LATERAL UNNEST(tags) AS tag
      WHERE
        tag ILIKE '%' || $1 || '%'
      ORDER BY score DESC
    )
    WHERE
      score >= 0.01;
  `

	return repo.findOptions(ctx, query, search)
}

func (repo PgSuggestionsRepo) FindMatchingIngredients(ctx context.Context, search string) ([]Option, error) {
	query := `
    SELECT name 
    FROM (
      SELECT
        name, similarity(name, $1) as score
      FROM
        ingredients
      WHERE 
        ingredients ILIKE '%' || $1 || '%'
      ORDER BY 
        score DESC
    )
    WHERE
      score >= 0.01;
  `

	return repo.findOptions(ctx, query, search)
}

func (repo PgSuggestionsRepo) FindAllIngredients(ctx context.Context) ([]Option, error) {
	query := `
    SELECT
      name
    FROM
      ingredients
    ORDER BY
      name ASC
  `

	return repo.findOptions(ctx, query)
}

func (r PgSuggestionsRepo) findOptions(ctx context.Context, query string, args ...any) ([]Option, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}
	defer rows.Close()

	options := make([]Option, 0)
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		options = append(options, Option{
			Label: n,
			Value: n,
		})
	}

	return options, nil
}
