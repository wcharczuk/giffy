package web

import (
	"html/template"
	"net/http"

	"github.com/blendlabs/go-exception"
)

// ViewResult is a result that renders a view.
type ViewResult struct {
	StatusCode int
	ViewModel  interface{}
	Template   string

	viewCache *template.Template
}

// Render renders the result to the given response writer.
func (vr *ViewResult) Render(w http.ResponseWriter, r *http.Request) error {
	if vr.viewCache == nil {
		return exception.New("<ViewResult>.viewCache is nil at Render")
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(vr.StatusCode)
	return exception.Wrap(vr.viewCache.ExecuteTemplate(w, vr.Template, vr.ViewModel))
}
