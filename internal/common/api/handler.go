package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
)

type RendererFunc func(http.ResponseWriter, *http.Request) render.Renderer

func HandleRendererFunc(fn RendererFunc, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := fn(w, r)

		if err := render.Render(w, r, resp); err != nil {
			logger.ErrorContext(r.Context(), "error rendering response", slog.Any("error", err))

			render.Render(w, r, ErrRender(err)) //nolint: errcheck
		}
	}
}
