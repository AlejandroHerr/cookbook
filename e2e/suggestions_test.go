package e2e_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlejandroHerr/cookbook/internal/common/logger"
	pgtestutil "github.com/AlejandroHerr/cookbook/internal/common/pg/testutil"
	"github.com/AlejandroHerr/cookbook/internal/suggestions"
	"github.com/AlejandroHerr/cookbook/internal/suggestions/pg"
	"github.com/AlejandroHerr/cookbook/internal/suggestions/pg/fixtures"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type suggestionTestSuite struct {
	db     *pgxpool.Pool
	server *httptest.Server
	repo   *pg.SuggestionsRepo
}

func TestSuggestions_E2E(t *testing.T) {
	t.Parallel()

	suite := setupSuggestionsTestSuite(t)

	testCases := []struct {
		entity   string
		name     string
		search   string
		minScore float64
		want     []string
	}{
		{
			entity:   "ingredients",
			name:     "Empty search",
			search:   "",
			minScore: 1.0,
			want:     []string{"Tomato", "Tomato Sauce", "Cherry Tomatoes", "Potato", "Sweet Potato", "Onion"},
		},
		{
			entity:   "ingredients",
			name:     "Exact match",
			search:   "Tomato",
			minScore: 0.6,
			want:     []string{"Tomato"},
		},
		{
			entity:   "ingredients",
			name:     "Partial match",
			search:   "tom",
			minScore: 0.1,
			want:     []string{"Tomato", "Tomato Sauce", "Cherry Tomatoes"},
		},
		{
			entity:   "ingredients",
			name:     "Typo tolerance",
			search:   "potatoe",
			minScore: 0.4,
			want:     []string{"Potato", "Sweet Potato"},
		},
		{
			entity:   "ingredients",
			name:     "No results",
			search:   "nonexistent",
			minScore: 0.1,
			want:     []string{},
		},
		{
			entity:   "tags",
			name:     "Empty search",
			search:   "",
			minScore: 0.0,
			want:     []string{"soup", "vegetarian", "easy", "stew", "meat", "dinner", "salad", "quick", "dessert", "baking", "chocolate", "vegan"},
		},
		{
			entity:   "tags",
			name:     "Exact match",
			search:   "meat",
			minScore: 0.6,
			want:     []string{"meat"},
		},
		{
			entity:   "tags",
			name:     "Partial match",
			search:   "veg",
			minScore: 0.1,
			want:     []string{"vegetarian", "vegan"},
		},
		{
			entity:   "tags",
			name:     "Typo tolerance",
			search:   "vegatarian",
			minScore: 0.5,
			want:     []string{"vegetarian"},
		},
		{
			entity:   "tags",
			name:     "No results",
			search:   "nonexistent",
			minScore: 0.1,
			want:     []string{},
		},
	}

	for _, tc := range testCases {
		t.Run("GET /"+tc.entity+" "+tc.name, func(t *testing.T) {
			t.Parallel()

			resp, err := http.Get(suite.server.URL + "/" + tc.entity + "?search=" + tc.search)
			require.NoError(t, err, "failed to make GET request")

			defer resp.Body.Close()

			require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status code")

			var got suggestions.GetSuggestionsReponse
			err = json.NewDecoder(resp.Body).Decode(&got)
			require.NoError(t, err, "response should be a list of suggestions")
		})
	}

	t.Run("GET /non-existent", func(t *testing.T) {
		t.Parallel()

		resp, err := http.Get(suite.server.URL + "/non-existent" + "?search=anything")
		require.NoError(t, err, "failed to make GET request")

		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode, "unexpected status code")
	})
}

func setupSuggestionsTestSuite(t *testing.T) *suggestionTestSuite {
	t.Helper()

	db := pgtestutil.MustConnect(t)
	fixtures.MustMakeFixtures(t, db)

	logger := logger.NewTestLogger()

	repo := pg.NewSuggestionsRepo(db)

	service := suggestions.NewService(repo, logger)

	router := suggestions.NewRouter(service, logger)

	server := httptest.NewServer(router)

	t.Cleanup(server.Close)

	return &suggestionTestSuite{
		db:     db,
		server: server,
		repo:   repo,
	}
}
