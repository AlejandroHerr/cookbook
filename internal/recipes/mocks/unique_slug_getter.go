//nolint:errcheck,wrapcheck
package mocks

import (
	"context"

	"github.com/AlejandroHerr/cookbook/internal/recipes"
	"github.com/stretchr/testify/mock"
)

var _ recipes.UniqueSlugGetter = (*UniqueSlugGetter)(nil)

type UniqueSlugGetter struct {
	mock.Mock
}

func (m *UniqueSlugGetter) GetUniqueSlug(ctx context.Context, entity string, title string) (string, error) {
	args := m.Called(ctx, entity, title)
	return args.Get(0).(string), args.Error(1)
}
