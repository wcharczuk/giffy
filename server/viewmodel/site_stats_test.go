package viewmodel

import (
	"testing"

	"github.com/blendlabs/go-assert"

	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/model"
)

func TestGetSiteStats(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := model.CreateTestUser(tx)
	assert.Nil(err)
	i, err := model.CreateTestImage(u.ID, tx)
	assert.Nil(err)
	_, err = model.CreateTestTagForImageWithVote(u.ID, i.ID, core.UUIDv4().ToShortString(), tx)
	assert.Nil(err)

	_, err = model.CreateTestTagForImageWithVote(u.ID, i.ID, core.UUIDv4().ToShortString(), tx)
	assert.Nil(err)

	stats, err := GetSiteStats(tx)
	assert.Nil(err)
	assert.NotNil(stats)
}
