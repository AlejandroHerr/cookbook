package recipes

import (
	"errors"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/go-playground/validator/v10"
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

func (r Recipe) Slug() string {
	return slug.Make(r.Title)
}

func (r Recipe) Fake(faker *gofakeit.Faker) (any, error) {
	title := faker.Adjective() + " " + faker.Adjective() + " " + faker.Dinner()
	description := faker.LoremIpsumParagraph(2, 3, 5, ".")
	headline := faker.LoremIpsumParagraph(2, 3, 5, ".")
	steps := faker.LoremIpsumParagraph(2, 3, 5, ".")
	url := faker.URL()
	servings := faker.UintRange(1, 10)

	tags := make([]string, 2)
	for i := 0; i < len(tags); i++ {
		tags[i] = faker.Word()
	}

	ingredients := make([]RecipeIngredient, 5)
	for i := 0; i < len(ingredients); i++ {
		ingredient, err := RecipeIngredient{}.Fake(faker) //nolint: exhaustruct
		if err != nil {
			return nil, err
		}

		ingredients[i] = ingredient.(RecipeIngredient) //nolint:errcheck
	}

	return Recipe{
		ID:          uuid.New(),
		Title:       title,
		Description: &description,
		URL:         &url,
		Tags:        tags,
		Ingredients: ingredients,
		Headline:    &headline,
		Steps:       &steps,
		Servings:    &servings,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

type RecipeIngredient struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Kind     *string   `json:"kind"`
	Unit     Unit      `json:"unit"`
	Quantity float64   `json:"amount"`
}

func (ri RecipeIngredient) Fake(faker *gofakeit.Faker) (any, error) {
	name := gofakeit.AdjectiveDescriptive() + " " + gofakeit.AdjectiveDescriptive() + " " + gofakeit.NounConcrete()
	kind := faker.Adjective()
	unit := faker.RandomString([]string{"g", "kg", "ml", "l", "tsp", "tbsp", "cup", "qt", "countable", "uncountable"})
	quantity := faker.Float64Range(0.1, 1000)

	return RecipeIngredient{
		ID:       uuid.New(),
		Name:     name,
		Kind:     &kind,
		Unit:     Unit(unit),
		Quantity: quantity,
	}, nil
}

type Unit string

func (u Unit) String() string {
	return string(u)
}

func (u Unit) Fake(faker *gofakeit.Faker) (any, error) {
	return faker.RandomString([]string{"g", "kg", "ml", "l", "tsp", "tbsp", "cup", "qt", "countable", "uncountable"}), nil
}

const (
	Kilo        Unit = "kilo"
	Gram        Unit = "g"
	Milligram   Unit = "mg"
	Liter       Unit = "l"
	Milliliter  Unit = "ml"
	Teaspoon    Unit = "tsp"
	Tablespoon  Unit = "tbsp"
	Cup         Unit = "cup"
	Quart       Unit = "qt"
	Countable   Unit = "countable"
	Uncountable Unit = "uncountable"
)

func NewUnit(s string) (Unit, error) {
	if !isUnit(s) {
		return "", errors.New("not a valid unit")
	}

	return Unit(s), nil
}

var (
	Units = []Unit{
		Kilo,
		Gram,
		Milligram,
		Liter,
		Milliliter,
		Teaspoon,
		Tablespoon,
		Cup,
		Quart,
		Countable,
		Uncountable,
	}
	UnitDisplayNames = map[Unit]string{
		Kilo:        "kg",
		Gram:        "g",
		Milligram:   "mg",
		Liter:       "L",
		Teaspoon:    "tsp",
		Tablespoon:  "tbsp",
		Cup:         "cup",
		Quart:       "qt",
		Countable:   "unit(s)",
		Uncountable: "some",
	}
)

func UnitValidation(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	return isUnit(value)
}

func isUnit(s string) bool {
	for _, u := range Units {
		if s == string(u) {
			return true
		}
	}

	return false
}
