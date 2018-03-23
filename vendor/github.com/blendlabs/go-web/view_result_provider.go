package web

import (
	"net/http"

	logger "github.com/blendlabs/go-logger"
)

const (
	// DefaultTemplateBadRequest is the default template name for bad request view results.
	DefaultTemplateBadRequest = "bad_request"

	// DefaultTemplateInternalError is the default template name for internal server error view results.
	DefaultTemplateInternalError = "error"

	// DefaultTemplateNotFound is the default template name for not found error view results.
	DefaultTemplateNotFound = "not_found"

	// DefaultTemplateNotAuthorized is the default template name for not authorized error view results.
	DefaultTemplateNotAuthorized = "not_authorized"
)

// NewViewResultProvider creates a new ViewResults object.
func NewViewResultProvider(log *logger.Logger, vc *ViewCache) *ViewResultProvider {
	return &ViewResultProvider{log: log, views: vc}
}

// ViewResultProvider returns results based on views.
type ViewResultProvider struct {
	log                       *logger.Logger
	badRequestTemplateName    string
	internalErrorTemplateName string
	notFoundTemplateName      string
	notAuthorizedTemplateName string
	views                     *ViewCache
}

// WithBadRequestTemplateName sets the bad request template.
func (vr *ViewResultProvider) WithBadRequestTemplateName(templateName string) *ViewResultProvider {
	vr.badRequestTemplateName = templateName
	return vr
}

// BadRequestTemplateName returns the bad request template.
func (vr *ViewResultProvider) BadRequestTemplateName() string {
	if len(vr.badRequestTemplateName) > 0 {
		return vr.badRequestTemplateName
	}
	return DefaultTemplateBadRequest
}

// WithInternalErrorTemplateName sets the bad request template.
func (vr *ViewResultProvider) WithInternalErrorTemplateName(templateName string) *ViewResultProvider {
	vr.internalErrorTemplateName = templateName
	return vr
}

// InternalErrorTemplateName returns the bad request template.
func (vr *ViewResultProvider) InternalErrorTemplateName() string {
	if len(vr.internalErrorTemplateName) > 0 {
		return vr.internalErrorTemplateName
	}
	return DefaultTemplateInternalError
}

// WithNotFoundTemplateName sets the not found request template name.
func (vr *ViewResultProvider) WithNotFoundTemplateName(templateName string) *ViewResultProvider {
	vr.notFoundTemplateName = templateName
	return vr
}

// NotFoundTemplateName returns the not found template name.
func (vr *ViewResultProvider) NotFoundTemplateName() string {
	if len(vr.notFoundTemplateName) > 0 {
		return vr.notFoundTemplateName
	}
	return DefaultTemplateNotFound
}

// WithNotAuthorizedTemplateName sets the bad request template.
func (vr *ViewResultProvider) WithNotAuthorizedTemplateName(templateName string) *ViewResultProvider {
	vr.notAuthorizedTemplateName = templateName
	return vr
}

// NotAuthorizedTemplateName returns the bad request template name.
func (vr *ViewResultProvider) NotAuthorizedTemplateName() string {
	if len(vr.notAuthorizedTemplateName) > 0 {
		return vr.notAuthorizedTemplateName
	}
	return DefaultTemplateNotAuthorized
}

// BadRequest returns a view result.
func (vr *ViewResultProvider) BadRequest(err error) Result {
	if vr.log != nil {
		vr.log.Warning(err)
	}

	return &ViewResult{
		StatusCode: http.StatusBadRequest,
		ViewModel:  err,
		Template:   vr.views.Templates().Lookup(vr.BadRequestTemplateName()),
	}
}

// InternalError returns a view result.
func (vr *ViewResultProvider) InternalError(err error) Result {
	if vr.log != nil {
		vr.log.Fatal(err)
	}

	return &ViewResult{
		StatusCode: http.StatusInternalServerError,
		ViewModel:  err,
		Template:   vr.views.Templates().Lookup(vr.InternalErrorTemplateName()),
	}
}

// NotFound returns a view result.
func (vr *ViewResultProvider) NotFound() Result {
	return &ViewResult{
		StatusCode: http.StatusNotFound,
		ViewModel:  nil,
		Template:   vr.views.Templates().Lookup(vr.NotFoundTemplateName()),
	}
}

// NotAuthorized returns a view result.
func (vr *ViewResultProvider) NotAuthorized() Result {
	return &ViewResult{
		StatusCode: http.StatusForbidden,
		ViewModel:  nil,
		Template:   vr.views.Templates().Lookup(vr.NotAuthorizedTemplateName()),
	}
}

// View returns a view result.
func (vr *ViewResultProvider) View(viewName string, viewModel interface{}) Result {
	return &ViewResult{
		StatusCode: http.StatusOK,
		ViewModel:  viewModel,
		Template:   vr.views.Templates().Lookup(viewName),
	}
}
