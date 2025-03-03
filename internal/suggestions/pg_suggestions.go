package suggestions

import (
	"context"
	"fmt"

	"github.com/AlejandroHerr/cookbook/internal/common/infra/db"
)

type PgSuggestionsRepo struct {
	pgxDB db.PGXDB
}

var _ Repo = (*PgSuggestionsRepo)(nil)

func MakePgSuggestionsRepo(pgxDB db.PGXDB) *PgSuggestionsRepo {
	return &PgSuggestionsRepo{
		pgxDB: pgxDB,
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
        name ILIKE '%' || $1 || '%'
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

func (repo PgSuggestionsRepo) findOptions(ctx context.Context, query string, args ...any) ([]Option, error) {
	rows, err := repo.pgxDB.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}
	defer rows.Close()

	options := make([]Option, 0)

	for rows.Next() {
		var n string
		if err = rows.Scan(&n); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		options = append(options, Option{
			Label: n,
			Value: n,
		})
	}

	return options, nil
}
