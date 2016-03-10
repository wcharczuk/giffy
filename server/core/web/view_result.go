package web

import (
	"html/template"

	"github.com/blendlabs/go-exception"
)

var viewCache *template.Template

// InitViewCache caches templates by path.
func InitViewCache(paths ...string) error {
	views, err := template.ParseFiles(paths...)
	if err != nil {
		LogError(err)
		return err
	}
	viewCache = template.Must(views, nil)
	return nil
}

// ViewResult is a result that renders a view.
type ViewResult struct {
	StatusCode int
	ViewModel  interface{}
	Template   string
}

// Render renders the template
func (vr *ViewResult) Render(ctx *HTTPContext) error {
	ctx.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	ctx.Response.WriteHeader(vr.StatusCode)
	viewErr := exception.Wrap(viewCache.ExecuteTemplate(ctx.Response, vr.Template, vr.ViewModel))
	return viewErr
}
