package suggestions

import (
	"context"
	"fmt"
	"sync"

	"github.com/AlejandroHerr/cookbook/internal/recipes"
)

type Repo interface {
	FindMatchingTags(ctx context.Context, search string) ([]Option, error)
	FindAllTags(ctx context.Context) ([]Option, error)
	FindMatchingIngredients(ctx context.Context, search string) ([]Option, error)
	FindAllIngredients(ctx context.Context) ([]Option, error)
}

type UseCases struct {
	repo Repo
}

func MakeUseCases(repo Repo) *UseCases {
	return &UseCases{
		repo: repo,
	}
}

func (u UseCases) GetTagsOptions(ctx context.Context, search string) ([]Option, error) {
	var options []Option

	var err error

	if search == "" {
		options, err = u.repo.FindAllTags(ctx)
		if err != nil {
			return nil, fmt.Errorf("find all tags: %w", err)
		}
	} else {
		options, err = u.repo.FindMatchingTags(ctx, search)
		if err != nil {
			return nil, fmt.Errorf("find matching tags: %w", err)
		}
	}

	return options, nil
}

func (u UseCases) GetIngredientsOptions(ctx context.Context, search string) ([]Option, error) {
	var options []Option

	var err error

	if search == "" {
		options, err = u.repo.FindAllIngredients(ctx)
		if err != nil {
			return nil, fmt.Errorf("find all ingredients: %w", err)
		}
	} else {
		options, err = u.repo.FindMatchingIngredients(ctx, search)
		if err != nil {
			return nil, fmt.Errorf("find matching ingredients: %w", err)
		}
	}

	return options, nil
}

var (
	once        sync.Once
	unitOptions = make([]Option, len(recipes.Units))
)

func (u UseCases) GetUnitsOptions(context.Context) ([]Option, error) {
	once.Do(func() {
		for i, u := range recipes.Units {
			unitOptions[i] = Option{
				Label: recipes.UnitDisplayNames[u],
				Value: u.String(),
			}
		}
	})

	return unitOptions, nil
}
