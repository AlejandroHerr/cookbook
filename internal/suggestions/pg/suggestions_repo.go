package pg

import (
	"context"
	"fmt"

	"github.com/AlejandroHerr/cookbook/internal/common"
	pgdb "github.com/AlejandroHerr/cookbook/internal/common/pg"
	"github.com/AlejandroHerr/cookbook/internal/sql"
	"github.com/AlejandroHerr/cookbook/internal/suggestions"
)

var _ suggestions.Repo = (*SuggestionsRepo)(nil)

type SuggestionsRepo struct {
	dbtx    pgdb.DBTX
	queries *sql.Queries
}

func NewSuggestionsRepo(dbtx pgdb.DBTX) *SuggestionsRepo {
	return &SuggestionsRepo{
		dbtx:    dbtx,
		queries: sql.New(),
	}
}

func (r SuggestionsRepo) FindSuggestedTags(ctx context.Context, search string, threshold float64) ([]suggestions.Option, error) {
	var options []suggestions.Option

	if search == "" {
		tags, err := r.queries.FindAllTags(ctx, r.dbtx)
		if err != nil {
			return nil, fmt.Errorf("query FindAllTags: %w", &common.UnexpectedError{Err: err})
		}

		for _, tag := range tags {
			options = append(options, suggestions.NewSimpleOption(tag))
		}
	} else {
		tags, err := r.queries.FindMatchingTags(ctx, r.dbtx, sql.FindMatchingTagsParams{
			Search:    search,
			Threshold: threshold,
		})
		if err != nil {
			return nil, fmt.Errorf("query FindMatchingTags: %w", &common.UnexpectedError{Err: err})
		}

		for _, tag := range tags {
			options = append(options, suggestions.NewSimpleOption(tag.Value))
		}
	}

	return options, nil
}

func (r SuggestionsRepo) FindSuggestedIngredients(ctx context.Context, search string, threshold float64) ([]suggestions.Option, error) {
	var options []suggestions.Option

	if search == "" {
		ingredient, err := r.queries.FindAllIngredients(ctx, r.dbtx)
		if err != nil {
			return nil, fmt.Errorf("query FindAllIngredients: %w", &common.UnexpectedError{Err: err})
		}

		for _, ingredient := range ingredient {
			options = append(options, suggestions.NewSimpleOption(ingredient))
		}
	} else {
		ingredients, err := r.queries.FindMatchingIngredients(ctx, r.dbtx, sql.FindMatchingIngredientsParams{
			Search:    search,
			Threshold: threshold,
		})
		if err != nil {
			return nil, fmt.Errorf("query FindMatchingIngredients: %w", &common.UnexpectedError{Err: err})
		}

		for _, ingredient := range ingredients {
			options = append(options, suggestions.NewSimpleOption(ingredient.Value))
		}
	}

	return options, nil
}
