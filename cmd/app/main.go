package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AlejandroHerr/cookbook/internal/common/api"
	"github.com/AlejandroHerr/cookbook/internal/common/logger"
	"github.com/AlejandroHerr/cookbook/internal/common/pg"
	"github.com/AlejandroHerr/cookbook/internal/completions"
	"github.com/AlejandroHerr/cookbook/internal/recipes"
	pgrecipes "github.com/AlejandroHerr/cookbook/internal/recipes/pg"
	"github.com/AlejandroHerr/cookbook/internal/suggestions"
	pgsuggestions "github.com/AlejandroHerr/cookbook/internal/suggestions/pg"
	"github.com/allegro/bigcache/v3"
	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
)

type Config struct {
	DB           *pg.Config
	OpenAIConfig *completions.OpenAIConfig
	Environment  string `env:"ENVIRONMENT" envDefault:"development"`
	LogLevel     string `env:"LOG_LEVEL" envDefault:"debug"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	logger := logger.New(logger.Config{
		Level:       config.LogLevel,
		Environment: config.Environment,
		App:         "rest-api",
		Version:     "1.0.0",
	})

	dbLogger := pg.NewPgxLogger(logger)

	dbPool, err := pg.Connect(
		context.Background(),
		config.DB,
		0,
		dbLogger,
	)
	if err != nil {
		return fmt.Errorf("error connecting to db: %w", err)
	}
	defer dbPool.Close()

	// Declare Recipes Router
	sessionManager := pg.NewTransactionManager(dbPool)
	ingredientsRepo := pgrecipes.NewIngredientsRepo(dbPool)
	recipesRepo := pgrecipes.NewRecipesRepo(dbPool)
	recipesService := recipes.NewService(sessionManager, recipesRepo, ingredientsRepo, logger.With("service", "recipes"))
	recipesRouter := recipes.NewRouter(recipesService, logger.With("service", "recipes-router"))

	// Declare Suggestions Router
	suggestionsRepo := pgsuggestions.NewSuggestionsRepo(dbPool)
	suggestionsUseCases := suggestions.NewService(suggestionsRepo, logger.With("service", "suggestions"))
	suggestionsRouter := suggestions.NewRouter(suggestionsUseCases, logger.With("service", "suggestions-router"))

	// Declare Completions Router
	cache, err := bigcache.New(context.Background(), bigcache.DefaultConfig(time.Hour))
	if err != nil {
		return fmt.Errorf("error creating cache: %w", err)
	}
	defer cache.Close()

	scrapper := completions.NewHTTPScrapper()
	aiService := completions.NewOpenAIService(config.OpenAIConfig, logger.With("service", "openai"))
	completionsUseCases := completions.NewService(cache, scrapper, aiService, logger)
	completionsRouter := completions.NewRouter(completionsUseCases, logger)

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(api.RequestIDMiddleware())
	r.Use(api.RequestLoggerMiddleware(logger))
	r.Use(middleware.URLFormat)
	r.Use(middleware.NoCache)
	r.Use(cors.Handler(cors.Options{ //nolint:exhaustruct
		AllowedOrigins: []string{"*"},
	}))
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Mount("/recipes", recipesRouter)
	r.Mount("/suggestions", suggestionsRouter)
	r.Mount("/completions", completionsRouter)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 3 * time.Second,
	}

	logger.InfoContext(context.Background(), "Server listening", "address", server.Addr)

	err = server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("error starting server: %w", err)
	}

	return nil
}

func loadConfig() (*Config, error) {
	config := &Config{
		DB:           &pg.Config{},                //nolint:exhaustruct
		OpenAIConfig: &completions.OpenAIConfig{}, //nolint:exhaustruct
	}
	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	return config, nil
}
