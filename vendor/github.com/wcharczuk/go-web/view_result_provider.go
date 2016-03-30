package web

import (
	"html/template"
	"net/http"
)

// NewViewResultProvider creates a new ViewResults object.
func NewViewResultProvider(logger Logger, viewCache *template.Template) *ViewResultProvider {
	return &ViewResultProvider{logger: logger, viewCache: viewCache}
}

// ViewResultProvider returns results based on views.
type ViewResultProvider struct {
	viewCache *template.Template
	logger    Logger
}

// View returns a view result.
func (vr *ViewResultProvider) View(viewName string, viewModel interface{}) ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusOK,
		ViewModel:  viewModel,
		Template:   viewName,
		viewCache:  vr.viewCache,
	}
}

// BadRequest returns a view result.
func (vr *ViewResultProvider) BadRequest(message string) ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusBadRequest,
		ViewModel:  message,
		Template:   "bad_request",
		viewCache:  vr.viewCache,
	}
}

// InternalError returns a view result.
func (vr *ViewResultProvider) InternalError(err error) ControllerResult {
	vr.logger.Errorf("%v", err)
	return &ViewResult{
		StatusCode: http.StatusInternalServerError,
		ViewModel:  err,
		Template:   "error",
		viewCache:  vr.viewCache,
	}
}

// NotFound returns a view result.
func (vr *ViewResultProvider) NotFound() ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusNotFound,
		ViewModel:  nil,
		Template:   "not_found",
		viewCache:  vr.viewCache,
	}
}

// NotAuthorized returns a view result.
func (vr *ViewResultProvider) NotAuthorized() ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusUnauthorized,
		ViewModel:  nil,
		Template:   "not_authorized",
		viewCache:  vr.viewCache,
	}
}
