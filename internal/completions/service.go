package completions

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

type Cache interface {
	Get(key string) ([]byte, error)
	Set(key string, entry []byte) error
}

type Scrapper interface {
	Scrap(ctx context.Context, url string) (string, error)
}

type AICompletioner interface {
	CompleteRecipe(ctx context.Context, content string) (*Recipe, error)
}

type Service struct {
	cache     Cache
	scrapper  Scrapper
	aiService AICompletioner
	logger    *slog.Logger
}

func NewService(
	cache Cache,
	scrapper Scrapper,
	recipeAnalyser AICompletioner,
	logger *slog.Logger,
) *Service {
	return &Service{
		cache:     cache,
		scrapper:  scrapper,
		aiService: recipeAnalyser,
		logger:    logger,
	}
}

func (s Service) CompleteRecipe(ctx context.Context, url string) (*Recipe, error) {
	if cached, err := s.cache.Get(url); err == nil {
		var result Recipe

		err = json.Unmarshal(cached, &result)
		if err == nil {
			return &result, nil
		}

		s.logger.WarnContext(
			ctx,
			"error unmarshalling value from cache",
			slog.String("url", url),
			slog.Any("error", err),
		)
	} else {
		s.logger.WarnContext(
			ctx,
			"error reading value from cache",
			slog.String("url", url),
			slog.Any("error", err),
		)
	}

	content, err := s.scrapper.Scrap(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("error scrapping url %s: %w", url, err)
	}

	completion, err := s.aiService.CompleteRecipe(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("error getting completions for url %s: %w", url, err)
	}

	if cached, err := json.Marshal(completion); err == nil {
		if err = s.cache.Set(url, cached); err != nil {
			s.logger.WarnContext(
				ctx,
				"error saving in cache",
				slog.String("url", url),
				slog.Any("error", err),
			)
		}
	} else {
		s.logger.WarnContext(
			ctx,
			"error marshaling for cache",
			slog.String("url", url),
			slog.Any("error", err),
		)
	}

	return completion, nil
}
