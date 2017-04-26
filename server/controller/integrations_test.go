package controller

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-web"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/model"
)

func TestSlack(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := model.DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := model.CreateTestUser(tx)
	assert.Nil(err)

	i, err := model.CreateTestImage(u.ID, tx)
	assert.Nil(err)

	_, err = model.CreateTestTagForImageWithVote(u.ID, i.ID, "__test", tx)
	assert.Nil(err)

	app := web.New()

	var res slackMessage

	app.SetLogger(logger.New(logger.NewEventFlagSetNone()))
	app.Register(Integrations{})
	err = app.Mock().WithTx(tx).WithVerb("POST").WithPathf("/integrations/slack").
		WithQueryString("team_id", core.UUIDv4().ToShortString()).
		WithQueryString("channel_id", core.UUIDv4().ToShortString()).
		WithQueryString("user_id", core.UUIDv4().ToShortString()).
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
	app := web.New()

	app.Register(Integrations{})
	res, err := app.Mock().WithVerb("POST").WithPathf("/integrations/slack").
		WithQueryString("team_id", core.UUIDv4().ToShortString()).
		WithQueryString("channel_id", core.UUIDv4().ToShortString()).
		WithQueryString("user_id", core.UUIDv4().ToShortString()).
		WithQueryString("team_doman", "test_domain").
		WithQueryString("channel_name", "test_channel").
		WithQueryString("user_name", "test_user").
		WithQueryString("text", "do").Bytes()

	assert.Nil(err)
	assert.NotNil(res)
	assert.Equal(slackErrorInvalidQuery, string(res))
}
