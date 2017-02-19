package webutil

import (
	"log"
	"net/url"
	"time"

	util "github.com/blendlabs/go-util"
	"github.com/blendlabs/go-util/env"
	web "github.com/blendlabs/go-web"
)

const (
	// SessionParamName is a param name.
	SessionParamName = "id"
)

// SecureCookies sets if we should issue secure cookies or not.
func SecureCookies(app *web.App) {
	app.Auth().SetSessionParamName(SessionParamName)
	app.Auth().SetCookieTimeout(func(rc *web.Ctx) *time.Time {
		return util.OptionalTime(time.Now().UTC().AddDate(0, 0, 7))
	})

	baseURL := env.Env().String(EnvVarBaseURL)
	if len(baseURL) > 0 {
		base, err := url.Parse(baseURL)
		if err != nil {
			log.Fatal(err)
		}
		if util.String.CaseInsensitiveEquals(base.Scheme, "https") {
			app.Auth().SetCookieAsSecure(true)
			app.Diagnostics().Infof("using secure cookies")
		}
	}
}
