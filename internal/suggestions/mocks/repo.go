//nolint:errcheck,wrapcheck
package mocks

import (
	"context"

	"github.com/AlejandroHerr/cookbook/internal/suggestions"
	"github.com/stretchr/testify/mock"
)

var _ suggestions.Repo = (*SuggestionsRepo)(nil)

type SuggestionsRepo struct {
	mock.Mock
}

func (m *SuggestionsRepo) FindSuggestedTags(ctx context.Context, search string, threshold float64) ([]suggestions.Option, error) {
	args := m.Called(ctx, search, threshold)

	return args.Get(0).([]suggestions.Option), args.Error(1)
}

func (m *SuggestionsRepo) FindSuggestedIngredients(ctx context.Context, search string, threshold float64) ([]suggestions.Option, error) {
	args := m.Called(ctx, search, threshold)

	return args.Get(0).([]suggestions.Option), args.Error(1)
}
