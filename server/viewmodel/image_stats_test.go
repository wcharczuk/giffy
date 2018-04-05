package viewmodel

import (
	"testing"

	assert "github.com/blend/go-sdk/assert"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/model"
)

func TestGetImageStats(t *testing.T) {
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

	_, err = model.CreateTestSearchHistory("slack", "foo", &i.ID, nil, tx)
	assert.Nil(err)
	_, err = model.CreateTestSearchHistory("slack", "foo", &i.ID, nil, tx)
	assert.Nil(err)
	_, err = model.CreateTestSearchHistory("slack", "foo", &i.ID, nil, tx)
	assert.Nil(err)
	_, err = model.CreateTestSearchHistory("slack", "foo", &i.ID, nil, tx)
	assert.Nil(err)

	imageStats, err := GetImageStats(i.ID, tx)
	assert.Nil(err)
	assert.Equal(imageStats.ImageID, i.ID)
	assert.Equal(2, imageStats.VotesTotal)
	assert.Equal(4, imageStats.Searches)
}
