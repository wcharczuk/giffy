package web

import (
	"fmt"
	"testing"
	"time"

	assert "github.com/blendlabs/go-assert"
	util "github.com/blendlabs/go-util"
)

func TestAuthManagerReadParam(t *testing.T) {
	assert := assert.New(t)

	am := NewAuthManager()

	rc, _ := New().Mock().WithFormValue(am.CookieName(), "form").CreateCtx(nil)
	assert.Empty(am.readParam(am.CookieName(), rc))

	rc, _ = New().Mock().WithHeader(am.CookieName(), "header").CreateCtx(nil)
	assert.Empty(am.readParam(am.CookieName(), rc))

	rc, _ = New().Mock().WithCookieValue(am.CookieName(), "cookie").CreateCtx(nil)
	assert.Equal("cookie", am.readParam(am.CookieName(), rc))
}

func TestAuthManagerLogin(t *testing.T) {
	assert := assert.New(t)

	app := New()
	rc, _ := app.Mock().CreateCtx(nil)

	am := NewAuthManager()
	session, err := am.Login("1", rc)
	assert.Nil(err)

	rc2, err := app.Mock().WithCookieValue(am.CookieName(), session.SessionID).CreateCtx(nil)
	assert.Nil(err)

	session, err = am.VerifySession(rc2)
	assert.Nil(err)
	assert.NotNil(session)
	assert.Equal("1", session.UserID)
}

func TestAuthManagerLoginSecure(t *testing.T) {
	assert := assert.New(t)

	app := New()
	rc, _ := app.Mock().CreateCtx(nil)

	am := NewAuthManager()
	am.SetSecret(GenerateSHA512Key())
	session, err := am.Login("1", rc)
	assert.Nil(err)

	secureSessionID, err := EncodeSignSessionID(session.SessionID, am.Secret())
	assert.Nil(err)

	rc2, err := app.Mock().
		WithCookieValue(am.CookieName(), session.SessionID).
		WithCookieValue(am.SecureCookieName(), secureSessionID).
		CreateCtx(nil)

	assert.Nil(err)

	valid, err := am.VerifySession(rc2)
	assert.Nil(err)
	assert.NotNil(valid)
	assert.Equal("1", valid.UserID)
}

func TestAuthManagerLoginSecureEmptySecure(t *testing.T) {
	assert := assert.New(t)

	app := New()
	rc, _ := app.Mock().CreateCtx(nil)

	am := NewAuthManager()
	am.SetSecret(GenerateSHA512Key())
	session, err := am.Login("1", rc)
	assert.Nil(err)

	rc2, err := app.Mock().
		WithCookieValue(am.CookieName(), session.SessionID).
		WithCookieValue(am.SecureCookieName(), "").
		CreateCtx(nil)

	assert.Nil(err)

	valid, err := am.VerifySession(rc2)
	assert.NotNil(err)
	assert.Equal(ErrSecureSessionIDEmpty, err)
	assert.Nil(valid)
}

func TestAuthManagerLoginSecureLongSecure(t *testing.T) {
	assert := assert.New(t)

	app := New()
	rc, _ := app.Mock().CreateCtx(nil)

	am := NewAuthManager()
	am.SetSecret(GenerateSHA512Key())
	session, err := am.Login("1", rc)
	assert.Nil(err)

	rc2, err := app.Mock().
		WithCookieValue(am.CookieName(), session.SessionID).
		WithCookieValue(am.SecureCookieName(), util.String.MustSecureRandom(LenSessionID<<1)).
		CreateCtx(nil)

	assert.Nil(err)

	valid, err := am.VerifySession(rc2)
	assert.NotNil(err)
	assert.Equal(ErrSecureSessionIDTooLong, err)
	assert.Nil(valid)
}

func TestAuthManagerLoginSecureSecureNotBase64(t *testing.T) {
	assert := assert.New(t)

	app := New()
	rc, _ := app.Mock().CreateCtx(nil)

	am := NewAuthManager()
	am.SetSecret(GenerateSHA512Key())
	session, err := am.Login("1", rc)
	assert.Nil(err)

	rc2, err := app.Mock().
		WithCookieValue(am.CookieName(), session.SessionID).
		WithCookieValue(am.SecureCookieName(), util.String.Random(LenSessionID)).
		CreateCtx(nil)

	assert.Nil(err)

	valid, err := am.VerifySession(rc2)
	assert.NotNil(err)
	assert.Equal(ErrSecureSessionIDInvalid, err)
	assert.Nil(valid)
}

func TestAuthManagerLoginSecureWrongKey(t *testing.T) {
	assert := assert.New(t)

	app := New()
	rc, _ := app.Mock().CreateCtx(nil)

	am := NewAuthManager()
	am.SetSecret(GenerateSHA512Key())
	session, err := am.Login("1", rc)
	assert.Nil(err)

	secureSessionID, err := EncodeSignSessionID(session.SessionID, GenerateSHA512Key())
	assert.Nil(err)

	rc2, err := app.Mock().
		WithCookieValue(am.CookieName(), session.SessionID).
		WithCookieValue(am.SecureCookieName(), secureSessionID).
		CreateCtx(nil)

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
	rc, _ := app.Mock().CreateCtx(nil)

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

	rc2, err := app.Mock().
		WithCookieValue(am.CookieName(), session.SessionID).
		CreateCtx(nil)

	assert.Nil(err)

	valid, err := am2.VerifySession(rc2)
	assert.Nil(err)
	assert.NotNil(valid)
	assert.Equal("1", valid.UserID)
}

func TestAuthManagerLogout(t *testing.T) {
	assert := assert.New(t)

	session := &Session{
		UserID:    "test_user",
		SessionID: NewSessionID(),
	}
	auth := NewAuthManager().WithSecret(util.Crypto.MustCreateKey(32))
	assert.True(auth.UseSessionCache())
	assert.NotEmpty(auth.Secret())
	auth.SessionCache().Upsert(session)

	// first we need to ensure the upsert worked
	req, err := New().Mock().
		WithCookieValue(auth.CookieName(), session.SessionID).
		WithCookieValue(auth.SecureCookieName(), MustEncodeSignSessionID(session.SessionID, auth.Secret())).
		CreateCtx(nil)

	assert.Nil(err)

	verified, err := auth.VerifySession(req)
	assert.Nil(err)
	assert.NotNil(verified)
	assert.NotNil(auth.SessionCache().Get(session.SessionID))

	req, err = New().Mock().
		WithCookieValue(auth.CookieName(), session.SessionID).
		WithCookieValue(auth.SecureCookieName(), MustEncodeSignSessionID(session.SessionID, auth.Secret())).
		CreateCtx(nil)

	assert.Nil(err)
	assert.Nil(err)
	assert.Nil(auth.Logout(req))
	assert.Nil(auth.SessionCache().Get(session.SessionID), "after logout, the session should not be cached anymore")

	sessionCookie := ReadSetCookieByName(req.Response().Header(), auth.CookieName())
	assert.NotNil(sessionCookie)
	assert.False(sessionCookie.Expires.IsZero(), fmt.Sprintf("%#v", sessionCookie)) //"we should have expired the session cookie on logout")
	assert.Equal(auth.CookiePath(), sessionCookie.Path)
	assert.True(sessionCookie.Expires.Before(time.Now().UTC()))

	secureSessionCookie := ReadSetCookieByName(req.Response().Header(), auth.SecureCookieName())
	assert.NotNil(secureSessionCookie)
	assert.False(secureSessionCookie.Expires.IsZero(), "we should have expired the secure session cookie on logout")
	assert.Equal(auth.CookiePath(), secureSessionCookie.Path)
	assert.True(secureSessionCookie.Expires.Before(time.Now().UTC()))

	// first we need to ensure the upsert worked
	req, err = New().Mock().
		WithCookieValue(auth.CookieName(), session.SessionID).
		WithCookieValue(auth.SecureCookieName(), MustEncodeSignSessionID(session.SessionID, auth.Secret())).
		CreateCtx(nil)
	assert.Nil(err)

	verified, err = auth.VerifySession(req)
	assert.Nil(err)
	assert.Nil(verified)
}

func TestAuthManagerLogoutRemoveHandler(t *testing.T) {
	assert := assert.New(t)

	session := &Session{
		UserID:    "test_user",
		SessionID: NewSessionID(),
	}
	var didCallRemoveHandler bool
	var removedSessionID string
	var stateValue interface{}
	auth := NewAuthManager().WithRemoveHandler(func(sessionID string, state State) error {
		didCallRemoveHandler = true
		removedSessionID = sessionID
		stateValue = state["foo"]
		return nil
	})

	assert.True(auth.UseSessionCache())
	auth.SessionCache().Upsert(session)

	req, err := New().Mock().
		WithCookieValue(auth.CookieName(), session.SessionID).
		WithStateValue("foo", "bar").
		CreateCtx(nil)

	assert.Nil(err)
	assert.Nil(auth.Logout(req))
	assert.Nil(auth.SessionCache().Get(session.SessionID), "after logout, the session should not be cached anymore")
	assert.True(didCallRemoveHandler)
	assert.Equal(session.SessionID, removedSessionID)
	assert.Equal("bar", stateValue)
}

func TestAuthManagerVerifySessionInvalidSessionID(t *testing.T) {
	assert := assert.New(t)

	auth := NewAuthManager()
	rc, _ := New().Mock().WithCookie(NewBasicCookie(auth.CookieName(), "")).CreateCtx(nil)
	session, err := auth.VerifySession(rc)
	assert.Nil(session)
	assert.Equal(ErrSessionIDEmpty, err)

	rc, _ = New().Mock().
		WithCookieValue(auth.CookieName(), util.String.Random(LenSessionIDBase64+1)).
		CreateCtx(nil)

	session, err = auth.VerifySession(rc)
	assert.Nil(session)
	assert.Equal(ErrSessionIDTooLong, err)
}

func TestAuthManagerVerifySessionInvalidSecureSessionID(t *testing.T) {
	assert := assert.New(t)

	rightKey := util.Crypto.MustCreateKey(32)
	secureAuth := NewAuthManager().WithSecret(rightKey)
	rc, _ := New().Mock().
		WithCookieValue(secureAuth.CookieName(), NewSessionID()).
		WithCookieValue(secureAuth.SecureCookieName(), "").CreateCtx(nil)

	session, err := secureAuth.VerifySession(rc)
	assert.Nil(session)
	assert.Equal(ErrSecureSessionIDEmpty, err)

	rc, _ = New().Mock().
		WithCookie(NewBasicCookie(secureAuth.CookieName(), NewSessionID())).
		WithCookie(NewBasicCookie(secureAuth.SecureCookieName(), util.String.Random(LenSessionIDBase64+1))).
		CreateCtx(nil)

	session, err = secureAuth.VerifySession(rc)
	assert.Nil(session)
	assert.Equal(ErrSecureSessionIDTooLong, err)

	rc, _ = New().Mock().
		WithCookieValue(secureAuth.CookieName(), NewSessionID()).
		WithCookieValue(secureAuth.SecureCookieName(), util.String.Random(64)).
		CreateCtx(nil)
	session, err = secureAuth.VerifySession(rc)
	assert.Nil(session)
	assert.Equal(ErrSecureSessionIDInvalid, err)

	sessionID := NewSessionID()
	wrongKey := util.Crypto.MustCreateKey(32)

	signed, err := SignSessionID(sessionID, wrongKey)
	assert.Nil(err)
	invalidSecureSessionID := Base64Encode(signed)

	rc, _ = New().Mock().
		WithCookieValue(secureAuth.CookieName(), sessionID).
		WithCookieValue(secureAuth.SecureCookieName(), invalidSecureSessionID).
		CreateCtx(nil)

	session, err = secureAuth.VerifySession(rc)
	assert.Nil(session)
	assert.Equal(ErrSecureSessionIDInvalid, err)
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

	rc2, err := app.Mock().WithCookieValue(am.CookieName(), sessionID).CreateCtx(nil)
	assert.Nil(err)

	valid, err := am.VerifySession(rc2)
	assert.Nil(err)
	assert.Equal(sessionID, valid.SessionID)
	assert.Equal("1", valid.UserID)
	assert.True(didCallHandler)

	rc3, err := app.Mock().WithCookieValue(am.CookieName(), NewSessionID()).CreateCtx(nil)
	assert.Nil(err)

	invalid, err := am.VerifySession(rc3)
	assert.Nil(err)
	assert.Nil(invalid)
}

func TestAuthManagerVerifySessionCached(t *testing.T) {
	assert := assert.New(t)

	auth := NewAuthManager().WithUseSessionCache(true)
	rc, err := New().Mock().CreateCtx(nil)
	assert.Nil(err)
	session, err := auth.Login("test_user", rc)
	assert.Nil(err)
	assert.NotNil(session)

	rc, err = New().Mock().WithCookieValue(auth.CookieName(), session.SessionID).CreateCtx(nil)
	assert.Nil(err)
	cachedSession, err := auth.VerifySession(rc)
	assert.Nil(err)
	assert.NotNil(cachedSession, "session should have been logged in")
	assert.Equal(session.SessionID, cachedSession.SessionID)
}

func TestAuthManagerVerifyUpdatesSessionExpiry(t *testing.T) {
	assert := assert.New(t)

	var count int
	var didCallPersistHandler bool
	var persistedSessionID string
	var stateValue interface{}
	auth := NewAuthManager().WithUseSessionCache(true).WithRollingSessionTimeout().WithSessionTimeoutProvider(func(ctx *Ctx) *time.Time {
		count++
		return util.OptionalTime(time.Now().UTC().Add(time.Duration(count) * time.Second))
	}).WithPersistHandler(func(ctx *Ctx, session *Session, state State) error {
		didCallPersistHandler = true
		persistedSessionID = session.SessionID
		stateValue = state["foo"]
		return nil
	})

	assert.True(auth.shouldUpdateSessionExpiry())

	rc, err := New().Mock().CreateCtx(nil)
	assert.Nil(err)
	session, err := auth.Login("test_user", rc)
	assert.Nil(err)
	assert.NotNil(session)

	originalSetCookie := ReadSetCookieByName(rc.Response().Header(), auth.CookieName())
	assert.False(originalSetCookie.Expires.IsZero())

	rc, err = New().Mock().WithCookieValue(auth.CookieName(), session.SessionID).WithStateValue("foo", "bar").CreateCtx(nil)
	assert.Nil(err)
	cachedSession, err := auth.VerifySession(rc)
	assert.Nil(err)
	assert.NotNil(cachedSession, "session should have been logged in")
	assert.Equal(session.SessionID, cachedSession.SessionID)
	assert.True(didCallPersistHandler)
	assert.Equal(session.SessionID, persistedSessionID)
	assert.Equal("bar", stateValue)

	updatedSetCookie := ReadSetCookieByName(rc.Response().Header(), auth.CookieName())
	assert.False(updatedSetCookie.Expires.IsZero())
	assert.True(updatedSetCookie.Expires.After(originalSetCookie.Expires),
		fmt.Sprintf("when we verify with rolling expiry, we should move the expires time forward: %v vs. %v",
			originalSetCookie.Expires.Format(time.RFC3339Nano),
			updatedSetCookie.Expires.Format(time.RFC3339Nano),
		))
}

func TestAuthManagerVerifySessionFetched(t *testing.T) {
	assert := assert.New(t)

	sessionID := NewSessionID()
	var didCallHandler bool
	auth := NewAuthManager().WithUseSessionCache(false).WithFetchHandler(func(sessionID string, state State) (*Session, error) {
		didCallHandler = true
		return &Session{
			SessionID: sessionID,
			UserID:    "test_user",
		}, nil
	})

	rc, _ := New().Mock().WithCookie(NewBasicCookie(auth.CookieName(), sessionID)).CreateCtx(nil)
	fetchedSession, err := auth.VerifySession(rc)
	assert.Nil(err)
	assert.True(didCallHandler)
	assert.NotNil(fetchedSession, "session should have been logged in")
	assert.Equal(sessionID, fetchedSession.SessionID)
	assert.Equal("test_user", fetchedSession.UserID)
}

func TestAuthManagerVerifySessionFetchedError(t *testing.T) {
	assert := assert.New(t)

	sessionID := NewSessionID()
	var didCallFetchHandler bool
	var didCallValidateHandler bool
	auth := NewAuthManager().WithUseSessionCache(false).WithFetchHandler(func(sessionID string, state State) (*Session, error) {
		didCallFetchHandler = true
		return nil, fmt.Errorf("this is only a test")
	}).WithValidateHandler(func(session *Session, state State) error {
		didCallValidateHandler = true
		return nil
	})

	rc, _ := New().Mock().
		WithCookieValue(auth.CookieName(), sessionID).
		CreateCtx(nil)

	fetchedSession, err := auth.VerifySession(rc)
	assert.NotNil(err)
	assert.Equal("this is only a test", err.Error())
	assert.True(didCallFetchHandler)
	assert.False(didCallValidateHandler)
	assert.Nil(fetchedSession)
	assert.False(IsErrSessionInvalid(err))
}

func TestAuthManagerVerifySessionValidated(t *testing.T) {
	assert := assert.New(t)

	sessionID := NewSessionID()
	var didCallFetchHandler bool
	var didCallValidateHandler bool
	auth := NewAuthManager().WithUseSessionCache(false).WithFetchHandler(func(sessionID string, state State) (*Session, error) {
		didCallFetchHandler = true
		return &Session{
			SessionID: sessionID,
			UserID:    "test_user",
		}, nil
	}).WithValidateHandler(func(session *Session, state State) error {
		didCallValidateHandler = true
		return nil
	})

	rc, _ := New().Mock().
		WithCookieValue(auth.CookieName(), sessionID).
		CreateCtx(nil)

	fetchedSession, err := auth.VerifySession(rc)
	assert.Nil(err)
	assert.True(didCallFetchHandler)
	assert.True(didCallValidateHandler, "if we provide a handler, it should be called if we have a session in the system")
	assert.NotNil(fetchedSession, "session should have been logged in")
	assert.Equal(sessionID, fetchedSession.SessionID)
	assert.Equal("test_user", fetchedSession.UserID)
}

func TestAuthManagerVerifySessionCachedRemoved(t *testing.T) {
	assert := assert.New(t)

	var removedSessionID string
	var didCallRemoveHandler bool
	var stateValue interface{}
	auth := NewAuthManager().WithUseSessionCache(true).WithRemoveHandler(func(sessionID string, state State) error {
		didCallRemoveHandler = true
		removedSessionID = sessionID
		stateValue = state["foo"]
		return nil
	})

	sessionID := NewSessionID()
	rc, err := New().Mock().WithCookieValue(auth.CookieName(), sessionID).WithStateValue("foo", "bar").CreateCtx(nil)
	assert.Nil(err)
	cachedSession, err := auth.VerifySession(rc)
	assert.Nil(err)
	assert.Nil(cachedSession, "session should not have been logged in")
	assert.True(didCallRemoveHandler)
	assert.Equal(sessionID, removedSessionID)
	assert.Equal("bar", stateValue, "state should be passed to the remove handler")
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

func TestAuthManagerIsCookieSecure(t *testing.T) {
	assert := assert.New(t)
	sm := NewAuthManager()
	assert.False(sm.CookiesHTTPSOnly())
	sm.WithCookiesHTTPSOnly(true)
	assert.True(sm.CookiesHTTPSOnly())
	sm.WithCookiesHTTPSOnly(false)
	assert.False(sm.CookiesHTTPSOnly())
}
