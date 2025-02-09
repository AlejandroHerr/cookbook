package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AlejandroHerr/cook-book-go/internal/common/infra/db"
	"github.com/AlejandroHerr/cook-book-go/internal/common/logging"
	"github.com/AlejandroHerr/cook-book-go/internal/recipes"
	"github.com/allegro/bigcache/v3"
	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
)

type Config struct {
	DB db.Config
	// OpenAIConfig completions.OpenAIConfig
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	logger, err := logging.CreateLogger()
	if err != nil {
		return fmt.Errorf("error creating logger: %w", err)
	}

	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	dbLogger := db.NewPgxLogger(logger.Named("pgx"))

	dbPool, err := db.Connect(
		context.Background(),
		&config.DB,
		0,
		dbLogger,
	)
	if err != nil {
		return fmt.Errorf("error connecting to db: %w", err)
	}
	defer dbPool.Close()

	cache, err := bigcache.New(context.Background(), bigcache.DefaultConfig(time.Hour))
	if err != nil {
		return fmt.Errorf("error creating cache: %w", err)
	}
	defer cache.Close()

	sessionManager := db.NewPgxTransactionManager(dbPool)

	ingredientsRepo := recipes.NewPgIngredientsRepo(dbPool)
	recipesRepo := recipes.NewPgRecipesRepository(dbPool)

	recipesUseCases := recipes.NewUseCases(sessionManager, recipesRepo, ingredientsRepo, logger)
	recipesRouter := recipes.NewRouter(recipesUseCases)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(middleware.NoCache)
	r.Use(cors.Handler(cors.Options{ //nolint:exhaustruct
		AllowedOrigins: []string{"*"},
	}))
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Mount("/recipes", recipesRouter)

	server := &http.Server{ //nolint: exhaustruct
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 3 * time.Second,
	}

	logger.Infow("Server listening", "address", server.Addr)

	err = server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("error starting server: %w", err)
	}

	return nil
}

func loadConfig() (*Config, error) {
	config := &Config{} //nolint:exhaustruct
	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	return config, nil
}
