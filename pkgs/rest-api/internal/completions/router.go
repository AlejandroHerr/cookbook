package completions

import (
	"net/http"

	"github.com/AlejandroHerr/cookbook/internal/common/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

func MakeRouter(useCases *UseCases) http.Handler {
	r := chi.NewRouter()

	r.Post("/recipe", completeRecipeHandler(useCases))

	return r
}

func completeRecipeHandler(useCases *UseCases) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := &CompleteRecipeRequest{} //nolint:exhaustruct
		if err := render.Bind(r, request); err != nil {
			if validationErrors, is := err.(validator.ValidationErrors); is {
				render.Render(w, r, api.NewErrValidationResponse(validationErrors)) //nolint: errcheck
				return
			}

			render.Render(w, r, api.ErrBadRequest(err)) //nolint: errcheck

			return
		}

		recipe, err := useCases.CompleteRecipe(r.Context(), request.URL)
		if err != nil {
			render.Render(w, r, api.ErrInternalServerError(err)) //nolint: errcheck
			return
		}

		if err := render.Render(w, r, &CompleteRecipeResponse{Recipe: *recipe}); err != nil {
			render.Render(w, r, api.ErrRender(err)) //nolint: errcheck
			return
		}
	}
}

type CompleteRecipeRequest struct {
	URL string `json:"url" validate:"required,url"`
}

func (req CompleteRecipeRequest) Bind(_ *http.Request) error {
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(req); err != nil {
		return err //nolint:wrapcheck
	}

	return nil
}

type CompleteRecipeResponse struct {
	Recipe Recipe `json:"recipe" tstype:",required"`
}

func (res CompleteRecipeResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}
