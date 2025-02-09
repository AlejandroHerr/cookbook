package completions

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type (
	Error struct {
		Err error
	}
	HTTPClientError struct {
		StatusCode int    `json:"statusCode"`
		Response   string `json:"response"`
	}
)

func (e *HTTPClientError) Error() string {
	return fmt.Sprintf("http client error with status %d and response %s", e.StatusCode, e.Response)
}

func (e *Error) Error() string {
	return e.Err.Error()
}

type HTTPScrapper struct{}

var _ Scrapper = (*HTTPScrapper)(nil)

func NewHTTPScrapper() *HTTPScrapper {
	return &HTTPScrapper{}
}

func (s HTTPScrapper) Scrap(ctx context.Context, url string) (string, error) {
	handlerCtx, cancel := context.WithTimeoutCause(ctx, 10*time.Second, errors.New("request timeout"))
	defer cancel()

	req, err := http.NewRequestWithContext(handlerCtx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("new request with context: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		response := new(strings.Builder)

		_, err = io.Copy(response, res.Body)
		if err != nil {
			return "", fmt.Errorf("read response: %w", err)
		}

		return "", &HTTPClientError{StatusCode: res.StatusCode, Response: response.String()}
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", &Error{Err: fmt.Errorf("new document from reade: %w", err)}
	}

	doc.Find(
		"script,header,iframe,link, img, svg, nav, style, .ad, .footer, .header, .sidebar,link",
	).Remove()

	content := doc.Find("body").Text()

	return content, nil
}
