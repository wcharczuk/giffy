package controller

import (
	"net/http"
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/r2"
	"github.com/blend/go-sdk/testutil"
	"github.com/blend/go-sdk/uuid"
	"github.com/blend/go-sdk/web"

	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/model"
)

func TestSlack(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := testutil.DefaultDB().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.NewTestManager(tx)

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	i, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, "__test")
	assert.Nil(err)

	app := web.MustNew()
	app.Log = logger.None()
	app.Register(Integrations{Model: &m, Config: config.MustNewFromEnv()})

	var res slackMessage
	_, err = web.MockMethod(app, http.MethodPost, "/integrations/slack",
		r2.OptQueryValue("team_id", uuid.V4().String()),
		r2.OptQueryValue("channel_id", uuid.V4().String()),
		r2.OptQueryValue("user_id", uuid.V4().String()),
		r2.OptQueryValue("team_doman", "test_domain"),
		r2.OptQueryValue("channel_name", "test_channel"),
		r2.OptQueryValue("user_name", "test_user"),
		r2.OptQueryValue("text", "__test"),
	).JSON(&res)

	assert.Nil(err)
	assert.NotNil(res)
	assert.NotEmpty(res.Attachments)
}

func TestSlackErrorsWithShortQuery(t *testing.T) {
	assert := assert.New(t)
	tx, err := testutil.DefaultDB().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.NewTestManager(tx)

	app := web.MustNew()
	app.Register(Integrations{Model: &m, Config: config.MustNewFromEnv()})

	contents, _, err := web.MockMethod(app, http.MethodPost, "/integrations/slack",
		r2.OptQueryValue("team_id", uuid.V4().String()),
		r2.OptQueryValue("channel_id", uuid.V4().String()),
		r2.OptQueryValue("user_id", uuid.V4().String()),
		r2.OptQueryValue("team_doman", "test_domain"),
		r2.OptQueryValue("channel_name", "test_channel"),
		r2.OptQueryValue("user_name", "test_user"),
		r2.OptQueryValue("text", "do"),
	).Bytes()

	assert.Nil(err)
	assert.NotEmpty(contents)
	assert.Equal(slackErrorInvalidQuery, string(contents))
}
