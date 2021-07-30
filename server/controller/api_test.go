package controller

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/crypto"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/oauth"
	"github.com/blend/go-sdk/uuid"
	"github.com/blend/go-sdk/web"

	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/viewmodel"
)

const (
	TestUserUUID = "a68aac8196e444d4a3e570192a20f369"
)

type testUserResponse struct {
	Meta     *APIResponseMeta `json:"meta"`
	Response *model.User      `json:"response"`
}

type testUsersResponse struct {
	Meta     *APIResponseMeta `json:"meta"`
	Response []model.User     `json:"response"`
}

type testImagesResponse struct {
	Meta     *APIResponseMeta  `json:"meta"`
	Response []viewmodel.Image `json:"response"`
}

type testCurrentUserResponse struct {
	Meta     *APIResponseMeta       `json:"meta"`
	Response *viewmodel.CurrentUser `json:"response"`
}

type testSiteStatsResponse struct {
	Meta     *APIResponseMeta `json:"meta"`
	Response *model.SiteStats `json:"response"`
}

type testTeamsResponse struct {
	Meta     *APIResponseMeta  `json:"meta"`
	Response []model.SlackTeam `json:"response"`
}

type testTeamResponse struct {
	Meta     *APIResponseMeta `json:"meta"`
	Response *model.SlackTeam `json:"response"`
}

func testCtx() context.Context {
	return context.TODO()
}

func TestAPIUsers(t *testing.T) {
	assert := assert.New(t)
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.Manager{DB: db.Default(), Tx: tx}

	auth, session := MockAuth(assert, &m, MockAdminLogin)
	defer MockLogout(assert, &m, auth, session)

	app := web.New()
	app.WithAuth(auth)
	app.Register(APIs{Model: &m, Config: config.MustNewFromEnv()})

	var res testUsersResponse
	err = app.Mock().WithCookieValue(auth.CookieName(), session.SessionID).WithPathf("/api/users").JSON(&res)
	assert.Nil(err)
	assert.NotEmpty(res.Response)
}

func TestAPIUsersNonAdmin(t *testing.T) {
	assert := assert.New(t)
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.Manager{DB: db.Default(), Tx: tx}

	auth, session := MockAuth(assert, &m, MockModeratorLogin)
	defer MockLogout(assert, &m, auth, session)

	app := web.New()
	app.WithAuth(auth)
	app.Register(APIs{Model: &m, Config: config.MustNewFromEnv()})

	var res testUsersResponse
	err = app.Mock().WithCookieValue(auth.CookieName(), session.SessionID).WithPathf("/api/users").JSON(&res)
	assert.Nil(err)
	assert.Empty(res.Response)
}

func TestAPIUserSearch(t *testing.T) {
	assert := assert.New(t)
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.Manager{DB: db.Default(), Tx: tx}

	auth, session := MockAuth(assert, &m, MockAdminLogin)
	defer MockLogout(assert, &m, auth, session)

	app := web.New()
	app.WithAuth(auth)
	app.Register(APIs{Model: &m, Config: config.MustNewFromEnv()})

	var res testUsersResponse
	err = app.Mock().
		WithCookieValue(auth.CookieName(), session.SessionID).
		WithPathf("/api/users.search").
		WithQueryString("query", "will").
		JSON(&res)

	assert.Nil(err)
	assert.NotEmpty(res.Response)
}

func TestAPIUserSearchNonAdmin(t *testing.T) {
	assert := assert.New(t)
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.Manager{DB: db.Default(), Tx: tx}

	auth, session := MockAuth(assert, &m, MockModeratorLogin)
	defer MockLogout(assert, &m, auth, session)

	app := web.New()
	app.WithAuth(auth)
	app.Register(APIs{Model: &m, Config: config.MustNewFromEnv()})

	var res testUsersResponse
	err = app.Mock().
		WithCookieValue(auth.CookieName(), session.SessionID).
		WithPathf("/api/users.search").
		WithQueryString("query", "will").
		JSON(&res)

	assert.Nil(err)
	assert.Empty(res.Response)
}

func TestAPIUser(t *testing.T) {
	assert := assert.New(t)
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.Manager{DB: db.Default(), Tx: tx}

	app := web.New()
	app.Register(APIs{Model: &m, Config: config.MustNewFromEnv()})

	var res testUserResponse
	err = app.Mock().WithPathf("/api/user/%s", TestUserUUID).JSON(&res)
	assert.Nil(err)
	assert.NotNil(res)
	assert.NotNil(res.Response)
	assert.Equal(TestUserUUID, res.Response.UUID)
}

func TestAPIImages(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.Manager{DB: db.Default(), Tx: tx}

	u, err := CreateTestModeratorUser(&m)
	assert.Nil(err)

	_, err = m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	app := web.New()
	app.Register(APIs{Model: &m, Config: config.MustNewFromEnv()})

	var res testImagesResponse
	err = app.Mock().WithPathf("/api/images").JSON(&res)
	assert.Nil(err)
	assert.NotNil(res)
	assert.NotEmpty(res.Response)
}

func TestAPIImagesRandom(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.Manager{DB: db.Default(), Tx: tx}

	u, err := CreateTestModeratorUser(&m)
	assert.Nil(err)

	_, err = m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	app := web.New()
	app.Register(APIs{Model: &m, Config: config.MustNewFromEnv()})

	var res testImagesResponse
	err = app.Mock().WithPathf("/api/images/random/10").JSON(&res)
	assert.Nil(err)
	assert.NotNil(res)
	assert.NotEmpty(res.Response)
}

func TestAPISiteStats(t *testing.T) {
	assert := assert.New(t)
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.Manager{DB: db.Default(), Tx: tx}

	app := web.New()
	app.Register(APIs{Model: &m, Config: config.MustNewFromEnv()})

	var res testSiteStatsResponse
	err = app.Mock().WithPathf("/api/stats").JSON(&res)
	assert.Nil(err)
	assert.NotNil(res.Response, "stats response is nil")
}

func TestAPISessionUser(t *testing.T) {
	assert := assert.New(t)
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.Manager{DB: db.Default(), Tx: tx}

	auth, session := MockAuth(assert, &m, MockAdminLogin)
	defer MockLogout(assert, &m, auth, session)

	app := web.New()
	app.WithAuth(auth)
	app.Register(APIs{Model: &m, Config: config.MustNewFromEnv(), OAuth: oauth.New().WithSecret(crypto.MustCreateKey(32))})

	var res testCurrentUserResponse
	err = app.Mock().WithCookieValue(auth.CookieName(), session.SessionID).WithPathf("/api/session.user").JSON(&res)
	assert.Nil(err)
	assert.NotNil(res.Response)
	assert.True(res.Response.IsLoggedIn)
	assert.NotEmpty(res.Response.UUID)
}

func TestAPISessionUserLoggedOut(t *testing.T) {
	assert := assert.New(t)
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.Manager{DB: db.Default(), Tx: tx}

	app := web.New()
	app.Register(APIs{Model: &m, Config: config.MustNewFromEnv(), OAuth: oauth.New().WithSecret(crypto.MustCreateKey(32))})

	var res testCurrentUserResponse
	err = app.Mock().WithPathf("/api/session.user").JSON(&res)
	assert.Nil(err)
	assert.NotNil(res.Response)
	assert.False(res.Response.IsLoggedIn)
	assert.Empty(res.Response.UUID)
}

func TestAPIGetTeamsNoAuth(t *testing.T) {
	assert := assert.New(t)
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.Manager{DB: db.Default(), Tx: tx}

	app := web.New()
	app.Register(APIs{Model: &m, Config: config.MustNewFromEnv()})

	var res testTeamsResponse
	err = app.Mock().WithPathf("/api/teams").JSON(&res)
	assert.Nil(err)
	assert.Equal(http.StatusForbidden, res.Meta.StatusCode)
}

func TestAPIGetTeams(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.Manager{DB: db.Default(), Tx: tx}

	auth, session, err := MockAdminLogin(&m)
	assert.Nil(err)

	team1 := &model.SlackTeam{
		CreatedUTC:          time.Now().UTC(),
		TeamID:              uuid.V4().String(),
		TeamName:            "Test Team",
		ContentRatingFilter: model.ContentRatingFilterDefault,
		CreatedByID:         uuid.V4().String(),
		CreatedByName:       "Test User",
		IsEnabled:           true,
	}
	err = m.Invoke(todo).Create(team1)
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

	err = m.Invoke(todo).Create(team2)
	assert.Nil(err)

	app := web.New()
	app.WithAuth(auth)

	app.Register(APIs{Model: &m, Config: config.MustNewFromEnv()})

	var res testTeamsResponse
	err = app.Mock().WithPathf("/api/teams").WithCookieValue(auth.CookieName(), session.SessionID).JSON(&res)
	assert.Nil(err)
	assert.Equal(http.StatusOK, res.Meta.StatusCode)
	assert.NotEmpty(res.Response)
	assert.True(len(res.Response) >= 2)
}

func TestAPIGetTeamNotAuthed(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.Manager{DB: db.Default(), Tx: tx}

	team1 := &model.SlackTeam{
		CreatedUTC:          time.Now().UTC(),
		TeamID:              uuid.V4().String(),
		TeamName:            "Test Team",
		ContentRatingFilter: model.ContentRatingFilterDefault,
		CreatedByID:         uuid.V4().String(),
		CreatedByName:       "Test User",
		IsEnabled:           true,
	}
	err = m.Invoke(todo).Create(team1)
	assert.Nil(err)

	app := web.New()
	app.Register(APIs{Model: &m, Config: config.MustNewFromEnv()})

	var res testTeamResponse
	err = app.Mock().WithPathf("/api/team/%s", team1.TeamID).JSON(&res)
	assert.Nil(err)
	assert.Equal(http.StatusForbidden, res.Meta.StatusCode)
}

func TestAPIGetTeam(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.Manager{DB: db.Default(), Tx: tx}

	auth, session, err := MockAdminLogin(&m)
	assert.Nil(err)

	team1 := &model.SlackTeam{
		CreatedUTC:          time.Now().UTC(),
		TeamID:              uuid.V4().String(),
		TeamName:            "Test Team",
		ContentRatingFilter: model.ContentRatingFilterDefault,
		CreatedByID:         uuid.V4().String(),
		CreatedByName:       "Test User",
		IsEnabled:           true,
	}
	err = m.Invoke(todo).Create(team1)
	assert.Nil(err)

	app := web.New()
	app.WithAuth(auth)
	app.Register(APIs{Model: &m, Config: config.MustNewFromEnv()})

	var res testTeamResponse
	err = app.Mock().WithPathf("/api/team/%s", team1.TeamID).WithCookieValue(auth.CookieName(), session.SessionID).JSON(&res)
	assert.Nil(err)
	assert.Equal(http.StatusOK, res.Meta.StatusCode)
	assert.NotNil(res.Response)
	assert.Equal(team1.TeamID, res.Response.TeamID)
}
