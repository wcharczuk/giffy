package web

import "net/http"

// NewViewResultProvider creates a new ViewResults object.
func NewViewResultProvider() *ViewResultProvider {
	return &ViewResultProvider{}
}

// ViewResultProvider returns results based on views.
type ViewResultProvider struct{}

// View returns a view result.
func (vr *ViewResultProvider) View(viewName string, viewModel interface{}) ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusOK,
		ViewModel:  viewModel,
		Template:   viewName,
	}
}

// BadRequest returns a view result.
func (vr *ViewResultProvider) BadRequest(message string) ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusBadRequest,
		ViewModel:  message,
		Template:   "bad_request",
	}
}

// InternalError returns a view result.
func (vr *ViewResultProvider) InternalError(err error) ControllerResult {
	LogErrorf("%v", err)
	return &ViewResult{
		StatusCode: http.StatusInternalServerError,
		ViewModel:  err,
		Template:   "error",
	}
}

// NotFound returns a view result.
func (vr *ViewResultProvider) NotFound() ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusNotFound,
		ViewModel:  nil,
		Template:   "not_found",
	}
}

// NotAuthorized returns a view result.
func (vr *ViewResultProvider) NotAuthorized() ControllerResult {
	return &ViewResult{
		StatusCode: http.StatusUnauthorized,
		ViewModel:  nil,
		Template:   "not_authorized",
	}
}
