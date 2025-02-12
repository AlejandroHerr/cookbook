package recipes

import (
	"context"
	"fmt"
	"log"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func MustMakeFixtures(count int) []*Recipe {
	recipes := make([]*Recipe, count)
	for i := 0; i < count; i++ {
		err := gofakeit.Struct(&recipes[i])
		if err != nil {
			panic(err)
		}
	}

	return recipes
}

func InstertFixtures(ctx context.Context, pool *pgxpool.Pool, fixtures []*Recipe) error {
	for _, recipe := range fixtures {
		row := pool.QueryRow(
			ctx,
			`
        INSERT INTO
          recipes (id, title, headline, description, steps, servings, url, tags, slug)
        VALUES
          ($1,$2,$3,$4,$5,$6,$7,$8,$9)
        RETURNING
          created_at, updated_at
      `,
			recipe.ID, recipe.Title, recipe.Headline, recipe.Description, recipe.Steps, recipe.Servings, recipe.URL, recipe.Tags, recipe.Slug(),
		)

		if err := row.Scan(&recipe.CreatedAt, &recipe.UpdatedAt); err != nil {
			return fmt.Errorf("scanning recipe: %v", err)
		}

		recipeIDs := make([]uuid.UUID, 0)
		ingredientIDs := make([]uuid.UUID, 0)
		ingredientNames := make([]string, 0)
		ingredientQuantities := make([]float64, 0)
		ingredientUnits := make([]string, 0)
		ingredientKinds := make([]string, 0)

		for _, ingredient := range recipe.Ingredients {
			recipeIDs = append(recipeIDs, recipe.ID)
			ingredientIDs = append(ingredientIDs, ingredient.ID)
			ingredientNames = append(ingredientNames, ingredient.Name)
			ingredientQuantities = append(ingredientQuantities, ingredient.Quantity)
			ingredientUnits = append(ingredientUnits, ingredient.Unit.String())
			ingredientKinds = append(ingredientKinds, *ingredient.Kind)
		}

		_, err := pool.Exec(
			ctx,
			`
        INSERT INTO
          ingredients (id, name, kind)
        SELECT * FROM UNNEST($1::uuid[], $2::text[], $3::text[])
      `,
			ingredientIDs, ingredientNames, ingredientKinds,
		)
		if err != nil {
			return fmt.Errorf("inserting ingredients: %v", err)
		}

		_, err = pool.Exec(
			ctx,
			`
        INSERT INTO
          recipe_ingredients (recipe_id, ingredient_id, unit, quantity)
        SELECT * FROM UNNEST($1::uuid[], $2::uuid[], $3::text[], $4::float[])
      `,
			recipeIDs, ingredientIDs, ingredientUnits, ingredientQuantities,
		)
		if err != nil {
			return fmt.Errorf("inserting recipe ingredients: %v", err)
		}
	}

	return nil
}

func MustCleanUpFixtures(ctx context.Context, pool *pgxpool.Pool) {
	_, err := pool.Exec(
		ctx,
		"DELETE FROM recipes",
	)
	if err != nil {
		log.Fatalf("cleaning up recipes: %v", err)
	}

	_, err = pool.Exec(
		ctx,
		"DELETE FROM Ingredients",
	)
	if err != nil {
		log.Fatalf("cleaning up ingredients: %v", err)
	}
}
