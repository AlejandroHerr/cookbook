package testutil

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/AlejandroHerr/cook-book-go/internal/core/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func RequireRecipeEquals(t *testing.T, expected model.Recipe, got model.Recipe, excludeDates, excludeID bool, msgAndArgs ...interface{}) {
	t.Helper()

	if excludeID {
		expected.ID = uuid.Nil
		got.ID = uuid.Nil
	}

	if excludeDates {
		expected.CreatedAt = time.Time{}
		expected.UpdatedAt = time.Time{}
		got.CreatedAt = time.Time{}
		got.UpdatedAt = time.Time{}
	}

	slices.SortFunc(expected.Ingredients, func(l, r model.RecipeIngredient) int {
		return strings.Compare(l.Name, r.Name)
	})
	slices.SortFunc(got.Ingredients, func(l, r model.RecipeIngredient) int {
		return strings.Compare(l.Name, r.Name)
	})

	require.Equal(t, expected, got, msgAndArgs...)
}
