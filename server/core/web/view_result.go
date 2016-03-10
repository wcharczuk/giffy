package web

import "net/http"

type ViewResult struct {
	StatusCode   int
	ViewModel    interface{}
	TemplateName string
}

func (vr *ViewResult) Render(w http.ResponseWriter) {

}
