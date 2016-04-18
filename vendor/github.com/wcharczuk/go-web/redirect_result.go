package web

import "net/http"

// RedirectResult is a result that should cause the browser to redirect.
type RedirectResult struct {
	RedirectURI string `json:"redirect_uri"`
}

// Render writes the result to the response.
func (rr *RedirectResult) Render(w http.ResponseWriter, r *http.Request) error {
	http.Redirect(w, r, rr.RedirectURI, http.StatusTemporaryRedirect)
	return nil
}
