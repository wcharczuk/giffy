package model

import (
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/util"
)

func TestModerationCreate(t *testing.T) {
	assert := assert.New(t)
	tx, err := DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	m := NewModeration(u.ID, ModerationVerbCreate, ModerationObjectImage, i.UUID, util.StringEmpty)
	err = DB().CreateInTx(m, tx)
	assert.Nil(err)

	var verify Moderation
	err = DB().InTx(tx).Get(&verify, m.UUID)
	assert.Nil(err)
	assert.False(verify.IsZero())
}

func TestGetModerationsForUser(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	for x := 0; x < 10; x++ {
		m := NewModeration(u.ID, ModerationVerbCreate, ModerationObjectImage, i.UUID, util.StringEmpty)
		err = DB().CreateInTx(m, tx)
		assert.Nil(err)
	}

	moderationLog, err := GetModerationForUserID(u.ID, tx)
	assert.Nil(err)
	firstEntry := moderationLog[0]
	assert.NotNil(firstEntry.Moderator)
	assert.False(firstEntry.Moderator.IsZero())

	assert.NotNil(firstEntry.Image)
	assert.False(firstEntry.Image.IsZero())
}

func TestGetModerationLogByCountAndOffset(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	for x := 0; x < 10; x++ {
		m := NewModeration(u.ID, ModerationVerbCreate, ModerationObjectImage, i.UUID, util.StringEmpty)
		err = DB().CreateInTx(m, tx)
		assert.Nil(err)
	}

	moderationLog, err := GetModerationLogByCountAndOffset(5, 0, tx)
	assert.Nil(err)
	assert.Len(5, moderationLog)

	firstEntry := moderationLog[0]
	assert.NotNil(firstEntry.Moderator)
	assert.False(firstEntry.Moderator.IsZero())

	assert.NotNil(firstEntry.Image)
	assert.False(firstEntry.Image.IsZero())
}
