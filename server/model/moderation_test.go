package model

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/spiffy"
)

func TestModerationCreate(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := spiffy.DefaultDb().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := createTestUser(tx)
	assert.Nil(err)

	i, err := createTestImage(u.ID, tx)
	assert.Nil(err)

	m := NewModeration(u.ID, ModerationVerbCreate, ModerationObjectImage, i.UUID)
	err = spiffy.DefaultDb().CreateInTransaction(m, tx)
	assert.Nil(err)

	var verify Moderation
	err = spiffy.DefaultDb().GetByIDInTransaction(&verify, tx, m.UUID)
	assert.Nil(err)
	assert.False(verify.IsZero())
}

func TestGetModerationLogByCountAndOffset(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := spiffy.DefaultDb().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := createTestUser(tx)
	assert.Nil(err)

	i, err := createTestImage(u.ID, tx)
	assert.Nil(err)

	for x := 0; x < 10; x++ {
		m := NewModeration(u.ID, ModerationVerbCreate, ModerationObjectImage, i.UUID)
		err = spiffy.DefaultDb().CreateInTransaction(m, tx)
		assert.Nil(err)
	}

	moderationLog, err := GetModerationLogByCountAndOffset(5, 0, tx)
	assert.Nil(err)
	assert.Len(moderationLog, 5)
	firstEntry := moderationLog[0]
	assert.NotNil(firstEntry.User)
	assert.False(firstEntry.User.IsZero())
}