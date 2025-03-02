package completions_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlejandroHerr/cookbook/internal/completions"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestHTTPScrapper(t *testing.T) {
	recipeBody := gofakeit.Sentence(10)
	errorBody := gofakeit.Sentence(10)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.Error(w, errorBody, http.StatusNotFound)
			return
		}

		fmt.Fprint(w, `<html>
      <body>
      <header>Greate Recipe</header>
      <div>`+recipeBody+`</div>
      </body>
      </html>`)
	}))

	t.Cleanup(func() {
		ts.Close()
	})

	s := completions.MakeHTTPScrapper()

	t.Run("Scrap", func(t *testing.T) {
		t.Parallel()
		t.Run("returns the scrapped URL", func(t *testing.T) {
			value, err := s.Scrap(context.Background(), ts.URL)

			require.NoError(t, err, "should not return an error")
			require.Contains(t, value, recipeBody, "should return the same recipe")
		})
		t.Run("returns an HTTPClientError when the http request fails", func(t *testing.T) {
			_, err := s.Scrap(context.Background(), ts.URL+"/not-found")

			var httpClientError *completions.HTTPClientError

			require.ErrorAs(t, err, &httpClientError, "should return an HTTPClientError")
			require.Equal(t, http.StatusNotFound, httpClientError.StatusCode, "should contain the HttpStausCode")
			require.Equal(t, errorBody+"\n", httpClientError.Response, "should contain the server response")
		})
	})
}
