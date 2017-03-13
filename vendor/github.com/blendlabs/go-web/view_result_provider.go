package web

import (
	"net/http"

	logger "github.com/blendlabs/go-logger"
)

const (
	// DefaultTemplateBadRequest is the default template name for bad request view results.
	DefaultTemplateBadRequest = "bad_request"

	// DefaultTemplateInternalServerError is the default template name for internal server error view results.
	DefaultTemplateInternalServerError = "error"

	// DefaultTemplateNotFound is the default template name for not found error view results.
	DefaultTemplateNotFound = "not_found"

	// DefaultTemplateNotAuthorized is the default template name for not authorized error view results.
	DefaultTemplateNotAuthorized = "not_authorized"
)

// NewViewResultProvider creates a new ViewResults object.
func NewViewResultProvider(log *logger.Agent, vc *ViewCache, r *Ctx) *ViewResultProvider {
	return &ViewResultProvider{diagnostics: log, viewCache: vc, ctx: r}
}

// ViewResultProvider returns results based on views.
type ViewResultProvider struct {
	diagnostics *logger.Agent
	ctx         *Ctx

	viewCache *ViewCache
}

// BadRequest returns a view result.
func (vr *ViewResultProvider) BadRequest(message string) Result {
	return &ViewResult{
		StatusCode: http.StatusBadRequest,
		ViewModel:  message,
		Template:   DefaultTemplateBadRequest,
		viewCache:  vr.viewCache,
	}
}

// InternalError returns a view result.
func (vr *ViewResultProvider) InternalError(err error) Result {
	if vr.diagnostics != nil {
		if vr.ctx != nil {
			vr.diagnostics.FatalWithReq(err, vr.ctx.Request)
		} else {
			vr.diagnostics.FatalWithReq(err, nil)
		}
	}

	return &ViewResult{
		StatusCode: http.StatusInternalServerError,
		ViewModel:  err,
		Template:   DefaultTemplateInternalServerError,
		viewCache:  vr.viewCache,
	}
}

// NotFound returns a view result.
func (vr *ViewResultProvider) NotFound() Result {
	return &ViewResult{
		StatusCode: http.StatusNotFound,
		ViewModel:  nil,
		Template:   DefaultTemplateNotFound,
		viewCache:  vr.viewCache,
	}
}

// NotAuthorized returns a view result.
func (vr *ViewResultProvider) NotAuthorized() Result {
	return &ViewResult{
		StatusCode: http.StatusForbidden,
		ViewModel:  nil,
		Template:   DefaultTemplateNotAuthorized,
		viewCache:  vr.viewCache,
	}
}

// View returns a view result.
func (vr *ViewResultProvider) View(viewName string, viewModel interface{}) Result {
	return &ViewResult{
		StatusCode: http.StatusOK,
		ViewModel:  viewModel,
		Template:   viewName,
		viewCache:  vr.viewCache,
	}
}

// Result doesnt return a view result.
func (vr *ViewResultProvider) Result(response interface{}) Result {
	panic("ViewResultProvider.Result is not implemented")
}
