package intelligence

import (
	"errors"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
)

type chatCompletionStreamReader struct {
	*openai.ChatCompletionStream
}

func (r chatCompletionStreamReader) Read(b []byte) (int, error) {
	chunk, err := r.Recv()
	if errors.Is(err, io.EOF) {
		return 0, err //nolint: wrapcheck
	}

	if err != nil {
		return 0, fmt.Errorf("error receiving from stream: %w", err)
	}

	delta := chunk.Choices[0].Delta.Content

	n := copy(b, delta)

	return n, nil
}
