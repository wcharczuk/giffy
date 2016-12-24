package web

import "net/http"

// RedirectResult is a result that should cause the browser to redirect.
type RedirectResult struct {
	Method      string `json:"redirect_method"`
	RedirectURI string `json:"redirect_uri"`
}

// Render writes the result to the response.
func (rr *RedirectResult) Render(rc *RequestContext) error {
	if len(rr.Method) > 0 {
		rc.Request.Method = rr.Method
		http.Redirect(rc.Response, rc.Request, rr.RedirectURI, http.StatusFound)
	} else {
		http.Redirect(rc.Response, rc.Request, rr.RedirectURI, http.StatusTemporaryRedirect)
	}

	return nil
}
