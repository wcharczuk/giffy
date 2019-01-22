package controller

import (
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/uuid"
	"github.com/blend/go-sdk/web"

	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/model"
)

func TestSlack(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	i, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, "__test")
	assert.Nil(err)

	app := web.New()

	var res slackMessage

	app.WithLogger(logger.None())
	app.Register(Integrations{Model: &m, Config: config.NewFromEnv()})

	err = app.Mock().WithVerb("POST").WithPathf("/integrations/slack").
		WithQueryString("team_id", uuid.V4().String()).
		WithQueryString("channel_id", uuid.V4().String()).
		WithQueryString("user_id", uuid.V4().String()).
		WithQueryString("team_doman", "test_domain").
		WithQueryString("channel_name", "test_channel").
		WithQueryString("user_name", "test_user").
		WithQueryString("text", "__test").
		JSON(&res)

	assert.Nil(err)
	assert.NotNil(res)
	assert.NotEmpty(res.Attachments)
}

func TestSlackErrorsWithShortQuery(t *testing.T) {
	assert := assert.New(t)
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.Manager{DB: db.Default(), Tx: tx}

	app := web.New()
	app.Register(Integrations{Model: &m, Config: config.NewFromEnv()})
	res, err := app.Mock().WithVerb("POST").WithPathf("/integrations/slack").
		WithQueryString("team_id", uuid.V4().String()).
		WithQueryString("channel_id", uuid.V4().String()).
		WithQueryString("user_id", uuid.V4().String()).
		WithQueryString("team_doman", "test_domain").
		WithQueryString("channel_name", "test_channel").
		WithQueryString("user_name", "test_user").
		WithQueryString("text", "do").Bytes()

	assert.Nil(err)
	assert.NotNil(res)
	assert.Equal(slackErrorInvalidQuery, string(res))
}
