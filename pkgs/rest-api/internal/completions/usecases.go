package completions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AlejandroHerr/cookbook/internal/common/logging"
)

type Cache interface {
	Get(key string) ([]byte, error)
	Set(key string, entry []byte) error
}

type Scrapper interface {
	Scrap(ctx context.Context, url string) (string, error)
}

type AIService interface {
	CompleteRecipe(ctx context.Context, content string) (*Recipe, error)
}

type UseCases struct {
	cache     Cache
	scrapper  Scrapper
	aiService AIService
	logger    logging.Logger
}

func MakeUseCases(
	cache Cache,
	scrapper Scrapper,
	recipeAnalyser AIService,
	logger logging.Logger,
) *UseCases {
	return &UseCases{
		cache:     cache,
		scrapper:  scrapper,
		aiService: recipeAnalyser,
		logger:    logger,
	}
}

func (u UseCases) CompleteRecipe(ctx context.Context, url string) (*Recipe, error) {
	if cached, err := u.cache.Get(url); err == nil {
		var result Recipe

		err = json.Unmarshal(cached, &result)
		if err == nil {
			return &result, nil
		}

		u.logger.Warnw("error unmarshalling value from cache", "url", url, "error", err)
	} else {
		u.logger.Warnw("error reading value from cache", "url", url, "error", err)
	}

	content, err := u.scrapper.Scrap(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("error scrapping url %s: %w", url, err)
	}

	completion, err := u.aiService.CompleteRecipe(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("error getting completions for url %s: %w", url, err)
	}

	if cached, err := json.Marshal(completion); err == nil {
		if err = u.cache.Set(url, cached); err != nil {
			u.logger.Errorw("error saving in cache", "url", url, "error", err)
		}
	} else {
		u.logger.Errorw("error marshaling for cache", "url", url, "error", err)
	}

	return completion, nil
}
