package ingredients

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Fixtures struct {
	Ingredients []Ingredient
}

func FakeIngredient() Ingredient {
	var ingredient Ingredient

	err := gofakeit.Struct(&ingredient)
	if err != nil {
		panic(fmt.Errorf("could not generate the fake ingredient: %w", err))
	}

	return ingredient
}

func GenerateFixtures(number int) *Fixtures {
	ingredientsList := make([]Ingredient, number)

	for i := 0; i < number; i++ {
		ingredient := FakeIngredient()

		ingredientsList[i] = ingredient
	}

	fixtures := Fixtures{
		Ingredients: ingredientsList,
	}

	return &fixtures
}

func LoadFixtures(ctx context.Context, pool *pgxpool.Pool, fakes *Fixtures) {
	handlerCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for _, ingredient := range fakes.Ingredients {
		_, err := pool.Exec(
			handlerCtx,
			"INSERT INTO ingredients (id, name, kind) VALUES ($1, $2, $3)",
			ingredient.ID(),
			ingredient.Name(),
			ingredient.Kind(),
		)
		if err != nil {
			panic(fmt.Errorf("could not insert the ingredient fixture: %w", err))
		}
	}
}

func CleanupFixtures(ctx context.Context, pool *pgxpool.Pool) {
	handlerCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := pool.Exec(handlerCtx, "DELETE FROM ingredients")
	if err != nil {
		panic(fmt.Errorf("could not clean the ingredients table: %w", err))
	}
}
