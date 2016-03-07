package model

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/spiffy"
)

func TestVote(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := spiffy.DefaultDb().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	votingUser, votingUserErr := createTestUser(tx)
	assert.Nil(votingUserErr)

	u, err := createTestUser(tx)
	assert.Nil(err)
	i, err := createTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := createTestTag(u.ID, i.ID, "winning", tx)
	assert.Nil(err)

	voteErr := Vote(votingUser.ID, i.ID, tag.ID, false, tx)
	assert.Nil(voteErr)

	voteRecord, voteRecordErr := GetImageTagVote(i.ID, tag.ID, tx)
	assert.Nil(voteRecordErr)
	assert.NotNil(voteRecord)
	assert.Zero(voteRecord.VotesTotal)
}
