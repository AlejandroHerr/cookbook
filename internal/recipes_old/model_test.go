package recipes_test

import (
	"encoding/json"
	"testing"

	"github.com/AlejandroHerr/cookbook/internal/ingredients"
	"github.com/AlejandroHerr/cookbook/internal/recipes"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRecipes(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		t.Run("creates a recipe with optional fields", func(t *testing.T) {
			t.Parallel()

			id, name := uuid.New(), gofakeit.Dinner()
			description, url := gofakeit.LoremIpsumParagraph(1, 5, 10, " "), gofakeit.URL()
			ingredientsList := []ingredients.Ingredient{
				ingredients.FakeIngredient(),
				ingredients.FakeIngredient(),
			}

			recipe := recipes.New(id, name, &description, &url, ingredientsList)

			assert.Equal(t, id, recipe.ID(), "should return the same id")
			assert.Equal(t, name, recipe.Name(), "should return the same name")
			assert.Equal(t, description, *recipe.Description(), "should return the same description")
			assert.Equal(t, url, *recipe.URL(), "should return the same url")
			assert.Equal(t, ingredientsList, recipe.Ingredients(), "should return the same ingredients")
		})
		t.Run("creates an ingredients without optional fields", func(t *testing.T) {
			t.Parallel()

			id, name := uuid.New(), gofakeit.Dinner()
			recipe := recipes.New(id, name, nil, nil, nil)

			assert.Equal(t, id, recipe.ID(), "should return the same id")
			assert.Equal(t, name, recipe.Name(), "should return the same name")
			assert.Nil(t, recipe.Description(), "should return nil")
			assert.Nil(t, recipe.URL(), "should return nil")
			assert.Empty(t, recipe.Ingredients(), "should return an empty list")
		})
	})
	t.Run("Marshals", func(t *testing.T) {
		t.Run("marshals an ingredient with optional fields", func(t *testing.T) {
			t.Parallel()

			id, name := uuid.New(), gofakeit.Dinner()
			description, url := gofakeit.LoremIpsumParagraph(1, 5, 10, " "), gofakeit.URL()
			ingredientsList := []ingredients.Ingredient{
				ingredients.FakeIngredient(),
				ingredients.FakeIngredient(),
			}
			recipe := recipes.New(id, name, &description, &url, ingredientsList)

			data, err := json.Marshal(recipe)
			assert.NoError(t, err, "should not return an error")

			var unmarshaledRecipe recipes.Recipe
			err = json.Unmarshal(data, &unmarshaledRecipe)
			assert.NoError(t, err, "should not return an error")
			assert.Equal(t, *recipe, unmarshaledRecipe, "should return the same ingredient")
			assert.Equal(t, recipe.Ingredients(), unmarshaledRecipe.Ingredients(), "should return the same ingredients")
		})
		t.Run("marshals an ingredient without optional fields", func(t *testing.T) {
			t.Parallel()

			id, name := uuid.New(), gofakeit.Dinner()
			recipe := recipes.New(id, name, nil, nil, nil)

			data, err := json.Marshal(recipe)
			assert.NoError(t, err, "should not return an error")

			var unmarshaledRecipe recipes.Recipe
			err = json.Unmarshal(data, &unmarshaledRecipe)
			assert.NoError(t, err, "should not return an error")
			assert.Equal(t, *recipe, unmarshaledRecipe, "should return the same ingredient")
		})
	})
	t.Run("AddIngredient", func(t *testing.T) {
		t.Run("adds an ingredient to the recipe", func(t *testing.T) {
			t.Parallel()

			recipe := recipes.FakeRecipe()
			ingredient := ingredients.FakeIngredient()

			recipe.AddIngredients(ingredient)

			assert.Contains(t, recipe.Ingredients(), ingredient, "should contain the ingredient")
		})
	})
}
