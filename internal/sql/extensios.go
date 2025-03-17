package sql

import (
	"fmt"

	"github.com/AlejandroHerr/cookbook/internal/common/utils"
	"github.com/AlejandroHerr/cookbook/internal/recipes"
	"github.com/google/uuid"
)

func (r Recipe) ToModel() (recipes.Recipe, error) {
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return recipes.Recipe{}, fmt.Errorf("parsing id %s: %w", r.ID, err)
	}

	servings, err := utils.Int32PtrToUintPtrSafe(r.Servings)
	if err != nil {
		return recipes.Recipe{}, fmt.Errorf("converting field servings from int32 to uint: %w", err)
	}

	prepTime, err := utils.Int32PtrToUintPtrSafe(r.PrepTime)
	if err != nil {
		return recipes.Recipe{}, fmt.Errorf("econverting field prepTime from int32 to uint: %w", err)
	}

	return recipes.Recipe{
		ID:          id,
		Title:       r.Title,
		Headline:    r.Headline,
		Description: r.Description,
		Steps:       r.Steps,
		Servings:    *servings,
		PrepTime:    prepTime,
		URL:         r.Url,
		Tags:        r.Tags,
		Slug:        r.Slug,
		CreatedAt:   r.CreatedAt.Time,
		UpdatedAt:   r.UpdatedAt.Time,
		Ingredients: nil,
	}, nil
}

func (ri GetRecipeIngredientsRow) ToModel() (recipes.RecipeIngredient, error) {
	parsedID, err := uuid.FromBytes(ri.ID.Bytes[:])
	if err != nil {
		return recipes.RecipeIngredient{}, fmt.Errorf("error converting id %s to uuid: %w", string(ri.ID.Bytes[:]), err)
	}

	u, err := recipes.NewUnit(ri.Unit)
	if err != nil {
		return recipes.RecipeIngredient{}, fmt.Errorf("error converting unit %s to recipes.Unit: %w", ri.Unit, err)
	}

	return recipes.RecipeIngredient{
		ID:       parsedID,
		Name:     *ri.Name,
		Kind:     ri.Kind,
		Unit:     u,
		Quantity: ri.Quantity,
	}, nil
}
