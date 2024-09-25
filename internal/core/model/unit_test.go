package model_test

import (
	"testing"

	"github.com/AlejandroHerr/cook-book-go/internal/core/model"
	"github.com/stretchr/testify/assert"
)

func TestUnit(t *testing.T) {
	t.Run("NewIngredient", func(t *testing.T) {
		t.Parallel()

		fixtures := model.MustMakeFixtures(10)

		assert.Len(t, fixtures.Ingredients, 10, "should return 10 ingredients")
	})
}
