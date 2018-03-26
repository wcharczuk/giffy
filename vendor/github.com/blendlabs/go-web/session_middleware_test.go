package web

import (
	"net/http"
	"testing"

	assert "github.com/blendlabs/go-assert"
	util "github.com/blendlabs/go-util"
)

func TestSessionAware(t *testing.T) {
	assert := assert.New(t)

	sessionID := util.String.MustSecureRandom(64)

	var didExecuteHandler bool
	var sessionWasSet bool

	app := New()
	app.GET("/", func(r *Ctx) Result {
		didExecuteHandler = true
		sessionWasSet = r.Session() != nil
		return r.Text().Result("COOL")
	}, SessionAware)

	app.Auth().SessionCache().Upsert(&Session{
		UserID:    util.String.Random(10),
		SessionID: sessionID,
	})

	meta, err := app.Mock().WithPathf("/").WithCookieValue(app.Auth().CookieName(), sessionID).ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusOK, meta.StatusCode)
	assert.Equal(ContentTypeText, meta.Headers.Get(HeaderContentType))
	assert.True(didExecuteHandler, "we should have triggered the hander")
	assert.True(sessionWasSet, "the session should have been set by the middleware")

	unsetMeta, err := app.Mock().WithPathf("/").ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusOK, unsetMeta.StatusCode)
	assert.False(sessionWasSet)
}

func TestSessionRequired(t *testing.T) {
	assert := assert.New(t)

	sessionID := util.String.MustSecureRandom(64)

	var sessionWasSet bool
	app := New()

	app.GET("/", func(r *Ctx) Result {
		sessionWasSet = r.Session() != nil
		return r.Text().Result("COOL")
	}, SessionRequired)

	app.Auth().SessionCache().Upsert(&Session{
		UserID:    util.String.Random(10),
		SessionID: sessionID,
	})

	unsetMeta, err := app.Mock().WithPathf("/").ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusForbidden, unsetMeta.StatusCode)
	assert.False(sessionWasSet)

	meta, err := app.Mock().WithPathf("/").WithCookieValue(app.Auth().CookieName(), sessionID).ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusOK, meta.StatusCode)
	assert.True(sessionWasSet)
}

func TestSessionRequiredCustomParamName(t *testing.T) {
	assert := assert.New(t)

	sessionID := util.String.MustSecureRandom(64)

	var sessionWasSet bool
	app := New()
	app.Auth().SetCookieName("web_auth")

	app.GET("/", func(r *Ctx) Result {
		sessionWasSet = r.Session() != nil
		return r.Text().Result("COOL")
	}, SessionRequired)

	app.Auth().SessionCache().Upsert(&Session{
		UserID:    util.String.Random(10),
		SessionID: sessionID,
	})

	unsetMeta, err := app.Mock().WithPathf("/").ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusForbidden, unsetMeta.StatusCode)
	assert.False(sessionWasSet)

	meta, err := app.Mock().WithPathf("/").WithCookieValue(app.Auth().CookieName(), sessionID).ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusOK, meta.StatusCode)
	assert.True(sessionWasSet)

	meta, err = app.Mock().WithPathf("/").WithCookieValue(DefaultCookieName, sessionID).ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusForbidden, meta.StatusCode)
	assert.True(sessionWasSet)
}
