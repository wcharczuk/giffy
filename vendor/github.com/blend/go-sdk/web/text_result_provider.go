package web

import (
	"fmt"
	"net/http"

	"github.com/blend/go-sdk/logger"
)

// NewTextResultProvider returns a new text result provider.
func NewTextResultProvider(log *logger.Logger) *TextResultProvider {
	return &TextResultProvider{log: log}
}

// TextResultProvider is the default response provider if none is specified.
type TextResultProvider struct {
	log *logger.Logger
}

// NotFound returns a text response.
func (trp *TextResultProvider) NotFound() Result {
	return &RawResult{
		StatusCode:  http.StatusNotFound,
		ContentType: ContentTypeText,
		Body:        []byte("Not Found"),
	}
}

// NotAuthorized returns a text response.
func (trp *TextResultProvider) NotAuthorized() Result {
	return &RawResult{
		StatusCode:  http.StatusForbidden,
		ContentType: ContentTypeText,
		Body:        []byte("Not Authorized"),
	}
}

// InternalError returns a text response.
func (trp *TextResultProvider) InternalError(err error) Result {
	if trp.log != nil {
		trp.log.Fatal(err)
	}

	if err != nil {
		return &RawResult{
			StatusCode:  http.StatusInternalServerError,
			ContentType: ContentTypeText,
			Body:        []byte(err.Error()),
		}
	}

	return &RawResult{
		StatusCode:  http.StatusInternalServerError,
		ContentType: ContentTypeText,
		Body:        []byte("An internal server error occurred."),
	}
}

// BadRequest returns a text response.
func (trp *TextResultProvider) BadRequest(err error) Result {
	if err != nil {
		return &RawResult{
			StatusCode:  http.StatusBadRequest,
			ContentType: ContentTypeText,
			Body:        []byte(fmt.Sprintf("Bad Request: %v", err)),
		}
	}
	return &RawResult{
		StatusCode:  http.StatusBadRequest,
		ContentType: ContentTypeText,
		Body:        []byte("Bad Request"),
	}
}

// Result returns a plaintext result.
func (trp *TextResultProvider) Result(response interface{}) Result {
	return &RawResult{
		StatusCode:  http.StatusOK,
		ContentType: ContentTypeText,
		Body:        []byte(fmt.Sprintf("%s", response)),
	}
}
