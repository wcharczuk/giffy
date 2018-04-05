package controller

import (
	"net/http"
	"testing"
	"time"

	assert "github.com/blend/go-sdk/assert"
	util "github.com/blend/go-sdk/util"
	"github.com/blend/go-sdk/uuid"
	web "github.com/blend/go-sdk/web"
	"github.com/wcharczuk/giffy/server/model"
)

func TestChartsSeaches(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := model.CreateTestUser(tx)
	assert.Nil(err)

	img, err := model.CreateTestImage(u.ID, tx)
	assert.Nil(err)

	tag, err := model.CreateTestTagForImageWithVote(u.ID, img.ID, uuid.V4().String(), tx)
	assert.Nil(err)

	for x := 0; x < 30; x++ {
		for y := 0; y < x; y++ {
			err = model.DB().CreateInTx(&model.SearchHistory{
				Source:       "slack",
				TimestampUTC: time.Now().UTC().AddDate(0, 0, -1*x),
				SearchQuery:  "test",
				DidFindMatch: true,
				ImageID:      util.OptionalInt64(img.ID),
				TagID:        util.OptionalInt64(tag.ID),
			}, tx)
			assert.Nil(err)
		}
	}

	app := web.New()
	app.Register(Chart{})
	contents, meta, err := app.Mock().WithTx(tx).WithPathf("/chart/searches").BytesWithMeta()
	assert.Nil(err)
	assert.NotZero(meta.ContentLength)
	assert.Equal(http.StatusOK, meta.StatusCode)
	assert.NotEmpty(contents, string(contents))
}
