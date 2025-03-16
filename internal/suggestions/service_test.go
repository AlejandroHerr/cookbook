package suggestions_test

import (
	"errors"
	"testing"

	"github.com/AlejandroHerr/cookbook/internal/common"
	"github.com/AlejandroHerr/cookbook/internal/suggestions"
	"github.com/AlejandroHerr/cookbook/internal/suggestions/mocks"
	"github.com/stretchr/testify/require"
)

type testService struct {
	mockRepo *mocks.SuggestionsRepo
	service  *suggestions.Service
}

func TestSuggestionsService(t *testing.T) {
	t.Parallel()
	t.Run("GetSuggestions", func(t *testing.T) {
		t.Parallel()

		t.Run("returns the suggestions for tags", func(t *testing.T) {
			service := newTestService()

			search := "some tags"

			expected := []suggestions.Option{}
			service.mockRepo.On("FindSuggestedTags", t.Context(), search, suggestions.THRESHOLD).Return(expected, nil)

			got, err := service.service.GetSuggestions(t.Context(), "tags", search)

			require.NoError(t, err, "should not return an error")
			require.Equal(t, expected, got, "should return the expected suggestions")

			service.mockRepo.AssertExpectations(t)
		})

		t.Run("returns an error when repo returns an error for tags", func(t *testing.T) {
			service := newTestService()

			search := "some tag"

			service.mockRepo.On("FindSuggestedTags", t.Context(), search, suggestions.THRESHOLD).Return([]suggestions.Option{}, errors.New("some error"))

			_, err := service.service.GetSuggestions(t.Context(), "tags", search)
			require.Error(t, err, "should return the error")

			service.mockRepo.AssertExpectations(t)
		})
		t.Run("returns the suggestions for ingredients", func(t *testing.T) {
			service := newTestService()

			search := "some ingredient"

			expected := []suggestions.Option{}
			service.mockRepo.On("FindSuggestedIngredients", t.Context(), search, suggestions.THRESHOLD).Return(expected, nil)

			got, err := service.service.GetSuggestions(t.Context(), "ingredients", search)

			require.NoError(t, err, "should not return an error")
			require.Equal(t, expected, got, "should return the expected suggestions")

			service.mockRepo.AssertExpectations(t)
		})

		t.Run("returns an error when the repo returns an error ingredients", func(t *testing.T) {
			service := newTestService()

			search := "some ingredient"

			service.mockRepo.On("FindSuggestedIngredients", t.Context(), search, suggestions.THRESHOLD).Return([]suggestions.Option{}, errors.New("some error"))

			_, err := service.service.GetSuggestions(t.Context(), "ingredients", search)
			require.Error(t, err, "should return the error")

			service.mockRepo.AssertExpectations(t)
		})

		t.Run("returns an error when the entity is not supported", func(t *testing.T) {
			service := newTestService()

			_, err := service.service.GetSuggestions(t.Context(), "unsupported", "search")
			require.Error(t, err, "should return an error")

			var notFoundErr *common.NotFoundError

			require.ErrorAs(t, err, &notFoundErr, "should return a NotFoundError")
		})
	})
}

func newTestService() *testService {
	mockRepo := new(mocks.SuggestionsRepo)
	service := suggestions.NewService(mockRepo, nil)

	return &testService{
		mockRepo: mockRepo,
		service:  service,
	}
}
