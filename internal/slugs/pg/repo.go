package pg

import (
	"context"
	"fmt"

	pgcommon "github.com/AlejandroHerr/cookbook/internal/common/pg"
)

type Repo struct {
	db pgcommon.DBTX
}

func NewRepo(db pgcommon.DBTX) *Repo {
	return &Repo{db: db}
}

func (r Repo) GetSlugs(ctx context.Context, entity string, slug string) ([]string, error) {
	rows, err := r.db.Query(ctx, `
SELECT slug FROM `+entity+`
WHERE slug = $1 
OR slug ~ ($1 || '-[0-9]+$')
`,
		slug)
	if err != nil {
		return nil, fmt.Errorf("query GetSlugs for entity %s and slug %s: %w", entity, slug, err)
	}

	defer rows.Close()

	var items []string

	for rows.Next() {
		var s string
		if err = rows.Scan(&s); err != nil {
			return nil, fmt.Errorf("scanning slug for entity %s and slug %s: %w", entity, slug, err)
		}

		items = append(items, s)
	}

	return items, nil
}
