package testutil

import (
	"fmt"

	"github.com/brianvoe/gofakeit/v7"
)

func MustMakeStructFixture(v any) {
	err := gofakeit.Struct(v)
	if err != nil {
		panic(fmt.Errorf("failed to make struct fixture: %w", err))
	}
}
