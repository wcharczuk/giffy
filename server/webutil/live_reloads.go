package webutil

import (
	web "github.com/blend/go-sdk/web"
)

// LiveReloads determine if the app should refresh the view cache on file changes.
func LiveReloads(app *web.App) {
	if app.Views().Cached() {
		app.Logger().SyncInfof("using cached views")
	} else {
		app.Logger().SyncInfof("using live view reloads")
	}
}
