package webutil

import (
	"github.com/blendlabs/go-util/env"
	web "github.com/blendlabs/go-web"
)

// LiveReloads determine if the app should refresh the view cache on file changes.
func LiveReloads(app *web.App) {
	app.View().SetLiveReload(env.Env().Bool("LIVE_RELOAD"))
	if app.View().LiveReload() {
		app.Diagnostics().Infof("using live view reloads")
	} else {
		app.Diagnostics().Infof("using cached views")
	}
}
