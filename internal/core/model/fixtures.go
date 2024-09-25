package model

import "github.com/brianvoe/gofakeit/v7"

type Fixtures struct {
	Ingredients []*Ingredient
	Recipes     []*Recipe
}

func MustMakeFixtures(count int) *Fixtures {
	ingredients := make([]*Ingredient, count)
	for i := 0; i < count; i++ {
		err := gofakeit.Struct(&ingredients[i])
		if err != nil {
			panic(err)
		}
	}

	recipes := make([]*Recipe, count)
	for i := 0; i < count; i++ {
		err := gofakeit.Struct(&recipes[i])
		if err != nil {
			panic(err)
		}

		recipeIngredients := make([]RecipeIngredient, 3)

		for j := 0; j < 3; j++ {
			ingredient := ingredients[(i*3+j)%count]

			i := RecipeIngredient{
				ID:       ingredient.ID,
				Name:     ingredient.Name,
				Kind:     ingredient.Kind,
				Unit:     Kilo,
				Quantity: gofakeit.Float64Range(0, 10),
			}

			recipeIngredients[j] = i
		}

		recipes[i].Ingredients = recipeIngredients
	}

	return &Fixtures{
		Ingredients: ingredients,
		Recipes:     recipes,
	}
}
