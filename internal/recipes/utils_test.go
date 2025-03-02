package recipes_test

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/AlejandroHerr/cookbook/internal/recipes"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type RecipeEqualityOptions struct {
	IgnoreCreatedAt   bool
	IgnoreUpdatedAt   bool
	IgnoreID          bool
	IgnoreIngredients bool
}

func RequireRecipeEqual(t *testing.T, expected, got recipes.Recipe, opts RecipeEqualityOptions, msgAndArgs ...interface{}) {
	t.Helper()

	if opts.IgnoreID {
		expected.ID = uuid.Nil
		got.ID = expected.ID
	}

	if opts.IgnoreCreatedAt {
		expected.CreatedAt = time.Now()
		got.CreatedAt = expected.CreatedAt
	}

	if opts.IgnoreUpdatedAt {
		expected.UpdatedAt = time.Now()
		got.UpdatedAt = expected.UpdatedAt
	}

	if opts.IgnoreIngredients {
		expected.Ingredients = nil
		got.Ingredients = nil
	} else {
		slices.SortFunc(expected.Ingredients, func(l, r recipes.RecipeIngredient) int {
			return strings.Compare(l.Name, r.Name)
		})
		slices.SortFunc(got.Ingredients, func(l, r recipes.RecipeIngredient) int {
			return strings.Compare(l.Name, r.Name)
		})
	}

	require.Equal(t, expected, got, msgAndArgs...)
}

func RequireRecipesEqual(t *testing.T, expected, got []recipes.Recipe, msgAndArgs ...interface{}) {
	t.Helper()

	slices.SortFunc(expected, func(l, r recipes.Recipe) int {
		return strings.Compare(l.Title, r.Title)
	})
	slices.SortFunc(got, func(l, r recipes.Recipe) int {
		return strings.Compare(l.Title, r.Title)
	})

	require.Equal(t, expected, got, msgAndArgs...)
}
