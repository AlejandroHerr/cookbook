package completions_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlejandroHerr/cookbook/internal/scrapper"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
)

func TestScrapper(t *testing.T) {
	t.Run("Scrap", func(t *testing.T) {
		t.Run("returns the scrapped URL and sets the cache", func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprint(w, `<html>
          <body>
          <header>Greate Recipe</header>
          <div>Prepare the food</div>
          </body>
          </html>`)
			}))
			defer ts.Close()

			c := cache.New(cache.NoExpiration, cache.NoExpiration)
			s := scrapper.New(c)

			value, err := s.Scrap(context.Background(), ts.URL)

			assert.NoError(t, err, "should not return an error")
			assert.Contains(t, value, "Prepare the food", "should return the same recipe")

			cached, found := c.Get(ts.URL)

			assert.True(t, found, "should set the cache")
			assert.Equal(t, value, cached, "should set the scrapped value as the cache")
		})
		t.Run("return the value from the cache if exists", func(t *testing.T) {
			t.Parallel()

			url, cached := gofakeit.URL(), gofakeit.Paragraph(5, 5, 5, "")

			c := cache.New(cache.NoExpiration, cache.NoExpiration)
			s := scrapper.New(c)

			c.Set(url, cached, cache.DefaultExpiration)

			value, err := s.Scrap(context.Background(), url)

			assert.NoError(t, err, "should not return an error")
			assert.Equal(t, cached, value, "should return the cached value")
		})
		t.Run("returns an HTTPClientError when the http request fails", func(t *testing.T) {
			t.Parallel()

			error := gofakeit.SentenceSimple()
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, error, http.StatusUnauthorized)
			}))
			defer ts.Close()

			c := cache.New(cache.NoExpiration, cache.NoExpiration)
			s := scrapper.New(c)

			_, err := s.Scrap(context.Background(), ts.URL)

			var httpClientError *scrapper.HTTPClientError

			assert.ErrorAs(t, err, &httpClientError, "should return an HTTPClientError")
			assert.Equal(t, http.StatusUnauthorized, httpClientError.StatusCode, "should contain the HttpStausCode")
			assert.Equal(t, error+"\n", httpClientError.Response, "should contain the server response")
		})
	})
}
