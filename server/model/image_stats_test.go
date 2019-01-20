package model

import (
	"testing"

	assert "github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/uuid"
)

func TestGetImageStats(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)
	i, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)
	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, uuid.V4().String())
	assert.Nil(err)
	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, uuid.V4().String())
	assert.Nil(err)

	_, err = m.CreateTestSearchHistory(todo, "slack", "foo", &i.ID, nil)
	assert.Nil(err)
	_, err = m.CreateTestSearchHistory(todo, "slack", "foo", &i.ID, nil)
	assert.Nil(err)
	_, err = m.CreateTestSearchHistory(todo, "slack", "foo", &i.ID, nil)
	assert.Nil(err)
	_, err = m.CreateTestSearchHistory(todo, "slack", "foo", &i.ID, nil)
	assert.Nil(err)

	imageStats, err := m.GetImageStats(todo, i.ID)
	assert.Nil(err)
	assert.Equal(imageStats.ImageID, i.ID)
	assert.Equal(2, imageStats.VotesTotal)
	assert.Equal(4, imageStats.Searches)
}
