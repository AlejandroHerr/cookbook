package api

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"` // user-level status message
	// AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(_ http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInternalServerError(err error) *ErrResponse {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     http.StatusText(http.StatusInternalServerError),
		ErrorText:      err.Error(),
	}
}

func ErrBadRequest(err error) *ErrResponse {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     http.StatusText(http.StatusBadRequest),
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) *ErrResponse {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusUnprocessableEntity,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

func ErrNotFound(resource string) *ErrResponse {
	return &ErrResponse{
		Err:            nil,
		HTTPStatusCode: http.StatusNotFound,
		StatusText:     http.StatusText(http.StatusNotFound),
		ErrorText:      resource + " not found",
	}
}

func ErrConflict(err error) *ErrResponse {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusConflict,
		StatusText:     http.StatusText(http.StatusConflict),
		ErrorText:      err.Error(),
	}
}

type ErrValidationDetail struct {
	Error string `json:"error"`
	Param string `json:"param"`
	Path  string `json:"path"`
}

type ErrValidationResponse struct {
	*ErrResponse
	Details []ErrValidationDetail `json:"details"`
}

func NewErrValidationResponse(err validator.ValidationErrors) *ErrValidationResponse {
	details := make([]ErrValidationDetail, 0, len(err))
	for _, e := range err {
		details = append(details, ErrValidationDetail{
			Error: e.Tag(),
			Param: e.Param(),
			Path:  e.Field(),
		})
	}

	return &ErrValidationResponse{
		ErrResponse: &ErrResponse{
			Err:            err,
			HTTPStatusCode: http.StatusBadRequest,
			StatusText:     http.StatusText(http.StatusBadRequest),
			ErrorText:      "Invalid payload",
		},
		Details: details,
	}
}
