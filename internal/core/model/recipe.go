package model

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

type Recipe struct {
	ID          uuid.UUID          `json:"id"`
	Title       string             `json:"title"`
	Headline    *string            `json:"headline"`
	Description *string            `json:"description,omitempty"`
	Steps       *string            `json:"steps"`
	Servings    *uint              `json:"servings"`
	URL         *string            `json:"url,omitempty"`
	Tags        []string           `json:"tags"`
	Ingredients []RecipeIngredient `json:"ingredients"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
}

type RecipeIngredient struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Kind     *string   `json:"kind"`
	Unit     Unit      `json:"unit"`
	Quantity float64   `json:"amount"`
}

func NewRecipe(
	id uuid.UUID,
	title string,
	headline *string,
	description *string,
	steps *string,
	servings *uint,
	url *string,
	tags []string,
	ingredientsList []RecipeIngredient,
) *Recipe {
	tagsList := tags
	if tagsList == nil {
		tagsList = make([]string, 0)
	}

	if ingredientsList == nil {
		ingredientsList = make([]RecipeIngredient, 0)
	}

	return &Recipe{
		ID:          id,
		Title:       title,
		Headline:    headline,
		Description: description,
		Steps:       steps,
		Servings:    servings,
		URL:         url,
		Tags:        tagsList,
		Ingredients: ingredientsList,
	}
}

func (r Recipe) Slug() string {
	return slug.Make(r.Title)
}

func (r Recipe) Fake(faker *gofakeit.Faker) (any, error) {
	title := faker.Adjective() + " " + faker.Dinner()
	description := faker.LoremIpsumParagraph(2, 3, 5, ".")
	headline := faker.LoremIpsumParagraph(2, 3, 5, ".")
	steps := faker.LoremIpsumParagraph(2, 3, 5, ".")
	url := faker.URL()

	return Recipe{
		ID:          uuid.New(),
		Title:       title,
		Description: &description,
		URL:         &url,
		Tags:        make([]string, 0),
		Ingredients: make([]RecipeIngredient, 0),
		Headline:    &headline,
		Steps:       &steps,
		Servings:    nil,
	}, nil
}
