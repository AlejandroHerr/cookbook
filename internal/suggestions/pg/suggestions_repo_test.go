//nolint:paralleltest
package pg_test

import (
	"testing"

	pgtestutil "github.com/AlejandroHerr/cookbook/internal/common/pg/testutil"
	"github.com/AlejandroHerr/cookbook/internal/suggestions"
	"github.com/AlejandroHerr/cookbook/internal/suggestions/pg"
	"github.com/AlejandroHerr/cookbook/internal/suggestions/pg/fixtures"
	"github.com/stretchr/testify/require"
)

func TestPgIngredients(t *testing.T) {
	db := pgtestutil.MustConnect(t)
	fixtures.MustMakeFixtures(t, db)
	repo := pg.NewSuggestionsRepo(db)

	t.Run("FindSuggestedIngredients", func(t *testing.T) {
		testCases := []struct {
			name     string
			search   string
			minScore float64
			want     []string
		}{
			{
				name:     "Empty search",
				search:   "",
				minScore: 1.0,
				want:     []string{"Tomato", "Tomato Sauce", "Cherry Tomatoes", "Potato", "Sweet Potato", "Onion"},
			},
			{
				name:     "Exact match",
				search:   "Tomato",
				minScore: 0.6,
				want:     []string{"Tomato"},
			},
			{
				name:     "Partial match",
				search:   "tom",
				minScore: 0.1,
				want:     []string{"Tomato", "Tomato Sauce", "Cherry Tomatoes"},
			},
			{
				name:     "Typo tolerance",
				search:   "potatoe",
				minScore: 0.4,
				want:     []string{"Potato", "Sweet Potato"},
			},
			{
				name:     "No results",
				search:   "nonexistent",
				minScore: 0.1,
				want:     []string{},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := repo.FindSuggestedIngredients(t.Context(), tc.search, tc.minScore)
				require.NoError(t, err, "should not fail")

				require.Len(t, got, len(tc.want), "should return the expected number of results")

				expected := make([]suggestions.Option, len(tc.want))
				for i, name := range tc.want {
					expected[i] = suggestions.NewSimpleOption(name)
				}

				require.ElementsMatch(t, expected, got, "should return the expected results")
			})
		}
	})
	t.Run("FindSuggestedTags", func(t *testing.T) {
		testCases := []struct {
			name     string
			search   string
			minScore float64
			want     []string
		}{
			{
				name:     "Empty search",
				search:   "",
				minScore: 0.0,
				want:     []string{"soup", "vegetarian", "easy", "stew", "meat", "dinner", "salad", "quick", "dessert", "baking", "chocolate", "vegan"},
			},
			{
				name:     "Exact match",
				search:   "meat",
				minScore: 0.6,
				want:     []string{"meat"},
			},
			{
				name:     "Partial match",
				search:   "veg",
				minScore: 0.1,
				want:     []string{"vegetarian", "vegan"},
			},
			{
				name:     "Typo tolerance",
				search:   "vegatarian",
				minScore: 0.5,
				want:     []string{"vegetarian"},
			},
			{
				name:     "No results",
				search:   "nonexistent",
				minScore: 0.1,
				want:     []string{},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := repo.FindSuggestedTags(t.Context(), tc.search, tc.minScore)
				require.NoError(t, err, "should not fail")

				require.Len(t, got, len(tc.want), "should return the expected number of results")

				expected := make([]suggestions.Option, len(tc.want))
				for i, tag := range tc.want {
					expected[i] = suggestions.NewSimpleOption(tag)
				}

				require.ElementsMatch(t, expected, got, "should return the expected results")
			})
		}
	})
}
