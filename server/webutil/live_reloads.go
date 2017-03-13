package webutil

import (
	"github.com/blendlabs/go-util/env"
	web "github.com/blendlabs/go-web"
)

// LiveReloads determine if the app should refresh the view cache on file changes.
func LiveReloads(app *web.App) {
	app.ViewCache().SetEnabled(env.Env().Bool("LIVE_RELOAD"))
	if app.ViewCache().Enabled() {
		app.Logger().Infof("using live view reloads")
	} else {
		app.Logger().Infof("using cached views")
	}
}
