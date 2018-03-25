package web

import (
	"bytes"
	"html/template"

	exception "github.com/blendlabs/go-exception"
)

const (
	// ErrUnsetViewTemplate is a common error.
	ErrUnsetViewTemplate Error = "view result template is unset"
)

// ViewResult is a result that renders a view.
type ViewResult struct {
	StatusCode int
	ViewModel  interface{}
	Template   *template.Template
}

// Render renders the result to the given response writer.
func (vr *ViewResult) Render(ctx *Ctx) (err error) {
	if vr.Template == nil {
		err = exception.New(ErrUnsetViewTemplate)
		return
	}
	ctx.Response().Header().Set(HeaderContentType, ContentTypeHTML)
	buffer := bytes.NewBuffer([]byte{})
	err = vr.Template.Execute(buffer, &ViewModel{
		Ctx:       ctx,
		ViewModel: vr.ViewModel,
	})
	if err != nil {
		err = exception.Wrap(err)
		return
	}

	ctx.Response().WriteHeader(vr.StatusCode)
	_, err = ctx.Response().Write(buffer.Bytes())
	if err != nil && ctx != nil && ctx.Logger() != nil {
		ctx.Logger().Error(err)
	}
	return nil
}
