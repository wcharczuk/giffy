package webutil

import (
	web "github.com/blendlabs/go-web"
)

// LiveReloads determine if the app should refresh the view cache on file changes.
func LiveReloads(app *web.App) {
	if app.Views().Cached() {
		app.Logger().Infof("using cached views")
	} else {
		app.Logger().Infof("using live view reloads")
	}
}
