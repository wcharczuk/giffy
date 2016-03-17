package model

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/spiffy"
)

func TestGetVoteLogsForUserID(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := spiffy.DefaultDb().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := createTestUser(tx)
	assert.Nil(err)
	i, err := createTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := createTestTagForImage(u.ID, i.ID, "winning", tx)
	assert.Nil(err)

	err = CreateOrIncrementVote(u.ID, i.ID, tag.ID, false, tx)
	assert.Nil(err)

	logEntries, err := GetVoteLogsForUserID(u.ID, tx)
	assert.Nil(err)
	assert.NotEmpty(logEntries)
}

func TestGetUserVoteForImageAndTag(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := spiffy.DefaultDb().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := createTestUser(tx)
	assert.Nil(err)
	i, err := createTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := createTestTagForImage(u.ID, i.ID, "winning", tx)
	assert.Nil(err)

	err = CreateOrIncrementVote(u.ID, i.ID, tag.ID, false, tx)
	assert.Nil(err)

	vote, err := GetUserVoteForImageAndTag(u.ID, i.ID, tag.ID, tx)
	assert.Nil(err)
	assert.False(vote.IsZero())
}
