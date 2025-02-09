package ingredients_test

import (
	"encoding/json"
	"testing"

	"github.com/AlejandroHerr/cookbook/internal/ingredients"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIngredients(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		t.Run("creates an ingredient with optional fields", func(t *testing.T) {
			t.Parallel()

			id, name, kind := uuid.New(), "tomato", "vegetable"
			ingredients := ingredients.New(id, name, &kind)

			assert.Equal(t, id, ingredients.ID(), "should return the same id")
			assert.Equal(t, name, ingredients.Name(), "should return the same name")
			assert.Equal(t, kind, *ingredients.Kind(), "should return the same kind")
		})
		t.Run("creates an ingredient without optional fields", func(t *testing.T) {
			t.Parallel()

			id, name := uuid.New(), "letuce"
			ingredients := ingredients.New(id, name, nil)

			assert.Equal(t, id, ingredients.ID(), "should return the same id")
			assert.Equal(t, name, ingredients.Name(), "should return the same name")
			assert.Nil(t, ingredients.Kind(), "should return nil")
		})
	})
	t.Run("Marshals", func(t *testing.T) {
		t.Run("marshals an ingredient with optional fields", func(t *testing.T) {
			t.Parallel()

			id, name, kind := uuid.New(), "tofu", "protein"
			ingredient := ingredients.New(id, name, &kind)

			data, err := json.Marshal(ingredient)
			assert.NoError(t, err, "should not return an error")

			var unmarshaledIngredient ingredients.Ingredient
			err = json.Unmarshal(data, &unmarshaledIngredient)
			assert.NoError(t, err, "should not return an error")
			assert.Equal(t, *ingredient, unmarshaledIngredient, "should return the same ingredient")
			assert.Equal(t, *ingredient.Kind(), *unmarshaledIngredient.Kind(), "should return the same kind")
		})
		t.Run("marshals an ingredient without optional fields", func(t *testing.T) {
			t.Parallel()

			id, name := uuid.New(), "potato"
			ingredient := ingredients.New(id, name, nil)

			data, err := json.Marshal(ingredient)
			assert.NoError(t, err, "should not return an error")

			var unmarshaledIngredient ingredients.Ingredient
			err = json.Unmarshal(data, &unmarshaledIngredient)
			assert.NoError(t, err, "should not return an error")
			assert.Equal(t, *ingredient, unmarshaledIngredient, "should return the same ingredient")
		})
	})
}
