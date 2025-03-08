package suggestions

import (
	"log/slog"
	"net/http"

	"github.com/AlejandroHerr/cookbook/internal/common/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func MakeRouter(useCases *UseCases, logger *slog.Logger) chi.Router {
	l := logger.With(slog.String("service", "suggestions-router"))
	r := chi.NewRouter()

	r.Get("/ingredients", getOptionsHander(useCases, "ingredients", l))
	r.Get("/tags", getOptionsHander(useCases, "tags", l))

	return r
}

func getOptionsHander(useCases *UseCases, entity string, logger *slog.Logger) http.HandlerFunc {
	return api.HandleRendererFunc(func(w http.ResponseWriter, r *http.Request) render.Renderer {
		search := r.URL.Query().Get("search")

		var options []Option

		var err error

		switch entity {
		case "ingredients":
			options, err = useCases.GetIngredientsOptions(r.Context(), search)
		case "tags":
			options, err = useCases.GetTagsOptions(r.Context(), search)
		default:
			return api.NotFound(entity + " options")

		}

		if err != nil {
			return api.InternalServerError(err)
		}

		return &GetSuggestionsReponse{Options: options}
	}, logger)
}

type GetSuggestionsReponse struct {
	Options []Option `json:"options" tstype:",required"`
}

func MakeSuggestionReponse(options []Option) *GetSuggestionsReponse {
	resp := &GetSuggestionsReponse{Options: options}

	return resp
}

func (rd *GetSuggestionsReponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}
