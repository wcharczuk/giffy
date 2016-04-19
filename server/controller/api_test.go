package controller

import (
	"database/sql"
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core/auth"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/go-web"
)

const (
	TestUserUUID = "a68aac8196e444d4a3e570192a20f369"
)

type testUserResponse struct {
	Meta     *web.APIResponseMeta `json:"meta"`
	Response *model.User          `json:"response"`
}

type testUsersResponse struct {
	Meta     *web.APIResponseMeta `json:"meta"`
	Response []model.User         `json:"response"`
}

type testImagesResponse struct {
	Meta     *web.APIResponseMeta `json:"meta"`
	Response []model.Image        `json:"response"`
}

func authUser(a *assert.Assertions, tx *sql.Tx, mockUserProvider func(*sql.Tx) (*auth.Session, error)) *auth.Session {
	session, err := mockUserProvider(tx)
	a.Nil(err)
	a.NotNil(session)
	return session
}

func TestAPIUsers(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	session := authUser(assert, tx, MockAdminLogin)
	defer auth.Logout(session.UserID, session.SessionID, nil, tx)

	app := web.New()
	app.IsolateTo(tx)
	app.Register(API{})

	var res testUsersResponse
	err = app.Mock().WithHeader(auth.SessionParamName, session.SessionID).WithPathf("/api/users").JSON(&res)
	assert.Nil(err)
	assert.NotEmpty(res.Response)
}

func TestAPIUsersNonAdmin(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	session := authUser(assert, tx, MockModeratorLogin)
	defer auth.Logout(session.UserID, session.SessionID, nil, tx)

	app := web.New()
	app.IsolateTo(tx)
	app.Register(API{})

	var res testUsersResponse
	err = app.Mock().WithHeader(auth.SessionParamName, session.SessionID).WithPathf("/api/users").JSON(&res)
	assert.Nil(err)
	assert.Empty(res.Response)
}

func TestAPIUserSearch(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	session := authUser(assert, tx, MockAdminLogin)
	defer auth.Logout(session.UserID, session.SessionID, nil, tx)

	app := web.New()
	app.IsolateTo(tx)
	app.Register(API{})

	var res testUsersResponse
	err = app.Mock().
		WithHeader(auth.SessionParamName, session.SessionID).
		WithPathf("/api/users.search").
		WithQueryString("query", "will").
		JSON(&res)
	assert.Nil(err)
	assert.NotEmpty(res.Response)
}

func TestAPIUserSearchNonAdmin(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	session := authUser(assert, tx, MockModeratorLogin)
	defer auth.Logout(session.UserID, session.SessionID, nil, tx)

	app := web.New()
	app.IsolateTo(tx)
	app.Register(API{})

	var res testUsersResponse
	err = app.Mock().
		WithHeader(auth.SessionParamName, session.SessionID).
		WithPathf("/api/users.search").
		WithQueryString("query", "will").
		JSON(&res)
	assert.Nil(err)
	assert.Empty(res.Response)
}

func TestAPIUser(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	app := web.New()
	app.IsolateTo(tx)
	app.Register(API{})

	var res testUserResponse
	err = app.Mock().WithPathf("/api/user/%s", TestUserUUID).JSON(&res)
	assert.Nil(err)
	assert.NotNil(res)
	assert.NotNil(res.Response)
	assert.Equal(TestUserUUID, res.Response.UUID)
}

func TestAPIImages(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	app := web.New()
	app.IsolateTo(tx)
	app.Register(API{})

	var res testImagesResponse
	err = app.Mock().WithPathf("/api/images").JSON(&res)
	assert.Nil(err)
	assert.NotNil(res)
	assert.NotEmpty(res.Response)
}

func TestAPIImagesRandom(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	app := web.New()
	app.IsolateTo(tx)
	app.Register(API{})

	var res testImagesResponse
	err = app.Mock().WithPathf("/api/images/random/10").JSON(&res)
	assert.Nil(err)
	assert.NotNil(res)
	assert.NotEmpty(res.Response)
}
