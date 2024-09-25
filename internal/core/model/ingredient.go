package model

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
)

type Ingredient struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Kind *string   `json:"kind,omitempty"`
}

func NewIngredient(id uuid.UUID, name string, kind *string, createdAt *time.Time, updatedAt *time.Time) *Ingredient {
	return &Ingredient{
		ID:   id,
		Name: name,
		Kind: kind,
	}
}

func (i Ingredient) Fake(gofakeit *gofakeit.Faker) (any, error) {
	name := gofakeit.AdjectiveDescriptive() + " " + gofakeit.NounConcrete()
	kind := gofakeit.NounAbstract()

	return Ingredient{
		ID:   uuid.New(),
		Name: name,
		Kind: &kind,
	}, nil
}
