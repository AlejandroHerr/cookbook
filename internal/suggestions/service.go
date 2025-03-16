package suggestions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/AlejandroHerr/cookbook/internal/common"
)

type Repo interface {
	FindSuggestedTags(ctx context.Context, search string, threshold float64) ([]Option, error)
	FindSuggestedIngredients(ctx context.Context, search string, threshold float64) ([]Option, error)
}

const THRESHOLD = 0.1

type Service struct {
	repo   Repo
	logger *slog.Logger
}

func NewService(repo Repo, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s Service) GetSuggestions(ctx context.Context, entity string, search string) ([]Option, error) {
	var options []Option

	var err error

	switch entity {
	case "ingredients":
		options, err = s.repo.FindSuggestedIngredients(ctx, search, THRESHOLD)
	case "tags":
		options, err = s.repo.FindSuggestedTags(ctx, search, THRESHOLD)
	default:
		return nil, &common.NotFoundError{Err: errors.New("no suggestions for " + entity)}
	}

	if err != nil {
		return nil, fmt.Errorf("finind suggestions for "+entity+": %w", err)
	}

	return options, nil
}
