package web

import (
	"html/template"
	"net/http"

	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/logger"
)

const (
	// DefaultTemplateNameBadRequest is the default template name for bad request view results.
	DefaultTemplateNameBadRequest = "bad_request"

	// DefaultTemplateNameInternalError is the default template name for internal server error view results.
	DefaultTemplateNameInternalError = "error"

	// DefaultTemplateNameNotFound is the default template name for not found error view results.
	DefaultTemplateNameNotFound = "not_found"

	// DefaultTemplateNameNotAuthorized is the default template name for not authorized error view results.
	DefaultTemplateNameNotAuthorized = "not_authorized"

	// DefaultTemplateBadRequest is a basic view.
	DefaultTemplateBadRequest = `<html><head><style>body { font-family: sans-serif; text-align: center; }</style></head><body><h4>Bad Request</h4></body><pre>{{ .ViewModel }}</pre></html>`

	// DefaultTemplateInternalError is a basic view.
	DefaultTemplateInternalError = `<html><head><style>body { font-family: sans-serif; text-align: center; }</style></head><body><h4>Internal Error</h4><pre>{{ .ViewModel }}</body></html>`

	// DefaultTemplateNotAuthorized is a basic view.
	DefaultTemplateNotAuthorized = `<html><head><style>body { font-family: sans-serif; text-align: center; }</style></head><body><h4>Not Authorized</h4></body></html>`

	// DefaultTemplateNotFound is a basic view.
	DefaultTemplateNotFound = `<html><head><style>body { font-family: sans-serif; text-align: center; }</style></head><body><h4>Not Found</h4></body></html>`
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
	return DefaultTemplateNameBadRequest
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
	return DefaultTemplateNameInternalError
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
	return DefaultTemplateNameNotFound
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
	return DefaultTemplateNameNotAuthorized
}

// BadRequest returns a view result.
func (vr *ViewResultProvider) BadRequest(err error) Result {
	if vr.log != nil {
		vr.log.Warning(err)
	}

	temp := vr.views.Templates().Lookup(vr.BadRequestTemplateName())
	if temp == nil {
		temp, _ = template.New("").Parse(DefaultTemplateBadRequest)
	}

	return &ViewResult{
		StatusCode: http.StatusBadRequest,
		ViewModel:  err,
		Template:   temp,
	}
}

// InternalError returns a view result.
func (vr *ViewResultProvider) InternalError(err error) Result {
	if vr.log != nil {
		vr.log.Fatal(err)
	}

	temp := vr.views.Templates().Lookup(vr.InternalErrorTemplateName())
	if temp == nil {
		temp, _ = template.New("").Parse(DefaultTemplateInternalError)
	}

	return &ViewResult{
		StatusCode: http.StatusInternalServerError,
		ViewModel:  err,
		Template:   temp,
	}
}

// NotFound returns a view result.
func (vr *ViewResultProvider) NotFound() Result {
	temp := vr.views.Templates().Lookup(vr.NotFoundTemplateName())
	if temp == nil {
		temp, _ = template.New("").Parse(DefaultTemplateNotFound)
	}

	return &ViewResult{
		StatusCode: http.StatusNotFound,
		Template:   temp,
	}
}

// NotAuthorized returns a view result.
func (vr *ViewResultProvider) NotAuthorized() Result {
	temp := vr.views.Templates().Lookup(vr.NotAuthorizedTemplateName())
	if temp == nil {
		temp, _ = template.New("").Parse(DefaultTemplateNotAuthorized)
	}

	return &ViewResult{
		StatusCode: http.StatusForbidden,
		Template:   temp,
	}
}

// View returns a view result.
func (vr *ViewResultProvider) View(viewName string, viewModel interface{}) Result {
	temp := vr.views.Templates().Lookup(viewName)
	if temp == nil {
		return vr.InternalError(exception.NewFromErr(ErrUnsetViewTemplate).WithMessagef("template: %s", viewName))
	}
	return &ViewResult{
		StatusCode: http.StatusOK,
		ViewModel:  viewModel,
		Provider:   vr,
		Template:   temp,
	}
}
