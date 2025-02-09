package i

import (
	"encoding/json"
	"fmt"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
)

type Ingredient struct {
	id   uuid.UUID
	name string
	kind *string
}

func (i Ingredient) ID() uuid.UUID {
	return i.id
}

func (i Ingredient) Name() string {
	return i.name
}

func (i Ingredient) Kind() *string {
	return i.kind
}

func New(uuid uuid.UUID, name string, kind *string) *Ingredient {
	return &Ingredient{
		id:   uuid,
		name: name,
		kind: kind,
	}
}

func (i *Ingredient) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(&struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
		Kind *string   `json:"kind,omitempty"`
	}{
		ID:   i.id,
		Name: i.name,
		Kind: i.kind,
	})
	if err != nil {
		return nil, fmt.Errorf("error marshaling ingredient: %w", err)
	}

	return data, nil
}

func (i *Ingredient) UnmarshalJSON(data []byte) error {
	var ingredient struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
		Kind *string   `json:"kind,omitempty"`
	}

	if err := json.Unmarshal(data, &ingredient); err != nil {
		return fmt.Errorf("error unmarshaling ingredient: %w", err)
	}

	i.id = ingredient.ID
	i.name = ingredient.Name
	i.kind = ingredient.Kind

	return nil
}

func (i Ingredient) Fake(gofakeit *gofakeit.Faker) (any, error) {
	name := gofakeit.AdjectiveDescriptive() + " " + gofakeit.NounConcrete()
	kind := gofakeit.NounAbstract()

	return Ingredient{
		id:   uuid.New(),
		name: name,
		kind: &kind,
	}, nil
}
