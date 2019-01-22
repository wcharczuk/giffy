package controller

import (
	"net/http"
	"testing"
	"time"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/uuid"
	"github.com/blend/go-sdk/web"

	"github.com/wcharczuk/giffy/server/model"
)

func TestChartsSeaches(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	img, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	tag, err := m.CreateTestTagForImageWithVote(todo, u.ID, img.ID, uuid.V4().String())
	assert.Nil(err)

	for x := 0; x < 30; x++ {
		for y := 0; y < x; y++ {
			err = m.Invoke(todo).Create(&model.SearchHistory{
				Source:       "slack",
				TimestampUTC: time.Now().UTC().AddDate(0, 0, -1*x),
				SearchQuery:  "test",
				DidFindMatch: true,
				ImageID:      &img.ID,
				TagID:        &tag.ID,
			})
			assert.Nil(err)
		}
	}

	app := web.New()
	app.Register(Chart{Model: &m})
	contents, meta, err := app.Mock().WithPathf("/chart/searches").BytesWithMeta()
	assert.Nil(err)
	assert.NotZero(meta.ContentLength)
	assert.Equal(http.StatusOK, meta.StatusCode)
	assert.NotEmpty(contents, string(contents))
}
