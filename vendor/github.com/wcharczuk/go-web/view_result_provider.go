package web

import "net/http"

// NewViewResultProvider creates a new ViewResults object.
func NewViewResultProvider(app *App, r *RequestContext) *ViewResultProvider {
	return &ViewResultProvider{app: app, requestContext: r}
}

// ViewResultProvider returns results based on views.
type ViewResultProvider struct {
	app            *App
	requestContext *RequestContext
}

// BadRequest returns a view result.
func (vr *ViewResultProvider) BadRequest(message string) ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusBadRequest,
		ViewModel:  message,
		Template:   "bad_request",
		viewCache:  vr.app.viewCache,
	}
}

// InternalError returns a view result.
func (vr *ViewResultProvider) InternalError(err error) ControllerResult {
	if vr.app != nil {
		vr.app.OnRequestError(vr.requestContext, err)
	}

	return &ViewResult{
		StatusCode: http.StatusInternalServerError,
		ViewModel:  err,
		Template:   "error",
		viewCache:  vr.app.viewCache,
	}
}

// NotFound returns a view result.
func (vr *ViewResultProvider) NotFound() ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusNotFound,
		ViewModel:  nil,
		Template:   "not_found",
		viewCache:  vr.app.viewCache,
	}
}

// NotAuthorized returns a view result.
func (vr *ViewResultProvider) NotAuthorized() ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusForbidden,
		ViewModel:  nil,
		Template:   "not_authorized",
		viewCache:  vr.app.viewCache,
	}
}

// View returns a view result.
func (vr *ViewResultProvider) View(viewName string, viewModel interface{}) ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusOK,
		ViewModel:  viewModel,
		Template:   viewName,
		viewCache:  vr.app.viewCache,
	}
}
