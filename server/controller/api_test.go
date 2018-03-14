package controller

import (
	"net/http"
	"testing"
	"time"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-util/uuid"
	"github.com/blendlabs/go-web"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/viewmodel"
	"github.com/wcharczuk/giffy/server/webutil"
)

const (
	TestUserUUID = "a68aac8196e444d4a3e570192a20f369"
)

type testUserResponse struct {
	Meta     *webutil.APIResponseMeta `json:"meta"`
	Response *model.User              `json:"response"`
}

type testUsersResponse struct {
	Meta     *webutil.APIResponseMeta `json:"meta"`
	Response []model.User             `json:"response"`
}

type testImagesResponse struct {
	Meta     *webutil.APIResponseMeta `json:"meta"`
	Response []model.Image            `json:"response"`
}

type testCurrentUserResponse struct {
	Meta     *webutil.APIResponseMeta `json:"meta"`
	Response *viewmodel.CurrentUser   `json:"response"`
}

type testSiteStatsResponse struct {
	Meta     *webutil.APIResponseMeta `json:"meta"`
	Response *viewmodel.SiteStats     `json:"response"`
}

type testTeamsResponse struct {
	Meta     *webutil.APIResponseMeta `json:"meta"`
	Response []model.SlackTeam        `json:"response"`
}

type testTeamResponse struct {
	Meta     *webutil.APIResponseMeta `json:"meta"`
	Response *model.SlackTeam         `json:"response"`
}

func TestAPIUsers(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	auth, session := MockAuth(assert, tx, MockAdminLogin)
	defer auth.Logout(session, nil)

	app := web.New()
	app.WithAuth(auth)
	app.Register(new(API))

	var res testUsersResponse
	err = app.Mock().WithTx(tx).WithHeader(auth.CookieName(), session.SessionID).WithPathf("/api/users").JSON(&res)
	assert.Nil(err)
	assert.NotEmpty(res.Response)
}

func TestAPIUsersNonAdmin(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	auth, session := MockAuth(assert, tx, MockModeratorLogin)
	defer auth.Logout(session, nil)

	app := web.New()
	app.WithAuth(auth)
	app.Register(API{})

	var res testUsersResponse
	err = app.Mock().WithTx(tx).WithHeader(auth.CookieName(), session.SessionID).WithPathf("/api/users").JSON(&res)
	assert.Nil(err)
	assert.Empty(res.Response)
}

func TestAPIUserSearch(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	auth, session := MockAuth(assert, tx, MockAdminLogin)
	defer auth.Logout(session, nil)

	app := web.New()
	app.WithAuth(auth)
	app.Register(new(API))

	var res testUsersResponse
	err = app.Mock().WithTx(tx).
		WithHeader(auth.CookieName(), session.SessionID).
		WithPathf("/api/users.search").
		WithQueryString("query", "will").
		JSON(&res)
	assert.Nil(err)
	assert.NotEmpty(res.Response)
}

func TestAPIUserSearchNonAdmin(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	auth, session := MockAuth(assert, tx, MockModeratorLogin)
	defer auth.Logout(session, nil)

	app := web.New()
	app.WithAuth(auth)
	app.Register(new(API))

	var res testUsersResponse
	err = app.Mock().WithTx(tx).
		WithHeader(auth.CookieName(), session.SessionID).
		WithPathf("/api/users.search").
		WithQueryString("query", "will").
		JSON(&res)
	assert.Nil(err)
	assert.Empty(res.Response)
}

func TestAPIUser(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	app := web.New()
	app.Register(new(API))

	var res testUserResponse
	err = app.Mock().WithTx(tx).WithPathf("/api/user/%s", TestUserUUID).JSON(&res)
	assert.Nil(err)
	assert.NotNil(res)
	assert.NotNil(res.Response)
	assert.Equal(TestUserUUID, res.Response.UUID)
}

func TestAPIImages(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := CreateTestModeratorUser(tx)
	assert.Nil(err)

	_, err = model.CreateTestImage(u.ID, tx)
	assert.Nil(err)

	app := web.New()
	app.Register(new(API))

	var res testImagesResponse
	err = app.Mock().WithTx(tx).WithPathf("/api/images").JSON(&res)
	assert.Nil(err)
	assert.NotNil(res)
	assert.NotEmpty(res.Response)
}

func TestAPIImagesRandom(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := CreateTestModeratorUser(tx)
	assert.Nil(err)

	_, err = model.CreateTestImage(u.ID, tx)
	assert.Nil(err)

	app := web.New()
	app.Register(new(API))

	var res testImagesResponse
	err = app.Mock().WithTx(tx).WithPathf("/api/images/random/10").JSON(&res)
	assert.Nil(err)
	assert.NotNil(res)
	assert.NotEmpty(res.Response)
}

func TestAPISiteStats(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	app := web.New()
	app.Register(new(API))

	var res testSiteStatsResponse
	err = app.Mock().WithTx(tx).WithPathf("/api/stats").JSON(&res)
	assert.Nil(err)
	assert.NotNil(res.Response, "stats response is nil")
}

func TestAPISessionUser(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	auth, session := MockAuth(assert, tx, MockAdminLogin)
	defer auth.Logout(session, nil)

	app := web.New()
	app.WithAuth(auth)
	app.Register(new(API))

	var res testCurrentUserResponse
	err = app.Mock().WithTx(tx).WithHeader(auth.CookieName(), session.SessionID).WithPathf("/api/session.user").JSON(&res)
	assert.Nil(err)
	assert.NotNil(res.Response)
	assert.True(res.Response.IsLoggedIn)
	assert.NotEmpty(res.Response.UUID)
}

func TestAPISessionUserLoggedOut(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	app := web.New()
	app.Register(new(API))

	var res testCurrentUserResponse
	err = app.Mock().WithTx(tx).WithPathf("/api/session.user").JSON(&res)
	assert.Nil(err)
	assert.NotNil(res.Response)
	assert.False(res.Response.IsLoggedIn)
	assert.Empty(res.Response.UUID)
}

func TestAPIGetTeamsNoAuth(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	app := web.New()
	app.Register(new(API))

	var res testTeamsResponse
	err = app.Mock().WithTx(tx).WithPathf("/api/teams").JSON(&res)
	assert.Nil(err)
	assert.Equal(http.StatusForbidden, res.Meta.StatusCode)
}

func TestAPIGetTeams(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	auth, session, err := MockAdminLogin(tx)

	team1 := &model.SlackTeam{
		CreatedUTC:          time.Now().UTC(),
		TeamID:              uuid.V4().String(),
		TeamName:            "Test Team",
		ContentRatingFilter: model.ContentRatingFilterDefault,
		CreatedByID:         uuid.V4().String(),
		CreatedByName:       "Test User",
		IsEnabled:           true,
	}
	err = model.DB().CreateInTx(team1, tx)
	assert.Nil(err)

	team2 := &model.SlackTeam{
		CreatedUTC:          time.Now().UTC(),
		TeamID:              uuid.V4().String(),
		TeamName:            "Test Team",
		ContentRatingFilter: model.ContentRatingFilterDefault,
		CreatedByID:         uuid.V4().String(),
		CreatedByName:       "Test User",
		IsEnabled:           true,
	}

	err = model.DB().CreateInTx(team2, tx)
	assert.Nil(err)

	app := web.New()
	app.WithAuth(auth)

	app.Register(new(API))

	var res testTeamsResponse
	err = app.Mock().WithTx(tx).WithPathf("/api/teams").WithHeader(auth.CookieName(), session.SessionID).JSON(&res)
	assert.Nil(err)
	assert.Equal(http.StatusOK, res.Meta.StatusCode)
	assert.NotEmpty(res.Response)
	assert.Len(res.Response, 2)
}

func TestAPIGetTeamNotAuthed(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	team1 := &model.SlackTeam{
		CreatedUTC:          time.Now().UTC(),
		TeamID:              uuid.V4().String(),
		TeamName:            "Test Team",
		ContentRatingFilter: model.ContentRatingFilterDefault,
		CreatedByID:         uuid.V4().String(),
		CreatedByName:       "Test User",
		IsEnabled:           true,
	}
	err = model.DB().CreateInTx(team1, tx)
	assert.Nil(err)

	app := web.New()

	app.Register(new(API))

	var res testTeamResponse
	err = app.Mock().WithTx(tx).WithPathf("/api/team/%s", team1.TeamID).JSON(&res)
	assert.Nil(err)
	assert.Equal(http.StatusForbidden, res.Meta.StatusCode)
}

func TestAPIGetTeam(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	auth, session, err := MockAdminLogin(tx)

	team1 := &model.SlackTeam{
		CreatedUTC:          time.Now().UTC(),
		TeamID:              uuid.V4().String(),
		TeamName:            "Test Team",
		ContentRatingFilter: model.ContentRatingFilterDefault,
		CreatedByID:         uuid.V4().String(),
		CreatedByName:       "Test User",
		IsEnabled:           true,
	}
	err = model.DB().CreateInTx(team1, tx)
	assert.Nil(err)

	app := web.New()
	app.WithAuth(auth)

	app.Register(new(API))

	var res testTeamResponse
	err = app.Mock().WithTx(tx).WithPathf("/api/team/%s", team1.TeamID).WithHeader(auth.CookieName(), session.SessionID).JSON(&res)
	assert.Nil(err)
	assert.Equal(http.StatusOK, res.Meta.StatusCode)
	assert.NotNil(res.Response)
	assert.Equal(team1.TeamID, res.Response.TeamID)
}
