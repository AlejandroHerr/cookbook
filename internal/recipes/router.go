package recipes

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"

	"github.com/AlejandroHerr/cookbook/internal/common"
	"github.com/AlejandroHerr/cookbook/internal/common/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

func NewRouter(service *Service, logger *slog.Logger) chi.Router {
	r := chi.NewRouter()

	r.Get("/", api.HandleRendererFunc(getAllRecipesHandler(service, logger), logger))
	r.Post("/", api.HandleRendererFunc(createRecipeHandler(service, logger), logger))
	r.Route("/{recipeIDSlug}", func(r chi.Router) {
		r.Use(recipeCtx(service, logger))
		r.Get("/", api.HandleRendererFunc(getRecipeHandler(logger), logger))
		r.Put("/", api.HandleRendererFunc(updateRecipeHandler(service, logger), logger))
		r.Delete("/", api.HandleRendererFunc(deleteRecipeHandler(service, logger), logger))
	})

	return r
}

func getAllRecipesHandler(service *Service, logger *slog.Logger) api.RendererFunc {
	return func(_ http.ResponseWriter, r *http.Request) render.Renderer {
		list, err := service.List(r.Context())
		if err != nil {
			logger.ErrorContext(r.Context(), "list recipes failed", slog.Any("error", err))

			return api.InternalServerError(err)
		}

		return newRecipesListResponse(list)
	}
}

func createRecipeHandler(service *Service, logger *slog.Logger) api.RendererFunc {
	return func(_ http.ResponseWriter, r *http.Request) render.Renderer {
		request := newCreateUpdateRecipeRequest()
		if err := render.Bind(r, request); err != nil {
			var validationErrors *validator.ValidationErrors
			if as := errors.As(err, &validationErrors); as {
				logger.WarnContext(r.Context(), "request validation failed", slog.Any("error", err))
				return api.ValidationBadRequest(*validationErrors)
			}

			logger.ErrorContext(r.Context(), "bind request failed", slog.Any("error", err))

			return api.BadRequest(err)
		}

		recipe, err := service.Create(r.Context(), request.CreateUpdateRecipeDTO)
		if err != nil {
			logger.ErrorContext(r.Context(), "create recipe failed", slog.Any("error", err))

			var duplicateErr *common.DuplicateError
			if as := errors.As(err, &duplicateErr); as {
				return api.ErrConflict(err)
			}

			return api.InternalServerError(err)
		}

		return newCreateRecipeResponse(recipe)
	}
}

type recipeCtxKey struct{}

func recipeCtx(service *Service, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recipeIDSlug := chi.URLParam(r, "recipeIDSlug")

			recipe, err := service.Get(r.Context(), recipeIDSlug)
			if err != nil {
				logger.ErrorContext(r.Context(), "get recipe failed", slog.Any("error", err))

				var notFoundErr *common.NotFoundError
				if is := errors.As(err, &notFoundErr); is {
					render.Render(w, r, api.NotFound("recipe", err)) //nolint: errcheck
					return
				}

				render.Render(w, r, api.InternalServerError(err)) //nolint: errcheck

				return
			}

			ctx := context.WithValue(r.Context(), recipeCtxKey{}, recipe)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getRecipeHandler(logger *slog.Logger) api.RendererFunc {
	return func(_ http.ResponseWriter, r *http.Request) render.Renderer {
		recipe, ok := r.Context().Value(recipeCtxKey{}).(*Recipe)
		if !ok {
			logger.ErrorContext(r.Context(), "get recipe from context failed")

			return api.NotFound("recipe", nil)
		}

		return newRecipeResponse(recipe)
	}
}

func updateRecipeHandler(service *Service, logger *slog.Logger) api.RendererFunc {
	return func(_ http.ResponseWriter, r *http.Request) render.Renderer {
		recipe, ok := r.Context().Value(recipeCtxKey{}).(*Recipe)
		if !ok {
			logger.ErrorContext(r.Context(), "get recipe from context failed")

			return api.NotFound("recipe", nil)
		}

		request := newCreateUpdateRecipeRequest()
		if err := render.Bind(r, request); err != nil {
			var validationErrors *validator.ValidationErrors
			if as := errors.As(err, &validationErrors); as {
				logger.WarnContext(r.Context(), "request validation failed", slog.Any("error", err))
				return api.ValidationBadRequest(*validationErrors)
			}

			logger.ErrorContext(r.Context(), "bind request failed", slog.Any("error", err))

			return api.BadRequest(err)
		}

		recipe, err := service.Update(r.Context(), recipe, request.CreateUpdateRecipeDTO)
		if err != nil {
			logger.ErrorContext(r.Context(), "update recipe failed", slog.Any("error", err))

			var duplicateErr *common.DuplicateError
			if as := errors.As(err, &duplicateErr); as {
				return api.ErrConflict(err)
			}

			return api.InternalServerError(err)
		}

		return newRecipeResponse(recipe)
	}
}

func deleteRecipeHandler(service *Service, logger *slog.Logger) api.RendererFunc {
	return func(_ http.ResponseWriter, r *http.Request) render.Renderer {
		recipe, ok := r.Context().Value(recipeCtxKey{}).(*Recipe)
		if !ok {
			logger.ErrorContext(r.Context(), "get recipe from context failed")

			return api.NotFound("recipe", nil)
		}

		err := service.Delete(r.Context(), recipe.ID.String())
		if err != nil {
			logger.ErrorContext(r.Context(), "delete recipe failed", slog.Any("error", err))

			return api.InternalServerError(err)
		}

		return &api.NoContentResponse{}
	}
}

var (
	once     sync.Once
	validate *validator.Validate
)

func Validator() *validator.Validate {
	once.Do(func() {
		validate = validator.New(validator.WithRequiredStructEnabled())

		err := validate.RegisterValidation("is-unit", UnitValidation)
		if err != nil {
			panic(err)
		}
	})

	return validate
}

type CreateUpdateRecipeRequest struct {
	*CreateUpdateRecipeDTO
}

func newCreateUpdateRecipeRequest() *CreateUpdateRecipeRequest {
	return &CreateUpdateRecipeRequest{CreateUpdateRecipeDTO: &CreateUpdateRecipeDTO{}}
}

func (req CreateUpdateRecipeRequest) Bind(_ *http.Request) error {
	if err := Validator().Struct(req.CreateUpdateRecipeDTO); err != nil {
		return err
	}

	return nil
}

type RecipesListResponse struct {
	Recipes []Recipe `json:"recipes" tstype:",required"`
}

func newRecipesListResponse(recipes []Recipe) *RecipesListResponse {
	return &RecipesListResponse{Recipes: recipes}
}

func (res RecipesListResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

type CreateRecipeResponse struct {
	*Recipe `tstype:",extends,required"`
}

func newCreateRecipeResponse(recipe *Recipe) *CreateRecipeResponse {
	return &CreateRecipeResponse{
		Recipe: recipe,
	}
}

func (res CreateRecipeResponse) Render(w http.ResponseWriter, _ *http.Request) error {
	w.WriteHeader(http.StatusCreated)

	return nil
}

type RecipeResponse struct {
	*Recipe
}

func newRecipeResponse(recipe *Recipe) *RecipeResponse {
	return &RecipeResponse{
		Recipe: recipe,
	}
}

func (res RecipeResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

type DeleteRecipeResponse struct{}

func (res DeleteRecipeResponse) Render(w http.ResponseWriter, _ *http.Request) error {
	w.WriteHeader(http.StatusNoContent)

	return nil
}
