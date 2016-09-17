package web

import (
	"html/template"
	"net/http"
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
func NewViewResultProvider(app *App, r *RequestContext) *ViewResultProvider {
	return &ViewResultProvider{app: app, requestContext: r}
}

// ViewResultProvider returns results based on views.
type ViewResultProvider struct {
	app            *App
	requestContext *RequestContext
}

func (vr ViewResultProvider) viewCache() *template.Template {
	if vr.app != nil {
		return vr.app.viewCache
	}
	return nil
}

// BadRequest returns a view result.
func (vr *ViewResultProvider) BadRequest(message string) ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusBadRequest,
		ViewModel:  message,
		Template:   DefaultTemplateBadRequest,
		viewCache:  vr.viewCache(),
	}
}

// InternalError returns a view result.
func (vr *ViewResultProvider) InternalError(err error) ControllerResult {
	if vr.app != nil {
		vr.app.diagnostics.Error(err)
	}

	return &ViewResult{
		StatusCode: http.StatusInternalServerError,
		ViewModel:  err,
		Template:   DefaultTemplateInternalServerError,
		viewCache:  vr.viewCache(),
	}
}

// NotFound returns a view result.
func (vr *ViewResultProvider) NotFound() ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusNotFound,
		ViewModel:  nil,
		Template:   DefaultTemplateNotFound,
		viewCache:  vr.viewCache(),
	}
}

// NotAuthorized returns a view result.
func (vr *ViewResultProvider) NotAuthorized() ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusForbidden,
		ViewModel:  nil,
		Template:   DefaultTemplateNotAuthorized,
		viewCache:  vr.viewCache(),
	}
}

// View returns a view result.
func (vr *ViewResultProvider) View(viewName string, viewModel interface{}) ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusOK,
		ViewModel:  viewModel,
		Template:   viewName,
		viewCache:  vr.viewCache(),
	}
}
