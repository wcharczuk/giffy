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
	log                   *logger.Logger
	badRequestTemplate    string
	internalErrorTemplate string
	notFoundTemplate      string
	notAuthorizedTemplate string
	views                 *ViewCache
}

// WithBadRequestTemplate sets the bad request template.
func (vr *ViewResultProvider) WithBadRequestTemplate(template string) *ViewResultProvider {
	vr.badRequestTemplate = template
	return vr
}

// BadRequestTemplate returns the bad request template.
func (vr *ViewResultProvider) BadRequestTemplate() string {
	if len(vr.badRequestTemplate) > 0 {
		return vr.badRequestTemplate
	}
	return DefaultTemplateBadRequest
}

// WithInternalErrorTemplate sets the bad request template.
func (vr *ViewResultProvider) WithInternalErrorTemplate(template string) *ViewResultProvider {
	vr.internalErrorTemplate = template
	return vr
}

// InternalErrorTemplate returns the bad request template.
func (vr *ViewResultProvider) InternalErrorTemplate() string {
	if len(vr.internalErrorTemplate) > 0 {
		return vr.internalErrorTemplate
	}
	return DefaultTemplateInternalError
}

// WithNotFoundTemplate sets the bad request template.
func (vr *ViewResultProvider) WithNotFoundTemplate(template string) *ViewResultProvider {
	vr.notFoundTemplate = template
	return vr
}

// NotFoundTemplate returns the bad request template.
func (vr *ViewResultProvider) NotFoundTemplate() string {
	if len(vr.notFoundTemplate) > 0 {
		return vr.notFoundTemplate
	}
	return DefaultTemplateNotFound
}

// WithNotAuthorizedTemplate sets the bad request template.
func (vr *ViewResultProvider) WithNotAuthorizedTemplate(template string) *ViewResultProvider {
	vr.notAuthorizedTemplate = template
	return vr
}

// NotAuthorizedTemplate returns the bad request template.
func (vr *ViewResultProvider) NotAuthorizedTemplate() string {
	if len(vr.notAuthorizedTemplate) > 0 {
		return vr.notAuthorizedTemplate
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
		Template:   vr.views.Templates().Lookup(vr.BadRequestTemplate()),
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
		Template:   vr.views.Templates().Lookup(vr.InternalErrorTemplate()),
	}
}

// NotFound returns a view result.
func (vr *ViewResultProvider) NotFound() Result {
	return &ViewResult{
		StatusCode: http.StatusNotFound,
		ViewModel:  nil,
		Template:   vr.views.Templates().Lookup(vr.NotFoundTemplate()),
	}
}

// NotAuthorized returns a view result.
func (vr *ViewResultProvider) NotAuthorized() Result {
	return &ViewResult{
		StatusCode: http.StatusForbidden,
		ViewModel:  nil,
		Template:   vr.views.Templates().Lookup(vr.NotAuthorizedTemplate()),
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
