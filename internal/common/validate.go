package common

import (
	"sync"

	"github.com/go-playground/validator/v10"

	"github.com/AlejandroHerr/cook-book-go/internal/core/model"
)

var (
	once     sync.Once
	validate *validator.Validate
)

func Validator() *validator.Validate {
	once.Do(func() {
		validate = validator.New(validator.WithRequiredStructEnabled())

		err := validate.RegisterValidation("is-unit", model.UnitValidation)
		if err != nil {
			panic(err)
		}
	})

	return validate
}
