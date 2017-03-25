package web

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/blendlabs/go-exception"
)

// ViewModel is a wrapping viewmodel.
type ViewModel struct {
	Ctx       *Ctx
	Template  string
	ViewModel interface{}
}

// ViewResult is a result that renders a view.
type ViewResult struct {
	StatusCode int
	ViewModel  interface{}
	Template   string

	viewCache *ViewCache
}

// Render renders the result to the given response writer.
func (vr *ViewResult) Render(ctx *Ctx) error {
	if vr.viewCache == nil {
		err := exception.New("<ViewResult>.viewCache is nil at Render()")
		http.Error(ctx.Response, err.Error(), http.StatusInternalServerError)
		return err
	}

	var viewTemplates *template.Template
	var err error

	if vr.viewCache.Enabled() {
		viewTemplates = vr.viewCache.Templates()
	} else {
		viewTemplates, err = vr.viewCache.Parse()
		if err != nil {
			http.Error(ctx.Response, err.Error(), http.StatusInternalServerError)
			return err
		}
	}
	if viewTemplates == nil {
		err := exception.New("<ViewResult>.viewCache.Templates is nil at Render()")
		http.Error(ctx.Response, err.Error(), http.StatusInternalServerError)
		return err
	}

	ctx.Response.Header().Set(HeaderContentType, ContentTypeHTML)

	buffer := bytes.NewBuffer([]byte{})
	err = viewTemplates.ExecuteTemplate(buffer, vr.Template, &ViewModel{
		Ctx:       ctx,
		Template:  vr.Template,
		ViewModel: vr.ViewModel,
	})

	if err != nil {
		buffer.Reset()

		err = vr.viewCache.Templates().ExecuteTemplate(buffer, DefaultTemplateInternalServerError, &ViewModel{
			Ctx:       ctx,
			Template:  DefaultTemplateInternalServerError,
			ViewModel: err,
		})

		ctx.Response.WriteHeader(http.StatusInternalServerError)
		ctx.Response.Write(buffer.Bytes())
		return err
	}
	ctx.Response.WriteHeader(vr.StatusCode)
	_, err = ctx.Response.Write(buffer.Bytes())
	return err
}
