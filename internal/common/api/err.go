package api

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type ErrorResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string      `json:"status"`          // user-level status message
	ErrorText  string      `json:"error,omitempty"` // application-level error message, for debugging
	Details    interface{} `json:"details,omitempty"`
}

func (e *ErrorResponse) Render(_ http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func InternalServerError(err error) *ErrorResponse {
	return &ErrorResponse{
		Err:            err,
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     http.StatusText(http.StatusInternalServerError),
		ErrorText:      err.Error(),
		Details:        nil,
	}
}

func BadRequest(err error) *ErrorResponse {
	return &ErrorResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     http.StatusText(http.StatusBadRequest),
		ErrorText:      err.Error(),
		Details:        nil,
	}
}

func ErrRender(err error) *ErrorResponse {
	return &ErrorResponse{
		Err:            err,
		HTTPStatusCode: http.StatusUnprocessableEntity,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
		Details:        nil,
	}
}

func NotFound(resource string) *ErrorResponse {
	return &ErrorResponse{
		Err:            nil,
		HTTPStatusCode: http.StatusNotFound,
		StatusText:     http.StatusText(http.StatusNotFound),
		ErrorText:      resource + " not found",
		Details:        nil,
	}
}

func ErrConflict(err error) *ErrorResponse {
	return &ErrorResponse{
		Err:            err,
		HTTPStatusCode: http.StatusConflict,
		StatusText:     http.StatusText(http.StatusConflict),
		ErrorText:      err.Error(),
		Details:        nil,
	}
}

type ErrValidationDetail struct {
	Error string `json:"error"`
	Param string `json:"param"`
	Path  string `json:"path"`
}

func ValidationBarRequest(err validator.ValidationErrors) *ErrorResponse {
	details := make([]ErrValidationDetail, 0, len(err))
	for _, e := range err {
		details = append(details, ErrValidationDetail{
			Error: e.Tag(),
			Param: e.Param(),
			Path:  e.Field(),
		})
	}

	return &ErrorResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     http.StatusText(http.StatusBadRequest),
		ErrorText:      "Invalid payload",
		Details:        details,
	}
}
