package slugs

import (
	"context"
	"fmt"

	"github.com/gosimple/slug"
)

type Repo interface {
	GetSlugs(ctx context.Context, entity string, name string) ([]string, error)
}

type Service struct {
	repo Repo
}

func NewService(repo Repo) *Service {
	return &Service{repo: repo}
}

func (s Service) GetUniqueSlug(ctx context.Context, entity string, name string) (string, error) {
	slug := slug.Make(name)

	slugs, err := s.repo.GetSlugs(ctx, entity, slug)
	if err != nil {
		return "", fmt.Errorf("querying existing slugs: %w", err)
	}

	if len(slugs) == 0 {
		return slug, nil
	}

	existingSlugs := make(map[string]bool)

	for _, s := range slugs {
		existingSlugs[s] = true
	}

	if !existingSlugs[slug] {
		return slug, nil
	}

	for i := 1; i < 1000; i++ {
		newSlug := fmt.Sprintf("%s-%d", slug, i)
		if !existingSlugs[newSlug] {
			return newSlug, nil
		}
	}

	return "", fmt.Errorf("could not find a unique slug for %s", name)
}
