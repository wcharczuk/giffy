package web

import (
	"net/http"

	"github.com/blendlabs/go-exception"
)

// ViewModel is a wrapping viewmodel.
type ViewModel struct {
	RequestContext *RequestContext
	Template       string
	ViewModel      interface{}
}

// ViewResult is a result that renders a view.
type ViewResult struct {
	StatusCode int
	ViewModel  interface{}
	Template   string

	viewCache *ViewCache
}

// Render renders the result to the given response writer.
func (vr *ViewResult) Render(rc *RequestContext) error {
	if vr.viewCache == nil {
		err := exception.New("<ViewResult>.viewCache is nil at Render()")
		http.Error(rc.Response, err.Error(), http.StatusInternalServerError)
		return err
	}

	if vr.viewCache.Templates() == nil {
		err := exception.New("<ViewResult>.viewCache.Templates is nil at Render()")
		http.Error(rc.Response, err.Error(), http.StatusInternalServerError)
		return exception.New("<ViewResult>.viewCache.Templates is nil at Render()")
	}

	rc.Response.Header().Set(HeaderContentType, ContentTypeHTML)
	rc.Response.WriteHeader(vr.StatusCode)
	err := vr.viewCache.Templates().ExecuteTemplate(rc.Response, vr.Template, &ViewModel{
		RequestContext: rc,
		Template:       vr.Template,
		ViewModel:      vr.ViewModel,
	})
	if err != nil {
		return vr.viewCache.Templates().ExecuteTemplate(rc.Response, DefaultTemplateInternalServerError, &ViewModel{
			RequestContext: rc,
			Template:       DefaultTemplateInternalServerError,
			ViewModel:      err,
		})
	}
	return err
}
