package recipes

import (
	"encoding/json"
	"fmt"

	"github.com/AlejandroHerr/cookbook/internal/ingredients"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
)

type Recipe struct {
	id          uuid.UUID
	name        string
	description *string
	url         *string
	ingredients []ingredients.Ingredient
}

func New(
	id uuid.UUID,
	name string,
	description *string,
	url *string,
	ingredientsList []ingredients.Ingredient,
) *Recipe {
	if ingredientsList == nil {
		ingredientsList = make([]ingredients.Ingredient, 0)
	}

	return &Recipe{
		id:          id,
		name:        name,
		description: description,
		url:         url,
		ingredients: ingredientsList,
	}
}

func (r Recipe) ID() uuid.UUID {
	return r.id
}

func (r Recipe) Name() string {
	return r.name
}

func (r Recipe) Description() *string {
	return r.description
}

func (r Recipe) URL() *string {
	return r.url
}

func (r Recipe) Ingredients() []ingredients.Ingredient {
	return r.ingredients
}

func (r *Recipe) AddIngredients(ingredients ...ingredients.Ingredient) {
	r.ingredients = append(r.ingredients, ingredients...)
}

func (r *Recipe) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(&struct {
		ID          uuid.UUID                `json:"id"`
		Name        string                   `json:"name"`
		Description *string                  `json:"description,omitempty"`
		URL         *string                  `json:"url,omitempty"`
		Ingredients []ingredients.Ingredient `json:"ingredients"`
	}{
		ID:          r.id,
		Name:        r.name,
		Description: r.description,
		URL:         r.url,
		Ingredients: r.ingredients,
	})
	if err != nil {
		return nil, fmt.Errorf("error marshaling recipe: %w", err)
	}

	return data, nil
}

func (r *Recipe) UnmarshalJSON(data []byte) error {
	var recipe struct {
		ID          uuid.UUID                `json:"id"`
		Name        string                   `json:"name"`
		Description *string                  `json:"description,omitempty"`
		URL         *string                  `json:"url,omitempty"`
		Ingredients []ingredients.Ingredient `json:"ingredients,omitempty"`
	}

	err := json.Unmarshal(data, &recipe)
	if err != nil {
		return fmt.Errorf("error unmarshaling recipe: %w", err)
	}

	r.id = recipe.ID
	r.name = recipe.Name
	r.description = recipe.Description
	r.url = recipe.URL

	if recipe.Ingredients == nil {
		recipe.Ingredients = make([]ingredients.Ingredient, 0)
	} else {
		r.ingredients = recipe.Ingredients
	}

	return nil
}

func (r Recipe) Fake(faker *gofakeit.Faker) (any, error) {
	name := faker.Adjective() + " " + faker.Dinner()
	descrption := faker.LoremIpsumParagraph(1, 3, 6, " ")
	url := faker.URL()

	return Recipe{
		id:          uuid.New(),
		name:        name,
		description: &descrption,
		url:         &url,
		ingredients: make([]ingredients.Ingredient, 0),
	}, nil
}
