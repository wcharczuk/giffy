package webutil

import (
	"time"

	web "github.com/blend/go-sdk/web"
	util "github.com/blendlabs/go-util"
)

// SecureCookies sets if we should issue secure cookies or not.
func SecureCookies(app *web.App) {
	app.Auth().SetSessionTimeoutProvider(func(rc *web.Ctx) *time.Time {
		return util.OptionalTime(time.Now().UTC().AddDate(0, 0, 7))
	})

	if app.Auth().CookiesHTTPSOnly() {
		app.Logger().SyncInfof("using secure cookies")
	}
}
