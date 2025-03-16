package completions_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/AlejandroHerr/cookbook/internal/common/logger"
	"github.com/AlejandroHerr/cookbook/internal/completions"
	"github.com/AlejandroHerr/cookbook/internal/completions/mocks"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	t.Parallel()
	t.Run("CompleteRecipe", func(t *testing.T) {
		t.Parallel()
		t.Run("returns the recipe from the cache", func(t *testing.T) {
			cache := new(mocks.Cache)
			scrapper := new(mocks.Scrapper)
			aiService := new(mocks.AIService)

			service := completions.NewService(cache, scrapper, aiService, logger.NewTestLogger())

			url := "http://example.com/recipe-0"

			expected := &completions.Recipe{ //nolint:exhaustruct
				Title: "Recipe 0",
			}
			cached, _ := json.Marshal(expected)
			cache.On("Get", url).Return(cached, nil)

			got, err := service.CompleteRecipe(t.Context(), url)
			require.NoError(t, err, "should not fail")
			require.Equal(t, expected, got, "should return the recipe")

			cache.AssertCalled(t, "Get", url)

			cache.AssertExpectations(t)
		})
		t.Run("scraps a recipe URL and use ai completions", func(t *testing.T) {
			cache := new(mocks.Cache)
			scrapper := new(mocks.Scrapper)
			aiService := new(mocks.AIService)

			service := completions.NewService(cache, scrapper, aiService, logger.NewTestLogger())

			url := "http://example.com/recipe-1"

			cache.On("Get", url).Return([]uint8{}, errors.New("not found"))
			cache.On("Set", url, mock.Anything).Return(nil)

			scrapedURL := gofakeit.Sentence(10)

			scrapper.On("Scrap", t.Context(), url).Return(scrapedURL, nil)

			expected := &completions.Recipe{ //nolint:exhaustruct
				Title: "Recipe 1",
			}

			aiService.On("CompleteRecipe", t.Context(), scrapedURL).Return(expected, nil)

			got, err := service.CompleteRecipe(t.Context(), url)

			require.NoError(t, err, "should not fail")
			require.Equal(t, expected, got, "should return the recipe")

			cache.AssertCalled(t, "Get", url)

			expectedBytes, _ := json.Marshal(expected)
			cache.AssertCalled(t, "Set", url, expectedBytes)

			scrapper.AssertCalled(t, "Scrap", t.Context(), url)

			aiService.AssertCalled(t, "CompleteRecipe", t.Context(), scrapedURL)

			cache.AssertExpectations(t)
			scrapper.AssertExpectations(t)
			aiService.AssertExpectations(t)
		})
	})
}
