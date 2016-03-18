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
