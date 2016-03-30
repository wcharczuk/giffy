package web

import "net/http"

// RedirectResult is a result that should cause the browser to redirect.
type RedirectResult struct {
	RedirectURI string `json:"redirect_uri"`
}

// Render writes the result to the response.
func (rr *RedirectResult) Render(ctx *RequestContext) error {
	http.Redirect(ctx.Response, ctx.Request, rr.RedirectURI, http.StatusTemporaryRedirect)
	return nil
}
