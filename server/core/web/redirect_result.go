package web

// RedirectResult is a result that should cause the browser to redirect.
type RedirectResult struct {
	RedirectURI string `json:"redirect_uri"`
}
