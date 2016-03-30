package web

import (
	"html/template"

	"github.com/blendlabs/go-exception"
)

// ViewResult is a result that renders a view.
type ViewResult struct {
	StatusCode int
	ViewModel  interface{}
	Template   string

	viewCache *template.Template
}

// Render renders the template
func (vr *ViewResult) Render(ctx *RequestContext) error {
	if vr.viewCache == nil {
		return exception.New("<ViewResult>.viewCache is nil at Render")
	}
	ctx.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	ctx.Response.WriteHeader(vr.StatusCode)
	return exception.Wrap(vr.viewCache.ExecuteTemplate(ctx.Response, vr.Template, vr.ViewModel))
}
