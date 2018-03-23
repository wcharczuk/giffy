package web

import (
	"testing"
	"time"

	assert "github.com/blendlabs/go-assert"
	util "github.com/blendlabs/go-util"
)

func TestAuthManagerLogin(t *testing.T) {
	assert := assert.New(t)

	app := New()
	rc, _ := app.Mock().Ctx(nil)

	am := NewAuthManager()
	session, err := am.Login("1", rc)
	assert.Nil(err)

	rc2, err := app.Mock().WithHeader(am.CookieName(), session.SessionID).Ctx(nil)
	assert.Nil(err)

	session, err = am.VerifySession(rc2)
	assert.Nil(err)
	assert.NotNil(session)
	assert.Equal("1", session.UserID)
}

func TestAuthManagerLoginSecure(t *testing.T) {
	assert := assert.New(t)

	app := New()
	rc, _ := app.Mock().Ctx(nil)

	am := NewAuthManager()
	am.SetSecret(GenerateSHA512Key())
	session, err := am.Login("1", rc)
	assert.Nil(err)

	secureSessionID, err := EncodeSignSessionID(session.SessionID, am.Secret())
	assert.Nil(err)

	rc2, err := app.Mock().WithHeader(am.CookieName(), session.SessionID).WithHeader(am.SecureCookieName(), secureSessionID).Ctx(nil)
	assert.Nil(err)

	valid, err := am.VerifySession(rc2)
	assert.Nil(err)
	assert.NotNil(valid)
	assert.Equal("1", valid.UserID)
}

func TestAuthManagerLoginSecureEmptySecure(t *testing.T) {
	assert := assert.New(t)

	app := New()
	rc, _ := app.Mock().Ctx(nil)

	am := NewAuthManager()
	am.SetSecret(GenerateSHA512Key())
	session, err := am.Login("1", rc)
	assert.Nil(err)

	rc2, err := app.Mock().WithHeader(am.CookieName(), session.SessionID).WithHeader(am.SecureCookieName(), "").Ctx(nil)
	assert.Nil(err)

	valid, err := am.VerifySession(rc2)
	assert.NotNil(err)
	assert.Equal(ErrSecureSessionIDEmpty, err)
	assert.Nil(valid)
}

func TestAuthManagerLoginSecureLongSecure(t *testing.T) {
	assert := assert.New(t)

	app := New()
	rc, _ := app.Mock().Ctx(nil)

	am := NewAuthManager()
	am.SetSecret(GenerateSHA512Key())
	session, err := am.Login("1", rc)
	assert.Nil(err)

	rc2, err := app.Mock().WithHeader(am.CookieName(), session.SessionID).WithHeader(am.SecureCookieName(), String.SecureRandom(LenSessionID<<1)).Ctx(nil)
	assert.Nil(err)

	valid, err := am.VerifySession(rc2)
	assert.NotNil(err)
	assert.Equal(ErrSecureSessionIDTooLong, err)
	assert.Nil(valid)
}

func TestAuthManagerLoginSecureSecureNotBase64(t *testing.T) {
	assert := assert.New(t)

	app := New()
	rc, _ := app.Mock().Ctx(nil)

	am := NewAuthManager()
	am.SetSecret(GenerateSHA512Key())
	session, err := am.Login("1", rc)
	assert.Nil(err)

	rc2, err := app.Mock().WithHeader(am.CookieName(), session.SessionID).WithHeader(am.SecureCookieName(), String.Random(LenSessionID)).Ctx(nil)
	assert.Nil(err)

	valid, err := am.VerifySession(rc2)
	assert.NotNil(err)
	assert.Equal(ErrSecureSessionIDInvalid, err)
	assert.Nil(valid)
}

func TestAuthManagerLoginSecureWrongKey(t *testing.T) {
	assert := assert.New(t)

	app := New()
	rc, _ := app.Mock().Ctx(nil)

	am := NewAuthManager()
	am.SetSecret(GenerateSHA512Key())
	session, err := am.Login("1", rc)
	assert.Nil(err)

	secureSessionID, err := EncodeSignSessionID(session.SessionID, GenerateSHA512Key())
	assert.Nil(err)

	rc2, err := app.Mock().WithHeader(am.CookieName(), session.SessionID).WithHeader(am.SecureCookieName(), secureSessionID).Ctx(nil)
	assert.Nil(err)

	valid, err := am.VerifySession(rc2)
	assert.NotNil(err)
	assert.Equal(ErrSecureSessionIDInvalid, err)
	assert.Nil(valid)
}

func TestAuthManagerLoginWithPersist(t *testing.T) {
	assert := assert.New(t)

	sessions := map[string]*Session{}

	app := New()
	rc, _ := app.Mock().Ctx(nil)

	didCallPersist := false
	am := NewAuthManager()
	am.SetPersistHandler(func(c *Ctx, s *Session, state State) error {
		didCallPersist = true
		sessions[s.SessionID] = s
		return nil
	})

	session, err := am.Login("1", rc)
	assert.Nil(err)
	assert.True(didCallPersist)

	am2 := NewAuthManager()
	am2.SetFetchHandler(func(sid string, state State) (*Session, error) {
		return sessions[sid], nil
	})

	rc2, err := app.Mock().WithHeader(am.CookieName(), session.SessionID).Ctx(nil)
	assert.Nil(err)

	valid, err := am2.VerifySession(rc2)
	assert.Nil(err)
	assert.NotNil(valid)
	assert.Equal("1", valid.UserID)
}

func TestAuthManagerVerifySessionWithFetch(t *testing.T) {
	assert := assert.New(t)

	app := New()

	sessions := map[string]*Session{}

	didCallHandler := false

	am := NewAuthManager()
	am.SetFetchHandler(func(sessionID string, state State) (*Session, error) {
		didCallHandler = true
		return sessions[sessionID], nil
	})
	sessionID := NewSessionID()
	sessions[sessionID] = NewSession("1", sessionID)

	rc2, err := app.Mock().WithHeader(am.CookieName(), sessionID).Ctx(nil)
	assert.Nil(err)

	valid, err := am.VerifySession(rc2)
	assert.Nil(err)
	assert.Equal(sessionID, valid.SessionID)
	assert.Equal("1", valid.UserID)
	assert.True(didCallHandler)

	rc3, err := app.Mock().WithHeader(am.CookieName(), NewSessionID()).Ctx(nil)
	assert.Nil(err)

	invalid, err := am.VerifySession(rc3)
	assert.Nil(err)
	assert.Nil(invalid)
}

func TestAuthManagerIsCookieSecure(t *testing.T) {
	assert := assert.New(t)
	sm := NewAuthManager()
	assert.False(sm.IsCookieHTTPSOnly())
	sm.SetCookieHTTPSOnly(true)
	assert.True(sm.IsCookieHTTPSOnly())
	sm.SetCookieHTTPSOnly(false)
	assert.False(sm.IsCookieHTTPSOnly())
}

func TestAuthManagerGenerateSessionTimeout(t *testing.T) {
	assert := assert.New(t)

	unset := NewAuthManager()
	assert.Nil(unset.GenerateSessionTimeout(nil))

	absolute := NewAuthManager()
	absolute.SetSessionTimeout(24 * time.Hour)
	expiresAt := absolute.GenerateSessionTimeout(nil)
	assert.NotNil(expiresAt)
	assert.InTimeDelta(*expiresAt, time.Now().UTC().Add(24*time.Hour), time.Minute)

	provided := NewAuthManager()
	provided.SetSessionTimeoutProvider(func(_ *Ctx) *time.Time {
		return util.OptionalTime(time.Now().UTC().Add(6 * time.Hour))
	})
	expiresAt = provided.GenerateSessionTimeout(nil)
	assert.NotNil(expiresAt)
	assert.InTimeDelta(*expiresAt, time.Now().UTC().Add(6*time.Hour), time.Minute)
}

func TestAuthManagerNilSessionRegression(t *testing.T) {
	assert := assert.New(t)

	app := New()

	auth := NewAuthManager()

	var didCall bool
	auth.SetRemoveHandler(func(ssid string, state State) error {
		didCall = true
		return nil
	})

	rc, _ := app.Mock().WithHeader(auth.CookieName(), NewSessionID()).Ctx(nil)
	auth.VerifySession(rc)
	assert.True(didCall)
}
