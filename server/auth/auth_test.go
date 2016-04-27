package auth

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/go-web"
)

func TestLogout(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := model.CreateTestUser(tx)
	assert.Nil(err)

	sessionID, err := Login(u.ID, nil, tx)
	assert.Nil(err)
	assert.NotEmpty(sessionID)

	session, err := VerifySession(sessionID, tx)
	assert.Nil(err)
	assert.NotNil(session)
	assert.Equal(u.ID, session.UserID)
	assert.NotNil(session.User)
	assert.Equal(u.ID, session.User.ID)

	err = Logout(u.ID, sessionID, nil, tx)
	assert.Nil(err)

	session, err = VerifySession(sessionID, tx)
	assert.Nil(err)
	assert.Nil(session)
}

func TestVerifySession(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := model.CreateTestUser(tx)
	assert.Nil(err)

	sessionID, err := Login(u.ID, nil, tx)
	assert.Nil(err)
	assert.NotEmpty(sessionID)

	session, err := VerifySession(sessionID, tx)
	assert.Nil(err)
	assert.NotNil(session)
	assert.Equal(u.ID, session.UserID)
	assert.NotNil(session.User)
	assert.Equal(u.ID, session.User.ID)
}

func TestSessionAware(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := model.CreateTestUser(tx)
	assert.Nil(err)

	sessionID, err := Login(u.ID, nil, tx)
	assert.Nil(err)
	assert.NotEmpty(sessionID)

	didRun := false
	hasSession := false
	app := web.New()
	app.IsolateTo(tx)
	app.GET("/", func(r *web.RequestContext) web.ControllerResult {
		didRun = true
		session := GetSession(r)
		hasSession = session != nil && session.UserID == u.ID
		return r.Raw([]byte("ok!"))
	}, SessionAware, web.InjectAPIProvider)

	err = app.Mock().WithPathf("/").WithHeader(SessionParamName, sessionID).Execute()
	assert.Nil(err)
	assert.True(didRun)
	assert.True(hasSession)
}

func TestSessionAwareInvalid(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := model.CreateTestUser(tx)
	assert.Nil(err)

	sessionID, err := Login(u.ID, nil, tx)
	assert.Nil(err)
	assert.NotEmpty(sessionID)

	didRun := false
	hasSession := false
	app := web.New()
	app.IsolateTo(tx)
	app.GET("/", func(r *web.RequestContext) web.ControllerResult {
		didRun = true
		session := GetSession(r)
		hasSession = session != nil && session.UserID == u.ID
		return r.Raw([]byte("ok!"))
	}, SessionAware, web.InjectAPIProvider)

	err = app.Mock().WithPathf("/").WithHeader(SessionParamName, "not_"+sessionID).Execute()
	assert.Nil(err)
	assert.True(didRun)
	assert.False(hasSession)
}

func TestSessionRequired(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := model.CreateTestUser(tx)
	assert.Nil(err)

	sessionID, err := Login(u.ID, nil, tx)
	assert.Nil(err)
	assert.NotEmpty(sessionID)

	didRun := false
	hasSession := false
	app := web.New()
	app.IsolateTo(tx)
	app.GET("/", func(r *web.RequestContext) web.ControllerResult {
		didRun = true
		session := GetSession(r)
		hasSession = session != nil && session.UserID == u.ID
		return r.Raw([]byte("ok!"))
	}, SessionRequired, web.InjectAPIProvider)

	err = app.Mock().WithPathf("/").WithHeader(SessionParamName, sessionID).Execute()
	assert.Nil(err)
	assert.True(didRun)
	assert.True(hasSession)
}

func TestSessionRequiredInvalid(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := model.CreateTestUser(tx)
	assert.Nil(err)

	sessionID, err := Login(u.ID, nil, tx)
	assert.Nil(err)
	assert.NotEmpty(sessionID)

	didRun := false
	hasSession := false
	app := web.New()
	app.IsolateTo(tx)
	app.GET("/", func(r *web.RequestContext) web.ControllerResult {
		didRun = true
		session := GetSession(r)
		hasSession = session != nil && session.UserID == u.ID
		return r.Raw([]byte("ok!"))
	}, SessionRequired, web.InjectAPIProvider)

	err = app.Mock().WithPathf("/").WithHeader(SessionParamName, "not_"+sessionID).Execute()
	assert.Nil(err)
	assert.False(didRun)
	assert.False(hasSession)
}
