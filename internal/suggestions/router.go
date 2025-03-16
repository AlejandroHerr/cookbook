package suggestions

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/AlejandroHerr/cookbook/internal/common"
	"github.com/AlejandroHerr/cookbook/internal/common/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func NewRouter(service *Service, logger *slog.Logger) chi.Router {
	r := chi.NewRouter()

	r.Get("/{entity}", api.HandleRendererFunc(getSuggestionsHandler(service, logger), logger))

	return r
}

func getSuggestionsHandler(service *Service, logger *slog.Logger) api.RendererFunc {
	return func(_ http.ResponseWriter, r *http.Request) render.Renderer {
		entity := chi.URLParam(r, "entity")
		if entity == "" {
			logger.ErrorContext(r.Context(), "entity is required")

			return api.BadRequest(errors.New("entity is required"))
		}

		search := r.URL.Query().Get("search")

		options, err := service.GetSuggestions(r.Context(), entity, search)
		if err != nil {
			logger.ErrorContext(r.Context(), "get suggestions failed", slog.Any("error", err))

			var notFoundErr *common.NotFoundError
			if errors.As(err, &notFoundErr) {
				return api.BadRequest(err)
			}

			return api.InternalServerError(err)
		}

		return newSuggestionReponse(options)
	}
}

type GetSuggestionsReponse struct {
	Options []Option `json:"options"`
}

func newSuggestionReponse(options []Option) *GetSuggestionsReponse {
	resp := &GetSuggestionsReponse{Options: options}

	return resp
}

func (rd *GetSuggestionsReponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}
