package recipes

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/AlejandroHerr/cook-book-go/internal/common"
	"github.com/AlejandroHerr/cook-book-go/internal/common/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func NewRouter(useCases *UseCases) chi.Router {
	r := chi.NewRouter()

	r.Get("/", getAllRecipesHandler(useCases))
	r.Post("/", createRecipeHandler(useCases))
	r.Route("/{recipeIDSlug}", func(r chi.Router) {
		r.Use(recipeCtx(useCases))
		r.Get("/", getRecipeHandler)
		r.Put("/", updateRecipeHandler(useCases))
		r.Delete("/", deleteRecipeHandler(useCases))
	})
	//
	return r
}

func getAllRecipesHandler(useCases *UseCases) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := useCases.GetAll(r.Context())
		if err != nil {
			render.Render(w, r, api.ErrInternalServerError(err)) //nolint: errcheck
			return
		}

		if err := render.Render(w, r, MakeGetRecipesResponse(list)); err != nil {
			render.Render(w, r, api.ErrRender(err)) //nolint: errcheck
			return
		}
	}
}

func createRecipeHandler(useCases *UseCases) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := makeCreateUpdateRecipeRequest()
		if err := render.Bind(r, request); err != nil {
			if validationErrors, is := err.(validator.ValidationErrors); is {
				render.Render(w, r, api.NewErrValidationResponse(validationErrors)) //nolint: errcheck
				return
			}

			render.Render(w, r, api.ErrBadRequest(err)) //nolint: errcheck

			return
		}

		recipe, err := useCases.Create(r.Context(), request.CreateUpdateRecipeDTO)
		if err != nil {
			var duplicateErr *common.ErrDuplicateKey

			if as := errors.As(err, &duplicateErr); as {
				render.Render(w, r, &api.ErrResponse{ //nolint: errcheck
					Err:            err,
					HTTPStatusCode: http.StatusConflict,
					StatusText:     http.StatusText(http.StatusConflict),
					ErrorText:      "duplicated key '" + duplicateErr.Key + "' found creating recipe.",
				})

				return
			}

			render.Render(w, r, api.ErrInternalServerError(err)) //nolint: errcheck

			return
		}

		if err := render.Render(w, r, makeCreateRecipeResponse(recipe)); err != nil {
			render.Render(w, r, api.ErrRender(err)) //nolint: errcheck
			return
		}
	}
}

type recipeCtxKey struct{}

func recipeCtx(useCases *UseCases) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recipeIDSlug := chi.URLParam(r, "recipeIDSlug")

			recipe, err := useCases.Get(r.Context(), recipeIDSlug)
			if err != nil {
				var notFoundErr *common.ErrNotFound
				if is := errors.As(err, &notFoundErr); is {
					render.Render(w, r, api.ErrNotFound("recipe")) //nolint: errcheck
					return
				}

				render.Render(w, r, api.ErrInternalServerError(err)) //nolint: errcheck

				return
			}

			ctx := context.WithValue(r.Context(), recipeCtxKey{}, recipe)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getRecipeHandler(w http.ResponseWriter, r *http.Request) {
	recipe, ok := r.Context().Value(recipeCtxKey{}).(*Recipe)
	if !ok {
		render.Render(w, r, api.ErrNotFound("recipe")) //nolint: errcheck
		return
	}

	if err := render.Render(w, r, makeGetRecipeResponse(recipe)); err != nil {
		render.Render(w, r, api.ErrRender(err)) //nolint: errcheck
		return
	}
}

func updateRecipeHandler(useCases *UseCases) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		recipe, ok := r.Context().Value(recipeCtxKey{}).(*Recipe)
		if !ok {
			render.Render(w, r, api.ErrNotFound("recipe")) //nolint: errcheck
			return
		}

		request := makeCreateUpdateRecipeRequest()
		if err := render.Bind(r, request); err != nil {
			if validationErrors, is := err.(validator.ValidationErrors); is {
				render.Render(w, r, api.NewErrValidationResponse(validationErrors)) //nolint: errcheck
				return
			}

			render.Render(w, r, api.ErrBadRequest(err)) //nolint: errcheck

			return
		}

		recipe, err := useCases.Update(r.Context(), recipe.ID, request.CreateUpdateRecipeDTO)
		if err != nil {
			var duplicateErr *common.ErrDuplicateKey

			if as := errors.As(err, &duplicateErr); as {
				render.Render(w, r, &api.ErrResponse{ //nolint: errcheck
					Err:            err,
					HTTPStatusCode: http.StatusConflict,
					StatusText:     http.StatusText(http.StatusConflict),
					ErrorText:      "duplicated key '" + duplicateErr.Key + "' found updating recipe.",
				})

				return
			}

			render.Render(w, r, api.ErrInternalServerError(err)) //nolint: errcheck

			return
		}

		if err := render.Render(w, r, makeUpdateUpdateRecipeResponse(recipe)); err != nil {
			render.Render(w, r, api.ErrRender(err)) //nolint: errcheck
			return
		}
	}
}

func deleteRecipeHandler(useCases *UseCases) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		recipe, ok := r.Context().Value(recipeCtxKey{}).(*Recipe)
		if !ok {
			render.Render(w, r, api.ErrNotFound("recipe")) //nolint: errcheck
			return
		}

		err := useCases.Delete(r.Context(), recipe.ID.String())
		if err != nil {
			render.Render(w, r, api.ErrInternalServerError(err)) //nolint: errcheck
			return
		}

		w.WriteHeader(http.StatusNoContent)
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

type RecipeWithoutIngredients struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Headline    *string   `json:"headline"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Description *string   `json:"description,omitempty"`
	Steps       *string   `json:"steps,omitempty"`
	Servings    *uint     `json:"servings,omitempty"`
	URL         *string   `json:"url,omitempty"`
	Tags        []string  `json:"tags"`
	Slug        string    `json:"slug"`
}

type GetRecipesResponse struct {
	Recipes []*RecipeWithoutIngredients `json:"recipes"`
}

func (res GetRecipesResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

func MakeGetRecipesResponse(recipes []*Recipe) *GetRecipesResponse {
	list := []*RecipeWithoutIngredients{}
	for _, r := range recipes {
		list = append(list, &RecipeWithoutIngredients{
			ID:          r.ID,
			Title:       r.Title,
			Headline:    r.Headline,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
			Description: r.Description,
			Steps:       r.Steps,
			Servings:    r.Servings,
			URL:         r.URL,
			Tags:        r.Tags,
			Slug:        r.Slug(),
		})
	}

	return &GetRecipesResponse{Recipes: list}
}

type createUpdateRecipeRequest struct {
	*CreateUpdateRecipeDTO
}

func makeCreateUpdateRecipeRequest() *createUpdateRecipeRequest {
	return &createUpdateRecipeRequest{
		CreateUpdateRecipeDTO: &CreateUpdateRecipeDTO{}, //nolint:exhaustruct
	}
}

func (req createUpdateRecipeRequest) Bind(_ *http.Request) error {
	if err := Validator().Struct(req); err != nil {
		return err //nolint:wrapcheck
	}

	return nil
}

type CreateRecipeResponse struct {
	*Recipe
	Slug string `json:"slug"`
}

func makeCreateRecipeResponse(recipe *Recipe) *CreateRecipeResponse {
	return &CreateRecipeResponse{
		Recipe: recipe,
		Slug:   recipe.Slug(),
	}
}

func (res CreateRecipeResponse) Render(w http.ResponseWriter, _ *http.Request) error {
	w.WriteHeader(http.StatusCreated)

	return nil
}

type GetRecipeResponse struct {
	*Recipe
	Slug string `json:"slug"`
}

func makeGetRecipeResponse(recipe *Recipe) *GetRecipeResponse {
	return &GetRecipeResponse{
		Recipe: recipe,
		Slug:   recipe.Slug(),
	}
}

func (res GetRecipeResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

type UpdateRecipeResponse struct {
	*Recipe
	Slug string `json:"slug"`
}

func makeUpdateUpdateRecipeResponse(recipe *Recipe) *UpdateRecipeResponse {
	return &UpdateRecipeResponse{
		Recipe: recipe,
		Slug:   recipe.Slug(),
	}
}

func (res UpdateRecipeResponse) Render(w http.ResponseWriter, _ *http.Request) error {
	w.WriteHeader(http.StatusOK)

	return nil
}
