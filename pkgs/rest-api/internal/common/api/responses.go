package api

import (
	"net/http"

	"github.com/go-chi/render"
)

type NoContentResponse struct{}

func NewNoConentResponse() *NoContentResponse {
	return &NoContentResponse{}
}

func (n *NoContentResponse) Render(_ http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusNoContent)

	return nil
}
