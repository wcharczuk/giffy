package webutil

import (
	web "github.com/blendlabs/go-web"
)

const (
	// EnvVarBaseURL is the env var for the base url.
	EnvVarBaseURL = "BASE_URL"
)

// BaseURL sets the base url / domain for the app from the environment.
func BaseURL(app *web.App) {
	app.Logger().SyncInfof("using base url: %v", app.BaseURL())
}
