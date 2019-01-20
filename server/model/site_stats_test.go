package model

import (
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/uuid"
)

func TestGetSiteStats(t *testing.T) {
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

	stats, err := m.GetSiteStats(todo)
	assert.Nil(err)
	assert.NotNil(stats)
}
