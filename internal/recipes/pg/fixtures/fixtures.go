package fixtures

import (
	"context"
	"testing"

	"github.com/AlejandroHerr/cookbook/internal/common/utils"
	"github.com/AlejandroHerr/cookbook/internal/recipes"
	"github.com/AlejandroHerr/cookbook/internal/sql"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func MustMakeFixtures(t *testing.T, count int, db *pgxpool.Pool) []recipes.Recipe {
	t.Helper()

	ctx := t.Context()

	tx, err := db.Begin(ctx)
	if err != nil {
		panic(err)
	}

	t.Cleanup(func() {
		_ = tx.Rollback(context.Background()) //nolint:usetesting
	})

	q := sql.New()
	list := make([]recipes.Recipe, count)

	for i := range count {
		recipe := new(recipes.Recipe)

		err = gofakeit.Struct(recipe)
		if err != nil {
			panic(err)
		}

		prepTime, err := utils.UintPtrToInt32PtrSafe(recipe.PrepTime)
		if err != nil {
			panic(err)
		}

		servings, err := utils.UintToInt32Safe(recipe.Servings)
		if err != nil {
			panic(err)
		}

		r, err := q.CreateRecipe(ctx, tx, sql.CreateRecipeParams{
			ID:          recipe.ID.String(),
			Title:       recipe.Title,
			Headline:    recipe.Headline,
			Description: recipe.Description,
			Steps:       recipe.Steps,
			PrepTime:    prepTime,
			Servings:    &servings,
			Url:         recipe.URL,
			Tags:        recipe.Tags,
			Slug:        recipe.Slug,
		})
		if err != nil {
			panic(err)
		}

		ingredients := make([]recipes.RecipeIngredient, len(recipe.Ingredients))

		for i, ingredient := range recipe.Ingredients {
			ci, err := q.CreateIngredient(ctx, tx, sql.CreateIngredientParams{
				Name: ingredient.Name,
				Kind: ingredient.Kind,
			})
			if err != nil {
				panic(err)
			}

			_, err = q.CreateRecipeIngredient(ctx, tx, sql.CreateRecipeIngredientParams{
				RecipeID:     r.ID,
				IngredientID: ci.ID,
				Unit:         ingredient.Unit.String(),
				Quantity:     ingredient.Quantity,
			})
			if err != nil {
				panic(err)
			}

			ingredients[i] = recipes.RecipeIngredient{
				ID:       uuid.MustParse(ci.ID),
				Name:     ingredient.Name,
				Kind:     ingredient.Kind,
				Quantity: ingredient.Quantity,
				Unit:     ingredient.Unit,
			}
		}

		createdRecipe, err := r.ToModel()
		if err != nil {
			panic(err)
		}

		createdRecipe.Ingredients = ingredients

		list[i] = createdRecipe
	}

	err = tx.Commit(t.Context())
	if err != nil {
		panic(err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(context.Background(), "TRUNCATE TABLE ingredients CASCADE") //nolint:usetesting
		if err != nil {
			panic(err)
		}

		_, err = db.Exec(context.Background(), "TRUNCATE TABLE recipes CASCADE") //nolint:usetesting
		if err != nil {
			panic(err)
		}
	})

	return list
}
