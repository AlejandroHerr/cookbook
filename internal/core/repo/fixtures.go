package repo

import (
	"context"
	"fmt"
	"log"

	"github.com/AlejandroHerr/cook-book-go/internal/core/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

const insertIngredientQuery = `
    INSERT INTO 
      Ingredients (id,name,kind)
    VALUES
      ($1,$2,$3)
  `

const insertRecipeQuery = `
    INSERT INTO
      recipes (id, title, headline, description, steps, servings, url, tags, slug)
    VALUES 
      ($1,$2,$3,$4,$5,$6,$7,$8,$9)
    RETURNING 
      id, title, headline, description, steps, servings, url, tags, created_at, updated_at
  `

const insertRecipeIngredient = `
    INSERT INTO
      recipe_ingredients (recipe_id, ingredient_id, unit, quantity) 
    VALUES 
      ($1,$2,$3,$4)
  `

func MustInstertFixtures(ctx context.Context, pool *pgxpool.Pool, fixtures *model.Fixtures) {
	for _, ingredient := range fixtures.Ingredients {
		_, err := pool.Exec(
			ctx,
			insertIngredientQuery,
			ingredient.ID,
			ingredient.Name,
			ingredient.Kind,
		)
		if err != nil {
			panic(fmt.Errorf("inserting ingredient fixture: %w", err))
		}
	}

	for _, recipe := range fixtures.Recipes {
		row := pool.QueryRow(
			ctx,
			insertRecipeQuery,
			recipe.ID,
			recipe.Title,
			recipe.Headline,
			recipe.Description,
			recipe.Steps,
			recipe.Servings,
			recipe.URL,
			recipe.Tags,
			recipe.Slug(),
		)

		if err := row.Scan(
			&recipe.ID,
			&recipe.Title,
			&recipe.Headline,
			&recipe.Description,
			&recipe.Steps,
			&recipe.Servings,
			&recipe.URL,
			&recipe.Tags,
			&recipe.CreatedAt,
			&recipe.UpdatedAt,
		); err != nil {
			log.Fatalf("inserting/scanning: %v", err)
		}

		for _, ingredient := range recipe.Ingredients {
			_, err := pool.Exec(ctx, insertRecipeIngredient, recipe.ID, ingredient.ID, ingredient.Unit, ingredient.Quantity)
			if err != nil {
				panic(fmt.Errorf("inserting recipe ingredient fixture: %w", err))
			}
		}
	}
}

func MustCleanup(ctx context.Context, pool *pgxpool.Pool) {
	_, err := pool.Exec(ctx, "DELETE FROM recipe_ingredients")
	if err != nil {
		panic(err)
	}

	_, err = pool.Exec(ctx, "DELETE FROM recipes")
	if err != nil {
		panic(err)
	}

	_, err = pool.Exec(ctx, "DELETE FROM ingredients")
	if err != nil {
		panic(err)
	}
}
