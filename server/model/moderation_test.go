package model

import (
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/util"
)

func TestModerationCreate(t *testing.T) {
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

	mod := NewModeration(u.ID, ModerationVerbCreate, ModerationObjectImage, i.UUID, util.StringEmpty)
	err = m.Invoke(todo).Create(mod)
	assert.Nil(err)

	var verify Moderation
	err = m.Invoke(todo).Get(&verify, mod.UUID)
	assert.Nil(err)
	assert.False(verify.IsZero())
}

func TestGetModerationsForUser(t *testing.T) {
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

	for x := 0; x < 10; x++ {
		mod := NewModeration(u.ID, ModerationVerbCreate, ModerationObjectImage, i.UUID, util.StringEmpty)
		err = m.Invoke(todo).Create(mod)
		assert.Nil(err)
	}

	moderationLog, err := m.GetModerationForUserID(todo, u.ID)
	assert.Nil(err)
	firstEntry := moderationLog[0]
	assert.NotNil(firstEntry.Moderator)
	assert.False(firstEntry.Moderator.IsZero())

	assert.NotNil(firstEntry.Image)
	assert.False(firstEntry.Image.IsZero())
}

func TestGetModerationLogByCountAndOffset(t *testing.T) {
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

	for x := 0; x < 10; x++ {
		mod := NewModeration(u.ID, ModerationVerbCreate, ModerationObjectImage, i.UUID, util.StringEmpty)
		err = m.Invoke(todo).Create(mod)
		assert.Nil(err)
	}

	moderationLog, err := m.GetModerationLogByCountAndOffset(todo, 5, 0)
	assert.Nil(err)
	assert.Len(moderationLog, 5)

	firstEntry := moderationLog[0]
	assert.NotNil(firstEntry.Moderator)
	assert.False(firstEntry.Moderator.IsZero())
	assert.NotNil(firstEntry.Image)
	assert.False(firstEntry.Image.IsZero())
}
