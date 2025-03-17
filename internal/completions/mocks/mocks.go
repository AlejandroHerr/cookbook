//nolint:wrapcheck,errcheck
package mocks

import (
	"context"

	"github.com/AlejandroHerr/cookbook/internal/completions"
	"github.com/stretchr/testify/mock"
)

var (
	_ completions.Cache          = (*Cache)(nil)
	_ completions.Scrapper       = (*Scrapper)(nil)
	_ completions.AICompletioner = (*AIService)(nil)
)

type Cache struct {
	mock.Mock
}

func (m *Cache) Get(key string) ([]byte, error) {
	args := m.Called(key)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *Cache) Set(key string, entry []byte) error {
	args := m.Called(key, entry)
	return args.Error(0)
}

type Scrapper struct {
	mock.Mock
}

func (m *Scrapper) Scrap(ctx context.Context, url string) (string, error) {
	args := m.Called(ctx, url)
	return args.String(0), args.Error(1)
}

type AIService struct {
	mock.Mock
}

func (m *AIService) CompleteRecipe(ctx context.Context, content string) (*completions.Recipe, error) {
	args := m.Called(ctx, content)
	return args.Get(0).(*completions.Recipe), args.Error(1)
}
