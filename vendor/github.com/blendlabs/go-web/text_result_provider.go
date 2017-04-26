package web

import (
	"fmt"
	"net/http"
)

// NewTextResultProvider returns a new text result provider.
func NewTextResultProvider(ctx *Ctx) *TextResultProvider {
	return &TextResultProvider{ctx: ctx}
}

// TextResultProvider is the default response provider if none is specified.
type TextResultProvider struct {
	ctx *Ctx
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
	if trp.ctx != nil {
		trp.ctx.logFatal(err)
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
func (trp *TextResultProvider) BadRequest(message string) Result {
	if len(message) > 0 {
		return &RawResult{
			StatusCode:  http.StatusBadRequest,
			ContentType: ContentTypeText,
			Body:        []byte(fmt.Sprintf("Bad Request: %s", message)),
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
