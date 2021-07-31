package model

import (
	"context"
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/testutil"
	"github.com/blend/go-sdk/uuid"
)

func TestGetSiteStats(t *testing.T) {
	assert := assert.New(t)
	todo := context.TODO()
	tx, err := testutil.DefaultDB().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := NewTestManager(tx)

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
