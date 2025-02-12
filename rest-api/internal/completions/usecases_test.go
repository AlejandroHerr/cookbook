package completions_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/AlejandroHerr/cookbook/internal/common/logging"
	"github.com/AlejandroHerr/cookbook/internal/completions"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUsecases(t *testing.T) {
	t.Run("CompleteRecipe", func(t *testing.T) {
		t.Parallel()
		t.Run("returns the recipe from the cache", func(t *testing.T) {
			cache := new(completions.MockCache)
			scrapper := new(completions.MockScrapper)
			aiService := new(completions.MockAIService)

			usecases := completions.MakeUseCases(cache, scrapper, aiService, logging.NewVoidLogger())

			url := "http://example.com/recipe-0"

			expected := &completions.Recipe{ //nolint:exhaustruct
				Title: "Recipe 0",
			}
			cached, _ := json.Marshal(expected)
			cache.On("Get", url).Return(cached, nil)

			got, err := usecases.CompleteRecipe(context.Background(), url)
			require.NoError(t, err, "should not fail")
			require.Equal(t, expected, got, "should return the recipe")

			cache.AssertCalled(t, "Get", url)

			cache.AssertExpectations(t)
		})
		t.Run("scraps a recipe URL and use ai completions", func(t *testing.T) {
			cache := new(completions.MockCache)
			scrapper := new(completions.MockScrapper)
			aiService := new(completions.MockAIService)

			usecases := completions.MakeUseCases(cache, scrapper, aiService, logging.NewVoidLogger())

			url := "http://example.com/recipe-1"

			cache.On("Get", url).Return([]uint8{}, errors.New("not found"))
			cache.On("Set", url, mock.Anything).Return(nil)

			scrapedURL := gofakeit.Sentence(10)

			scrapper.On("Scrap", context.Background(), url).Return(scrapedURL, nil)

			expected := &completions.Recipe{ //nolint:exhaustruct
				Title: "Recipe 1",
			}

			aiService.On("CompleteRecipe", context.Background(), scrapedURL).Return(expected, nil)

			got, err := usecases.CompleteRecipe(context.Background(), url)

			require.NoError(t, err, "should not fail")
			require.Equal(t, expected, got, "should return the recipe")

			cache.AssertCalled(t, "Get", url)

			expectedBytes, _ := json.Marshal(expected)
			cache.AssertCalled(t, "Set", url, expectedBytes)

			scrapper.AssertCalled(t, "Scrap", context.Background(), url)

			aiService.AssertCalled(t, "CompleteRecipe", context.Background(), scrapedURL)

			cache.AssertExpectations(t)
			scrapper.AssertExpectations(t)
			aiService.AssertExpectations(t)
		})
	})
}
