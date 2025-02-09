package completions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type OpenAIConfig struct {
	OpenAIKey string `env:"OPENAI_KEY,notEmpty,required" json:"-"`
}

type OpenAIService struct {
	openAiClient *openai.Client
}

var _ Completer = (*OpenAIService)(nil)

func NewOpenAIService(config *OpenAIConfig) *OpenAIService {
	return &OpenAIService{
		openAiClient: openai.NewClient(config.OpenAIKey),
	}
}

const (
	//nolint: lll,unused
	recipePrompt = `Please extract the title and ingredients and some tags to categorize it from this recipe.
  For ingredients please provide:
    - name of the ingredient
    - the amount without unit
    - the unit, has to be wether volume ones (liter, tsp, tbsp,...) or mass (grams)
    - think if the unit must be 'units', ie a countable of the whole, or just it has no unit because it is uncountable.If it is uncountalbe just write none.
  Also tag the recipe, think about the origin of the food, the kind of diet , if it's for winter, kind of food (ie, soup, meal, breakfast)
  Answer everything only in english. If the text provided is not a recipe, do not invent.`
	//nolint: lll
	recipePromptV2 = `Please analyze the recipe I’m providing and return:
The title
A 2-line heading of the recipe, highlighting the most important characteristics of the dish being prepared.
A description of the recipe that is being prepared. It should be one paragraph long or longer. Don't describe the content but the dish being prepared.
Serving count
Ingredients with amounts and units (if the unit is units specify it)(if the amount is a fraction write it with decimals)
If you don't know the unit, do not invent and write uncountable.
If the unit is not a unit of volume or mass, but is made of whole units, write countable.  
Suggested tags that fit the recipe type or diet or origin (like 'low-carb,' 'snack,' or 'breakfast')
Also return me the steps I have to follow.
Answer in english`
)

func (s OpenAIService) CompleteRecipe(ctx context.Context, content string) (*Recipe, error) {
	var result Recipe

	schema, err := jsonschema.GenerateSchemaForType(result)
	if err != nil {
		return nil, fmt.Errorf("error generating schema: %w", err)
	}

	res, err := s.openAiClient.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4o20240806,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: recipePromptV2,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: content,
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "Recipe",
				Schema: schema,
				Strict: true,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create chat completion schema: %w", err)
	}

	reader := chatCompletionStreamReader{res}
	defer func() {
		err = reader.Close()
		if err != nil {
			fmt.Printf("error closing read: %s", err)
		}
	}()

	decoder := json.NewDecoder(reader)

	for {
		if err = decoder.Decode(&result); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("decode stream: %w", err)
		}
	}

	return &result, nil
}

type chatCompletionStreamReader struct {
	*openai.ChatCompletionStream
}

func (r chatCompletionStreamReader) Read(b []byte) (int, error) {
	chunk, err := r.Recv()
	if errors.Is(err, io.EOF) {
		return 0, err //nolint: wrapcheck
	}

	if err != nil {
		return 0, fmt.Errorf("recv stream reader: %w", err)
	}

	delta := chunk.Choices[0].Delta.Content

	n := copy(b, delta)

	return n, nil
}
