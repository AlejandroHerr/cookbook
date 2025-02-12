package completions

import (
	"context"

	"github.com/stretchr/testify/mock"
)

var (
	_ Cache     = (*MockCache)(nil)
	_ Scrapper  = (*MockScrapper)(nil)
	_ AIService = (*MockAIService)(nil)
)

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(key string) ([]byte, error) {
	args := m.Called(key)
	return args.Get(0).([]byte), args.Error(1) //nolint:wrapcheck,errcheck
}

func (m *MockCache) Set(key string, entry []byte) error {
	args := m.Called(key, entry)
	return args.Error(0) //nolint:wrapcheck
}

type MockScrapper struct {
	mock.Mock
}

func (m *MockScrapper) Scrap(ctx context.Context, url string) (string, error) {
	args := m.Called(ctx, url)
	return args.String(0), args.Error(1)
}

type MockAIService struct {
	mock.Mock
}

func (m *MockAIService) CompleteRecipe(ctx context.Context, content string) (*Recipe, error) {
	args := m.Called(ctx, content)
	return args.Get(0).(*Recipe), args.Error(1) //nolint:wrapcheck,errcheck
}
