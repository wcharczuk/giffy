package controller

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/go-web"
)

func TestSlack(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := spiffy.DefaultDb().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := model.CreateTestUser(tx)
	assert.Nil(err)

	i, err := model.CreateTestImage(u.ID, tx)
	assert.Nil(err)

	_, err = model.CreateTestTagForImageWithVote(u.ID, i.ID, "__test", tx)
	assert.Nil(err)

	app := web.New()

	var res slackResponse

	app.IsolateTo(tx)
	app.Register(Integrations{})
	err = app.Mock().WithPathf("/integrations/slack").
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
