package webutil

import (
	"log"
	"net/url"

	"github.com/blendlabs/go-util/env"
	web "github.com/blendlabs/go-web"
)

const (
	// EnvVarBaseURL is the env var for the base url.
	EnvVarBaseURL = "BASE_URL"
)

// BaseURL sets the base url / domain for the app from the environment.
func BaseURL(app *web.App) {
	baseURL := env.Env().String(EnvVarBaseURL)
	if len(baseURL) > 0 {
		base, err := url.Parse(baseURL)
		if err != nil {
			log.Fatal(err)
		}
		app.SetDomain(base.Host)
		app.Logger().Infof("using domain root: %s", base.Host)
	}
}
