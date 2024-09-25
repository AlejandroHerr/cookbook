package model

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

type Unit string

func (u Unit) String() string {
	return string(u)
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
