package suggestions

import (
	"net/http"

	"github.com/AlejandroHerr/cookbook/internal/common/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func MakeRouter(useCases *UseCases) chi.Router {
	r := chi.NewRouter()

	r.Get("/ingredients", getOptionsHander(useCases, "ingredients"))
	r.Get("/tags", getOptionsHander(useCases, "tags"))
	r.Get("/units", getOptionsHander(useCases, "units"))

	return r
}

func getOptionsHander(useCases *UseCases, entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")

		var options []Option

		var err error

		switch entity {
		case "ingredients":
			options, err = useCases.GetIngredientsOptions(r.Context(), search)
		case "tags":
			options, err = useCases.GetTagsOptions(r.Context(), search)
		case "units":
			options, err = useCases.GetUnitsOptions(r.Context())
		default:
			render.Render(w, r, api.ErrNotFound(entity+" options")) //nolint: errcheck

			return
		}

		if err != nil {
			render.Render(w, r, api.ErrInternalServerError(err)) //nolint: errcheck
			return
		}

		if err := render.RenderList(w, r, MakeSuggestionResponse(options)); err != nil {
			render.Render(w, r, api.ErrRender(err)) //nolint: errcheck
			return
		}
	}
}

type SuggestionReponse struct {
	*Option
}

func MakeSuggestionReponse(option *Option) *SuggestionReponse {
	resp := &SuggestionReponse{Option: option}

	return resp
}

func (rd *SuggestionReponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

func MakeSuggestionResponse(options []Option) []render.Renderer {
	list := []render.Renderer{}
	for _, option := range options {
		list = append(list, MakeSuggestionReponse(&option))
	}

	return list
}
