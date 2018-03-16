package webutil

import (
	"time"

	util "github.com/blendlabs/go-util"
	web "github.com/blendlabs/go-web"
)

// SecureCookies sets if we should issue secure cookies or not.
func SecureCookies(app *web.App) {
	app.Auth().SetSessionTimeoutProvider(func(rc *web.Ctx) *time.Time {
		return util.OptionalTime(time.Now().UTC().AddDate(0, 0, 7))
	})

	if app.Auth().CookieHTTPSOnly() {
		app.Logger().SyncInfof("using secure cookies")
	}
}
