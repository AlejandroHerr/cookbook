package completions

type Recipe struct {
	Title       string `json:"title"`
	Ingredients []struct {
		Name     string `json:"name"`
		Quantity string `json:"quantity"`
		Unit     string `json:"unit"`
	} `json:"ingredients"`
	Tags        []string `json:"tags"`
	Servings    int      `json:"servings"`
	Steps       []string `json:"steps"`
	Description string   `json:"description"`
	Headline    string   `json:"headline"`
}
