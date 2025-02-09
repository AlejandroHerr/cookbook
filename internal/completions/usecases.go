package completions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AlejandroHerr/cookbook/internal/common/logging"
	"github.com/allegro/bigcache/v3"
)

type Scrapper interface {
	Scrap(ctx context.Context, url string) (string, error)
}

type Completer interface {
	CompleteRecipe(ctx context.Context, content string) (*Recipe, error)
}

type UseCases struct {
	cache          *bigcache.BigCache
	scrapper       Scrapper
	recipeAnalyser Completer
	logger         logging.Logger
}

func NewUseCases(
	cache *bigcache.BigCache,
	scrapper Scrapper,
	recipeAnalyser Completer,
	logger logging.Logger,
) *UseCases {
	return &UseCases{
		cache:          cache,
		scrapper:       scrapper,
		recipeAnalyser: recipeAnalyser,
		logger:         logger,
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

	completion, err := u.recipeAnalyser.CompleteRecipe(ctx, content)
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
