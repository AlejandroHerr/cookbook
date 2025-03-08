package completions

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/AlejandroHerr/cookbook/internal/common/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

func MakeRouter(useCases *UseCases, logger *slog.Logger) http.Handler {
	l := logger.With(slog.String("service", "completions-router"))
	r := chi.NewRouter()

	r.Post("/recipe", api.HandleRendererFunc(completeRecipeHandler(useCases), l))

	return r
}

func completeRecipeHandler(useCases *UseCases) api.RendererFunc {
	return func(w http.ResponseWriter, r *http.Request) render.Renderer {
		request := &CompleteRecipeRequest{} //nolint:exhaustruct
		if err := render.Bind(r, request); err != nil {
			var validationErrors *validator.ValidationErrors
			if as := errors.As(err, &validationErrors); as {
				return api.ValidationBarRequest(*validationErrors)
			}

			return api.BadRequest(err)

		}

		recipe, err := useCases.CompleteRecipe(r.Context(), request.URL)
		if err != nil {
			return api.InternalServerError(err)
		}

		return &CompleteRecipeResponse{Recipe: *recipe}
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
