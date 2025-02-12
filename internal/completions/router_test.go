package completions_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlejandroHerr/cookbook/internal/common/logging"
	"github.com/AlejandroHerr/cookbook/internal/completions"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRouter(t *testing.T) {
	cache := new(completions.MockCache)
	scrapper := new(completions.MockScrapper)
	aiService := new(completions.MockAIService)
	useCases := completions.MakeUseCases(cache, scrapper, aiService, logging.NewVoidLogger())
	router := completions.MakeRouter(useCases)

	r := chi.NewRouter()
	r.Mount("/completions", router)

	server := httptest.NewServer(r)

	t.Cleanup(server.Close)

	t.Run("POST /recipe", func(t *testing.T) {
		t.Parallel()
		t.Run("returns a recipe completion", func(t *testing.T) {
			url := "http://example.com/recipe"

			cache.On("Get", url).Return([]uint8{}, errors.New("not found"))
			cache.On("Set", url, mock.Anything).Return(nil)

			scrapedURL := gofakeit.Sentence(10)

			scrapper.On("Scrap", mock.Anything, url).Return(scrapedURL, nil)

			expected := &completions.Recipe{ //nolint:exhaustruct
				Title:       "Recipe",
				Description: gofakeit.Sentence(10),
			}

			aiService.On("CompleteRecipe", mock.Anything, scrapedURL).Return(expected, nil)

			req := completions.CompleteRecipeRequest{
				URL: url,
			}
			jsonBody, err := json.Marshal(req)
			require.NoError(t, err, "error marshaling CompleteRecipeRequest")

			resp, err := http.Post(server.URL+"/completions/recipe", "application/json", bytes.NewBuffer(jsonBody))
			require.NoError(t, err, "request should not fail")
			defer resp.Body.Close()

			var fetched completions.CompleteRecipeResponse
			err = json.NewDecoder(resp.Body).Decode(&fetched)
			require.NoError(t, err, "should be a CompleteRecipeResponse")

			require.Equal(t, *expected, fetched.Recipe, "should return the recipe")

			cache.AssertCalled(t, "Get", url)

			expectedBytes, _ := json.Marshal(expected)
			cache.AssertCalled(t, "Set", url, expectedBytes)

			scrapper.AssertCalled(t, "Scrap", mock.Anything, url)

			aiService.AssertCalled(t, "CompleteRecipe", mock.Anything, scrapedURL)
		})
		t.Run("returns a Bad Request error if the url is invalid", func(t *testing.T) {
			cache := new(completions.MockCache)
			scrapper := new(completions.MockScrapper)
			aiService := new(completions.MockAIService)

			usecases := completions.MakeUseCases(cache, scrapper, aiService, logging.NewVoidLogger())

			url := "http://example.com/recipe-3"

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
