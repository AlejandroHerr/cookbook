package ingredients

import (
	"github.com/google/uuid"
)

type Repository interface {
	Find(id uuid.UUID) (*Ingredient, error)
	FindByName(name string) (*Ingredient, error)
	Save(ingredient *Ingredient) error
}
