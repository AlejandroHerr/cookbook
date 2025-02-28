package completions

import "github.com/AlejandroHerr/cookbook/internal/recipes"

type Recipe struct {
	Title       string `json:"title"`
	Ingredients []struct {
		Name     string       `json:"name"`
		Quantity float64      `json:"quantity"`
		Unit     recipes.Unit `json:"unit"`
	} `json:"ingredients"`
	Tags        []string `json:"tags"`
	Servings    int      `json:"servings"`
	Steps       []string `json:"steps"`
	Description string   `json:"description"`
	Headline    string   `json:"headline"`
	PrepTime    uint     `json:"prepTime"`
}
