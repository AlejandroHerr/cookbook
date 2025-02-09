package recipes

import (
	"context"
	"fmt"
	"time"

	"github.com/AlejandroHerr/cookbook/internal/ingredients"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Fixtures struct {
	Ingredients []ingredients.Ingredient
	Recipes     []Recipe
}

func FakeRecipe() Recipe {
	var recipe Recipe

	err := gofakeit.Struct(&recipe)
	if err != nil {
		panic(fmt.Errorf("could not generate the fake recipe: %w", err))
	}

	return recipe
}

func GenerateFixtures(number int) *Fixtures {
	ingredientsFixtures := ingredients.GenerateFixtures(number * 2)
	recipesList := make([]Recipe, number)

	for i := 0; i < number; i++ {
		recipe := FakeRecipe()

		recipesList[i] = recipe
	}

	fixtures := Fixtures{
		Ingredients: ingredientsFixtures.Ingredients,
		Recipes:     recipesList,
	}

	return &fixtures
}

func LoadFixtures(ctx context.Context, pool *pgxpool.Pool, fixtures *Fixtures) {
	ingredients.LoadFixtures(ctx, pool, &ingredients.Fixtures{Ingredients: fixtures.Ingredients})

	handlerCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for _, recipe := range fixtures.Recipes {
		_, err := pool.Exec(
			handlerCtx,
			"INSERT INTO recipes (id, name, description, url) VALUES ($1, $2, $3, $4)",
			recipe.ID(),
			recipe.Name(),
			recipe.Description(),
			recipe.URL(),
		)
		if err != nil {
			panic(fmt.Errorf("could not insert the ingredient fixture: %w", err))
		}

		for _, ingredient := range recipe.Ingredients() {
			_, err := pool.Exec(
				handlerCtx,
				"INSERT INTO recipe_ingredients (recipe_id, ingredient_id) VALUES ($1, $2)",
				recipe.ID(),
				ingredient.ID(),
			)
			if err != nil {
				panic(fmt.Errorf("could not insert the recipe_ingredient fixture: %w", err))
			}
		}
	}
}

func CleanupFixtures(ctx context.Context, pool *pgxpool.Pool) {
	handlerCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := pool.Exec(handlerCtx, "DELETE FROM recipe_ingredients")
	if err != nil {
		panic(fmt.Errorf("could not clean the recipe_ingredients table: %w", err))
	}

	_, err = pool.Exec(handlerCtx, "DELETE FROM recipes")
	if err != nil {
		panic(fmt.Errorf("could not clean the recipes table: %w", err))
	}

	ingredients.CleanupFixtures(ctx, pool)
}
