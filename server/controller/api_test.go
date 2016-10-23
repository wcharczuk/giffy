package controller

import (
	"testing"
	"time"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-util"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/auth"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/viewmodel"
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

type testCurrentUserResponse struct {
	Meta     *web.APIResponseMeta   `json:"meta"`
	Response *viewmodel.CurrentUser `json:"response"`
}

type testSiteStatsResponse struct {
	Meta     *web.APIResponseMeta `json:"meta"`
	Response *viewmodel.SiteStats `json:"response"`
}

type testTeamsResponse struct {
	Meta     *web.APIResponseMeta `json:"meta"`
	Response []model.SlackTeam    `json:"response"`
}

type testTeamResponse struct {
	Meta     *web.APIResponseMeta `json:"meta"`
	Response *model.SlackTeam     `json:"response"`
}

func TestAPIUsers(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	session := MockAuth(assert, tx, MockAdminLogin)
	defer auth.Logout(session.UserID, session.SessionID, nil, tx)

	app := web.New()
	app.IsolateTo(tx)
	app.Register(new(API))

	var res testUsersResponse
	err = app.Mock().WithHeader(auth.SessionParamName, session.SessionID).WithPathf("/api/users").FetchResponseAsJSON(&res)
	assert.Nil(err)
	assert.NotEmpty(res.Response)
}

func TestAPIUsersNonAdmin(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	session := MockAuth(assert, tx, MockModeratorLogin)
	defer auth.Logout(session.UserID, session.SessionID, nil, tx)

	app := web.New()
	app.IsolateTo(tx)
	app.Register(API{})

	var res testUsersResponse
	err = app.Mock().WithHeader(auth.SessionParamName, session.SessionID).WithPathf("/api/users").FetchResponseAsJSON(&res)
	assert.Nil(err)
	assert.Empty(res.Response)
}

func TestAPIUserSearch(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	session := MockAuth(assert, tx, MockAdminLogin)
	defer auth.Logout(session.UserID, session.SessionID, nil, tx)

	app := web.New()
	app.IsolateTo(tx)
	app.Register(new(API))

	var res testUsersResponse
	err = app.Mock().
		WithHeader(auth.SessionParamName, session.SessionID).
		WithPathf("/api/users.search").
		WithQueryString("query", "will").
		FetchResponseAsJSON(&res)
	assert.Nil(err)
	assert.NotEmpty(res.Response)
}

func TestAPIUserSearchNonAdmin(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	session := MockAuth(assert, tx, MockModeratorLogin)
	defer auth.Logout(session.UserID, session.SessionID, nil, tx)

	app := web.New()
	app.IsolateTo(tx)
	app.Register(new(API))

	var res testUsersResponse
	err = app.Mock().
		WithHeader(auth.SessionParamName, session.SessionID).
		WithPathf("/api/users.search").
		WithQueryString("query", "will").
		FetchResponseAsJSON(&res)
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
	app.Register(new(API))

	var res testUserResponse
	err = app.Mock().WithPathf("/api/user/%s", TestUserUUID).FetchResponseAsJSON(&res)
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

	u, err := CreateTestModeratorUser(tx)
	assert.Nil(err)

	_, err = model.CreateTestImage(u.ID, tx)
	assert.Nil(err)

	app := web.New()
	app.IsolateTo(tx)
	app.Register(new(API))

	var res testImagesResponse
	err = app.Mock().WithPathf("/api/images").FetchResponseAsJSON(&res)
	assert.Nil(err)
	assert.NotNil(res)
	assert.NotEmpty(res.Response)
}

func TestAPIImagesRandom(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := CreateTestModeratorUser(tx)
	assert.Nil(err)

	_, err = model.CreateTestImage(u.ID, tx)
	assert.Nil(err)

	app := web.New()
	app.IsolateTo(tx)
	app.Register(new(API))

	var res testImagesResponse
	err = app.Mock().WithPathf("/api/images/random/10").FetchResponseAsJSON(&res)
	assert.Nil(err)
	assert.NotNil(res)
	assert.NotEmpty(res.Response)
}

func TestAPISiteStats(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	app := web.New()
	app.IsolateTo(tx)
	app.Register(new(API))

	var res testSiteStatsResponse
	err = app.Mock().WithPathf("/api/stats").FetchResponseAsJSON(&res)
	assert.Nil(err)
	assert.NotNil(res.Response, "stats response is nil")
}

func TestAPISessionUser(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	session := MockAuth(assert, tx, MockAdminLogin)
	defer auth.Logout(session.UserID, session.SessionID, nil, tx)

	app := web.New()
	app.IsolateTo(tx)
	app.Register(new(API))

	var res testCurrentUserResponse
	err = app.Mock().WithHeader(auth.SessionParamName, session.SessionID).WithPathf("/api/session.user").FetchResponseAsJSON(&res)
	assert.Nil(err)
	assert.NotNil(res.Response)
	assert.True(res.Response.IsLoggedIn)
	assert.NotEmpty(res.Response.UUID)
}

func TestAPISessionUserLoggedOut(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	app := web.New()
	app.IsolateTo(tx)
	app.Register(new(API))

	var res testCurrentUserResponse
	err = app.Mock().WithPathf("/api/session.user").FetchResponseAsJSON(&res)
	assert.Nil(err)
	assert.NotNil(res.Response)
	assert.False(res.Response.IsLoggedIn)
	assert.Empty(res.Response.UUID)
}

func TestAPIGetTeams(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	team1 := &model.SlackTeam{
		CreatedUTC:          time.Now().UTC(),
		TeamID:              util.UUIDv4().ToShortString(),
		TeamName:            "Test Team",
		ContentRatingFilter: model.ContentRatingFilterDefault,
		CreatedByID:         util.UUIDv4().ToShortString(),
		CreatedByName:       "Test User",
		IsEnabled:           true,
	}
	err = model.DB().CreateInTx(team1, tx)
	assert.Nil(err)

	team2 := &model.SlackTeam{
		CreatedUTC:          time.Now().UTC(),
		TeamID:              util.UUIDv4().ToShortString(),
		TeamName:            "Test Team",
		ContentRatingFilter: model.ContentRatingFilterDefault,
		CreatedByID:         util.UUIDv4().ToShortString(),
		CreatedByName:       "Test User",
		IsEnabled:           true,
	}

	err = model.DB().CreateInTx(team2, tx)
	assert.Nil(err)

	app := web.New()
	app.IsolateTo(tx)
	app.Register(new(API))

	var res testTeamsResponse
	err = app.Mock().WithPathf("/api/teams").FetchResponseAsJSON(&res)
	assert.Nil(err)
	assert.NotEmpty(res.Response)
	assert.Len(res.Response, 2)
}
