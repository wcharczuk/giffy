package model

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/spiffy"
)

func TestSetTagVotes(t *testing.T) {
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

	err = SetVoteCount(i.ID, tag.ID, 101, 100, tx)
	assert.Nil(err)

	itv, err := GetVoteSummary(i.ID, tag.ID, tx)
	assert.Nil(err)
	assert.False(itv.IsZero())

	assert.Equal(101, itv.VotesFor)
	assert.Equal(100, itv.VotesAgainst)
	assert.Equal(1, itv.VotesTotal)
}

func TestCreateOrIncrementVote(t *testing.T) {
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
	tag, err := createTestTagForImage(u.ID, i.ID, "winning", tx)
	assert.Nil(err)

	voteErr := CreateOrIncrementVote(votingUser.ID, i.ID, tag.ID, false, tx)
	assert.Nil(voteErr)

	voteRecord, voteRecordErr := GetVoteSummary(i.ID, tag.ID, tx)
	assert.Nil(voteRecordErr)
	assert.NotNil(voteRecord)
	assert.Zero(voteRecord.VotesTotal)
}

func TestGetImagesForTagID(t *testing.T) {
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

	imagesForTag, err := GetImagesForTagID(tag.ID, tx)
	assert.Nil(err)
	assert.NotEmpty(imagesForTag)
}

func TestGetTagsForImageID(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := spiffy.DefaultDb().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := createTestUser(tx)
	assert.Nil(err)
	i, err := createTestImage(u.ID, tx)
	assert.Nil(err)
	_, err = createTestTagForImage(u.ID, i.ID, "winning", tx)
	assert.Nil(err)

	tagsForImage, err := GetTagsForImageID(i.ID, tx)
	assert.Nil(err)
	assert.NotEmpty(tagsForImage)
}
