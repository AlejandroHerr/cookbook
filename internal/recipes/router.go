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

func MakeRouter(useCases *UseCases, logger *slog.Logger) chi.Router {
	l := logger.With(slog.String("service", "recipes-router"))
	r := chi.NewRouter()

	r.Get("/", api.HandleRendererFunc(getAllRecipesHandler(useCases), l))
	r.Post("/", api.HandleRendererFunc(createRecipeHandler(useCases), l))
	r.Route("/{recipeIDSlug}", func(r chi.Router) {
		r.Use(recipeCtx(useCases))
		r.Get("/", api.HandleRendererFunc(getRecipeHandler, l))
		r.Put("/", api.HandleRendererFunc(updateRecipeHandler(useCases), l))
		r.Delete("/", api.HandleRendererFunc(deleteRecipeHandler(useCases), l))
	})

	return r
}

func getAllRecipesHandler(useCases *UseCases) api.RendererFunc {
	return func(w http.ResponseWriter, r *http.Request) render.Renderer {
		list, err := useCases.GetAll(r.Context())
		if err != nil {
			return api.InternalServerError(err)
		}

		return newRecipesListResponse(list)
	}
}

func createRecipeHandler(useCases *UseCases) api.RendererFunc {
	return func(w http.ResponseWriter, r *http.Request) render.Renderer {
		request := newCreateUpdateRecipeRequest()
		if err := render.Bind(r, request); err != nil {
			var validationErrors *validator.ValidationErrors
			if as := errors.As(err, &validationErrors); as {
				return api.ValidationBarRequest(*validationErrors)
			}

			return api.BadRequest(err)
		}

		recipe, err := useCases.Create(r.Context(), request.CreateUpdateRecipeDTO)
		if err != nil {
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

func recipeCtx(useCases *UseCases) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recipeIDSlug := chi.URLParam(r, "recipeIDSlug")

			recipe, err := useCases.Get(r.Context(), recipeIDSlug)
			if err != nil {
				var notFoundErr *common.NotFoundError
				if is := errors.As(err, &notFoundErr); is {
					render.Render(w, r, api.NotFound("recipe")) //nolint: errcheck
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

func getRecipeHandler(w http.ResponseWriter, r *http.Request) render.Renderer {
	recipe, ok := r.Context().Value(recipeCtxKey{}).(*Recipe)
	if !ok {
		return api.NotFound("recipe")
	}

	return newRecipeResponse(recipe)
}

func updateRecipeHandler(useCases *UseCases) api.RendererFunc {
	return func(w http.ResponseWriter, r *http.Request) render.Renderer {
		recipe, ok := r.Context().Value(recipeCtxKey{}).(*Recipe)
		if !ok {
			return api.NotFound("recipe")
		}

		request := newCreateUpdateRecipeRequest()
		if err := render.Bind(r, request); err != nil {
			var validationErrors *validator.ValidationErrors
			if as := errors.As(err, &validationErrors); as {
				return api.ValidationBarRequest(*validationErrors)
			}

			return api.BadRequest(err)
		}

		recipe, err := useCases.Update(r.Context(), recipe, request.CreateUpdateRecipeDTO)
		if err != nil {
			var duplicateErr *common.DuplicateError
			if as := errors.As(err, &duplicateErr); as {
				return api.ErrConflict(err)
			}

			return api.InternalServerError(err)
		}

		return newRecipeResponse(recipe)
	}
}

func deleteRecipeHandler(useCases *UseCases) api.RendererFunc {
	return func(w http.ResponseWriter, r *http.Request) render.Renderer {
		recipe, ok := r.Context().Value(recipeCtxKey{}).(*Recipe)
		if !ok {
			return api.NotFound("recipe")
		}

		err := useCases.Delete(r.Context(), recipe.ID.String())
		if err != nil {
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

type CreatedRecipeResponse struct {
	*Recipe
}

func makeCreatedRecipeResponse(recipe *Recipe) *CreatedRecipeResponse {
	return &CreatedRecipeResponse{
		Recipe: recipe,
	}
}

func (res CreatedRecipeResponse) Render(w http.ResponseWriter, _ *http.Request) error {
	w.WriteHeader(http.StatusCreated)

	return nil
}

type DeleteRecipeResponse struct{}

func (res DeleteRecipeResponse) Render(w http.ResponseWriter, _ *http.Request) error {
	w.WriteHeader(http.StatusNoContent)

	return nil
}
