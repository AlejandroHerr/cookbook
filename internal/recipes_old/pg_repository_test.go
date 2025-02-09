package recipes_test

import (
	"testing"

	"github.com/AlejandroHerr/cookbook/internal/core"
	"github.com/AlejandroHerr/cookbook/internal/ingredients"
	"github.com/AlejandroHerr/cookbook/internal/recipes"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRecipesRepository(t *testing.T) {
	t.Run("Find", func(t *testing.T) {
		t.Run("returns a recipe", func(t *testing.T) {
			t.Parallel()

			recipe := fixtures.Recipes[0]

			repo := recipes.NewPgRepository(testDBPool)

			foundRecipe, err := repo.Find(recipe.ID())
			assert.NoError(t, err, "should not return an error")

			assert.Equal(t, recipe, *foundRecipe, "should return the same recipe")
		})
		t.Run("return not found if the recipe does not exist", func(t *testing.T) {
			t.Parallel()

			repo := recipes.NewPgRepository(testDBPool)

			foundRecipe, err := repo.Find(uuid.New())
			assert.Nil(t, foundRecipe, "should return nil")

			var notFoundError *core.NotFoundError

			assert.ErrorAs(t, err, &notFoundError, "should return a NotFoundError")
		})
	})
	t.Run("FindByName", func(t *testing.T) {
		t.Run("returns a recipe", func(t *testing.T) {
			t.Parallel()

			recipe := fixtures.Recipes[1]

			repo := recipes.NewPgRepository(testDBPool)

			foundRecipe, err := repo.FindByName(recipe.Name())
			assert.NoError(t, err, "should not return an error")

			assert.Equal(t, recipe, *foundRecipe, "should return the same recipe")
		})
		t.Run("return not found if the recipe does not exist", func(t *testing.T) {
			t.Parallel()

			repo := recipes.NewPgRepository(testDBPool)

			foundRecipe, err := repo.FindByName("Blue Car")
			assert.Nil(t, foundRecipe, "should return nil")

			var notFoundError *core.NotFoundError

			assert.ErrorAs(t, err, &notFoundError, "should return a NotFoundError")
		})
	})
	t.Run("Save", func(t *testing.T) {
		t.Run("saves a recipe", func(t *testing.T) {
			t.Parallel()

			recipe := recipes.FakeRecipe()

			recipe.AddIngredients(fixtures.Ingredients[0])
			recipe.AddIngredients(fixtures.Ingredients[1])

			repo := recipes.NewPgRepository(testDBPool)

			err := repo.Save(&recipe)
			assert.NoError(t, err, "should not return an error")

			foundRecipe, err := repo.Find(recipe.ID())
			assert.NoError(t, err, "should not return an error")

			assert.Equal(t, recipe, *foundRecipe, "should return the same recipe")
		})
		t.Run("saves a recipe without required fields", func(t *testing.T) {
			t.Parallel()

			recipe := recipes.New(uuid.New(), gofakeit.Adjective()+" "+gofakeit.Dinner(), nil, nil, nil)

			repo := recipes.NewPgRepository(testDBPool)

			err := repo.Save(recipe)
			assert.NoError(t, err, "should not return an error")

			foundRecipe, err := repo.Find(recipe.ID())
			assert.NoError(t, err, "should not return an error")

			assert.Equal(t, *recipe, *foundRecipe, "should return the same recipe")
		})
		t.Run("returns a DuplicateKeyError if the recipe already exists", func(t *testing.T) {
			t.Parallel()
			recipe := recipes.New(fixtures.Recipes[0].ID(), gofakeit.Adjective()+" "+gofakeit.Dinner(), nil, nil, nil)

			repo := recipes.NewPgRepository(testDBPool)

			err := repo.Save(recipe)

			var duplicateKeyError *core.DuplicateKeyError

			assert.ErrorAs(t, err, &duplicateKeyError, "should return a DuplicateKeyError")
		})
		t.Run("returns a ConstraintError if the recipe has an invalid ingredient", func(t *testing.T) {
			t.Parallel()

			recipe := recipes.FakeRecipe()
			recipe.AddIngredients(ingredients.FakeIngredient())

			repo := recipes.NewPgRepository(testDBPool)

			err := repo.Save(&recipe)

			var constraintError *core.ConstraintError

			assert.ErrorAs(t, err, &constraintError, "should return a ConstraintError")
		})
	})
}

///
// func TestRecipesRepository_Save_SavesRecipe(t *testing.T) {
// 	recipe := model.NewRecipe(uuid.New(), gofakeit.Adjective()+" "+gofakeit.Word(), nil, nil, nil)
//
// 	recipe.AddIngredient(testFixtures.Ingredients[5])
// 	recipe.AddIngredient(testFixtures.Ingredients[6])
//
// 	repo := repositories.NewRecipesRepository(testPool)
//
// 	err := repo.Save(recipe)
//
// 	assert.NoError(t, err, "no error")
//
// 	foundRecipe, err := repo.Find(recipe.ID())
// 	assert.NoError(t, err, "should not return an error")
// 	assert.Equal(t, recipe, *foundRecipe, "should return the same recipe")
// }
//
// func TestRecipesRepository_Save_Failes0(t *testing.T) {
// 	recipe := model.NewRecipe(uuid.New(), testFixtures.Recipes[0].Name(), nil, nil, nil)
//
// 	repo := repositories.NewRecipesRepository(testPool)
//
// 	err := repo.Save(recipe)
//
// 	var duplicateKeyError *core.DuplicateKeyError
//
// 	assert.ErrorAs(t, err, &duplicateKeyError, "should return a DuplicateKeyError")
// }
//
// func TestRecipesRepository_Save_Failes(t *testing.T) {
// 	recipe := model.NewRecipe(uuid.New(), gofakeit.Adjective()+" "+gofakeit.Word(), nil, nil, nil)
//
// 	recipe.AddIngredient(testFixtures.Ingredients[5])
// 	recipe.AddIngredient(testFixtures.Ingredients[6])
// 	recipe.AddIngredient(model.NewIngredient(uuid.New(), gofakeit.Word(), nil))
//
// 	repo := repositories.NewRecipesRepository(testPool)
//
// 	err := repo.Save(recipe)
//
// 	var constraintError *core.ConstraintError
//
// 	assert.ErrorAs(t, err, &constraintError, "should return a ConstraintError")
// }
