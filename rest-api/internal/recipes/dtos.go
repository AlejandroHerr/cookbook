package recipes

import (
	"fmt"

	"github.com/brianvoe/gofakeit/v7"
)

type CreateUpdateRecipeDTO struct {
	Title       string                      `json:"title" validate:"required,min=1"`
	URL         string                      `json:"url" validate:"omitempty,url,min=1"`
	Tags        []string                    `json:"tags" validate:"omitempty,dive,min=1"`
	Ingredients []CreateRecipeIngredientDTO `json:"ingredients" validate:"omitempty,required,dive,required"`
	Servings    uint                        `json:"servings" validate:"omitempty,gte=1"`
	Description string                      `json:"description" validate:"omitempty,min=1"`
	Headline    string                      `json:"headline" validate:"omitempty,min=1"`
	Steps       string                      `json:"steps" validate:"omitempty,min=1"`
}

type CreateRecipeIngredientDTO struct {
	Quantity float64 `json:"quantity" validate:"required"`
	Unit     Unit    `json:"unit" validate:"required,is-unit"`
	Name     string  `json:"name" validate:"required"`
}

func (i CreateUpdateRecipeDTO) Fake(gofakeit *gofakeit.Faker) (any, error) {
	ingredients := make([]CreateRecipeIngredientDTO, 5)

	for i := range ingredients {
		var ingredient CreateRecipeIngredientDTO

		err := gofakeit.Struct(&ingredient)
		if err != nil {
			return nil, fmt.Errorf("faking recipe ingredient: %w", err)
		}

		ingredients[i] = ingredient
	}

	return CreateUpdateRecipeDTO{
		Title:       gofakeit.Sentence(10),
		Description: gofakeit.Sentence(100),
		Ingredients: ingredients,
		Tags:        []string{gofakeit.Adjective(), gofakeit.Adjective()},
		Servings:    gofakeit.UintRange(0, 100),
		URL:         gofakeit.URL(),
		Headline:    gofakeit.Sentence(10),
		Steps:       gofakeit.Sentence(10),
	}, nil
}

func (i CreateRecipeIngredientDTO) Fake(gofakeit *gofakeit.Faker) (any, error) {
	name := gofakeit.Adjective() + " " + gofakeit.Name()
	unit := Units[gofakeit.IntRange(0, len(Units)-1)]

	return CreateRecipeIngredientDTO{
		Name:     name,
		Unit:     unit,
		Quantity: gofakeit.Float64Range(0, 100),
	}, nil
}
