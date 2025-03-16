//nolint:exhaustruct
package fixtures

import (
	"context"
	"testing"

	"github.com/AlejandroHerr/cookbook/internal/sql"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5/pgxpool"
)

func MustMakeFixtures(t *testing.T, db *pgxpool.Pool) {
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

	ingredientNames := []string{
		"Tomato",
		"Tomato Sauce",
		"Cherry Tomatoes",
		"Potato",
		"Sweet Potato",
		"Onion",
	}

	for _, name := range ingredientNames {
		_, err = q.CreateIngredient(ctx, tx, sql.CreateIngredientParams{
			Name: name,
		})
		if err != nil {
			panic(err)
		}
	}

	recipes := []struct {
		title string
		tags  []string
	}{
		{title: "Tomato Soup", tags: []string{"soup", "vegetarian", "easy"}},
		{title: "Beef Stew", tags: []string{"stew", "meat", "dinner"}},
		{title: "Garden Salad", tags: []string{"salad", "vegetarian", "quick"}},
		{title: "Chocolate Cake", tags: []string{"dessert", "baking", "chocolate"}},
		{title: "Vegetable Stir Fry", tags: []string{"vegan", "quick", "dinner"}},
	}

	for _, recipe := range recipes {
		_, err = q.CreateRecipe(ctx, tx, sql.CreateRecipeParams{
			ID:    uuid.New().String(),
			Title: recipe.title,
			Tags:  recipe.tags,
			Slug:  slug.Make(recipe.title),
		})
		if err != nil {
			panic(err)
		}
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
}
