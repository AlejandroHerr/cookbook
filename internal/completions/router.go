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

func NewRouter(service *Service, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Post("/recipe", api.HandleRendererFunc(completeRecipeHandler(service, logger), logger))

	return r
}

func completeRecipeHandler(service *Service, logger *slog.Logger) api.RendererFunc {
	return func(_ http.ResponseWriter, r *http.Request) render.Renderer {
		request := &CompleteRecipeRequest{} //nolint:exhaustruct
		if err := render.Bind(r, request); err != nil {
			var validationErrors *validator.ValidationErrors
			if as := errors.As(err, &validationErrors); as {
				logger.WarnContext(r.Context(), "request validation failed", slog.Any("error", err))

				return api.ValidationBadRequest(*validationErrors)
			}

			logger.ErrorContext(r.Context(), "bind request failed", slog.Any("error", err))

			return api.BadRequest(err)
		}

		recipe, err := service.CompleteRecipe(r.Context(), request.URL)
		if err != nil {
			logger.ErrorContext(r.Context(), "complete recipe failed", slog.Any("error", err))

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
