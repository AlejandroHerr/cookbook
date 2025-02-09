package ingredients_test

import (
	"testing"

	"github.com/AlejandroHerr/cookbook/internal/core"
	"github.com/AlejandroHerr/cookbook/internal/ingredients"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIngredientsRepository(t *testing.T) {
	t.Run("Find", func(t *testing.T) {
		t.Run("returns the ingredient by ID", func(t *testing.T) {
			t.Parallel()

			r := ingredients.NewPgRepository(testDBPool)

			ingredient := fixtures.Ingredients[0]

			foundIngredient, err := r.Find(ingredient.ID())

			assert.NoError(t, err, "should not return an error")
			assert.Equal(t, ingredient, *foundIngredient, "should return the same ingredient")
		})
		t.Run("when the ingredient does not exists, then returns a NotFoundError", func(t *testing.T) {
			t.Parallel()

			r := ingredients.NewPgRepository(testDBPool)

			_, err := r.Find(uuid.New())

			var notFoundError *core.NotFoundError

			assert.ErrorAs(t, err, &notFoundError, "should return a core.NotFoundError")
		})
	})
	t.Run("FindByName", func(t *testing.T) {
		t.Run("returns the ingredient by Name", func(t *testing.T) {
			t.Parallel()

			r := ingredients.NewPgRepository(testDBPool)

			ingredient := fixtures.Ingredients[1]

			foundIngredient, err := r.FindByName(ingredient.Name())

			assert.NoError(t, err, "should not return an error")
			assert.Equal(t, ingredient, *foundIngredient, "should return the same ingredient")
		})
		t.Run("when the ingredient does not exists, then returns a NotFoundError", func(t *testing.T) {
			t.Parallel()

			r := ingredients.NewPgRepository(testDBPool)

			_, err := r.FindByName("Not an ingredient")

			var notFoundError *core.NotFoundError

			assert.ErrorAs(t, err, &notFoundError, "should return a core.NotFoundError")
		})
	})
	t.Run("Save", func(t *testing.T) {
		t.Run("creates an ingredient in the database", func(t *testing.T) {
			t.Parallel()

			r := ingredients.NewPgRepository(testDBPool)

			ingredient := ingredients.FakeIngredient()

			err := r.Save(&ingredient)

			assert.NoError(t, err, "should not return an error")

			foundIngredient, _ := r.Find(ingredient.ID())

			assert.Equal(t, ingredient, *foundIngredient, "should return the same ingredient")
		})
		t.Run("returns a DuplicateKeyError when some key is duplicated", func(t *testing.T) {
			t.Parallel()

			r := ingredients.NewPgRepository(testDBPool)

			ingredient := fixtures.Ingredients[2]

			err := r.Save(&ingredient)

			var duplicateKeyError *core.DuplicateKeyError

			assert.ErrorAs(t, err, &duplicateKeyError, "should return a core.DuplicateKeyError")
		})
	})
}
