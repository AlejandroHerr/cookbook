package slugs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AlejandroHerr/cookbook/internal/slugs"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/gosimple/slug"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSlugsService(t *testing.T) {
	t.Parallel()
	t.Run("GetSlugs", func(t *testing.T) {
		t.Parallel()
		t.Run("returns the slug if it is not used", func(t *testing.T) {
			repo := new(MockRepo)
			service := slugs.NewService(repo)

			entity := "entity"
			name := gofakeit.Sentence(10)

			slugs := []string{}

			repo.On("GetSlugs", t.Context(), entity, slug.Make(name)).Return(slugs, nil)

			got, err := service.GetUniqueSlug(t.Context(), entity, name)
			require.NoError(t, err)

			require.Equal(t, slug.Make(name), got, "should return the slug")

			repo.AssertExpectations(t)
		})
		t.Run("returns a unique slug if it is used", func(t *testing.T) {
			repo := new(MockRepo)
			service := slugs.NewService(repo)

			entity := "entity"
			name := gofakeit.Sentence(10)

			slugs := []string{slug.Make(name), slug.Make(name) + "-1", slug.Make(name) + "-2", slug.Make(name) + "-vegan"}

			repo.On("GetSlugs", t.Context(), entity, slug.Make(name)).Return(slugs, nil)

			got, err := service.GetUniqueSlug(t.Context(), entity, name)
			require.NoError(t, err)

			require.Equal(t, slug.Make(name)+"-3", got, "should return the slug")

			repo.AssertExpectations(t)
		})
		t.Run("returns an error if the repo fails", func(t *testing.T) {
			repo := new(MockRepo)
			service := slugs.NewService(repo)

			entity := "entity"
			name := gofakeit.Sentence(10)

			repo.On("GetSlugs", t.Context(), entity, slug.Make(name)).Return([]string{}, gofakeit.Error())

			_, got := service.GetUniqueSlug(t.Context(), entity, name)
			require.Error(t, got)

			repo.AssertExpectations(t)
		})

		t.Run("returns an error if it cannot find a unique slug after n=1000", func(t *testing.T) {
			repo := new(MockRepo)
			service := slugs.NewService(repo)

			entity := "entity"
			name := gofakeit.Sentence(10)

			slugs := make([]string, 1000)
			for i := range slugs {
				if i == 0 {
					slugs[i] = slug.Make(name)
				} else {
					slugs[i] = fmt.Sprintf("%s-%d", slug.Make(name), i)
				}

				repo.On("GetSlugs", t.Context(), entity, slug.Make(name)).Return(slugs, nil)
			}

			_, got := service.GetUniqueSlug(t.Context(), entity, name)
			require.Error(t, got)
		})
	})
}

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) GetSlugs(ctx context.Context, entity string, name string) ([]string, error) {
	args := m.Called(ctx, entity, name)
	return args.Get(0).([]string), args.Error(1)
}
