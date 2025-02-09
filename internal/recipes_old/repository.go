package recipes

import (
	"github.com/google/uuid"
)

type Repository interface {
	Find(id uuid.UUID) (*Recipe, error)
	FindByName(name string) (*Recipe, error)
	Save(recipe *Recipe) error
}
