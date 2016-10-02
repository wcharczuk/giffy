package model

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestSetVoteSummaryVoteCounts(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)
	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := CreateTestTagForImageWithVote(u.ID, i.ID, "__test_winning", tx)
	assert.Nil(err)

	err = SetVoteSummaryVoteCounts(i.ID, tag.ID, 101, 100, tx)
	assert.Nil(err)

	itv, err := GetVoteSummary(i.ID, tag.ID, tx)
	assert.Nil(err)
	assert.False(itv.IsZero())

	assert.Equal(101, itv.VotesFor)
	assert.Equal(100, itv.VotesAgainst)
	assert.Equal(1, itv.VotesTotal)
}

func TestSetVoteSummaryTagID(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)
	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := CreateTestTagForImageWithVote(u.ID, i.ID, "__test_winning", tx)
	assert.Nil(err)

	tag2, err := CreateTestTag(u.ID, "__test_winning_two", tx)
	assert.Nil(err)

	err = SetVoteSummaryTagID(i.ID, tag.ID, tag2.ID, tx)
	assert.Nil(err)

	votesForOldTag, err := GetVoteSummariesForTag(tag.ID, tx)
	assert.Nil(err)
	assert.Empty(votesForOldTag)

	votesForNewTag, err := GetVoteSummariesForTag(tag2.ID, tx)
	assert.Nil(err)
	assert.NotEmpty(votesForNewTag)
}

func TestCreateOrUpdateVote(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	votingUser, votingUserErr := CreateTestUser(tx)
	assert.Nil(votingUserErr)

	u, err := CreateTestUser(tx)
	assert.Nil(err)
	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := CreateTestTagForImageWithVote(u.ID, i.ID, "__test_winning", tx)
	assert.Nil(err)

	_, voteErr := CreateOrUpdateVote(votingUser.ID, i.ID, tag.ID, false, tx)
	assert.Nil(voteErr)

	voteRecord, voteRecordErr := GetVoteSummary(i.ID, tag.ID, tx)
	assert.Nil(voteRecordErr)
	assert.NotNil(voteRecord)
	assert.Zero(voteRecord.VotesTotal)
}

func TestGetImagesForTagID(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)
	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := CreateTestTagForImageWithVote(u.ID, i.ID, "__test_winning", tx)
	assert.Nil(err)

	imagesForTag, err := GetImagesForTagID(tag.ID, tx)
	assert.Nil(err)
	assert.NotEmpty(imagesForTag)
}

func TestGetTagsForImageID(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)
	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)
	_, err = CreateTestTagForImageWithVote(u.ID, i.ID, "__test_winning", tx)
	assert.Nil(err)

	tagsForImage, err := GetTagsForImageID(i.ID, tx)
	assert.Nil(err)
	assert.NotEmpty(tagsForImage)
}

func TestGetSummariesForImage(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)
	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := CreateTestTagForImageWithVote(u.ID, i.ID, "__test_winning", tx)
	assert.Nil(err)

	err = SetVoteSummaryVoteCounts(i.ID, tag.ID, 101, 100, tx)
	assert.Nil(err)

	summaries, err := GetVoteSummariesForImage(i.ID, tx)
	assert.Nil(err)
	assert.NotEmpty(summaries)
}

func TestGetSummariesForTag(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)
	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := CreateTestTagForImageWithVote(u.ID, i.ID, "__test_winning", tx)
	assert.Nil(err)

	err = SetVoteSummaryVoteCounts(i.ID, tag.ID, 101, 100, tx)
	assert.Nil(err)

	summaries, err := GetVoteSummariesForTag(tag.ID, tx)
	assert.Nil(err)
	assert.NotEmpty(summaries)
}

func TestReconcileVoteSummaryTotals(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)
	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := CreateTestTagForImageWithVote(u.ID, i.ID, "__test_winning", tx)
	assert.Nil(err)

	u2, err := CreateTestUser(tx)
	assert.Nil(err)

	_, err = CreateOrUpdateVote(u2.ID, i.ID, tag.ID, true, tx)
	assert.Nil(err)

	verify, err := GetVoteSummary(i.ID, tag.ID, tx)
	assert.Nil(err)
	assert.False(verify.IsZero())

	err = ReconcileVoteSummaryTotals(i.ID, tag.ID, tx)
	assert.Nil(err)

	verify2, err := GetVoteSummary(i.ID, tag.ID, tx)
	assert.Nil(err)
	assert.False(verify2.IsZero())

	assert.Equal(verify.VotesFor, verify2.VotesFor)
	assert.Equal(verify.VotesAgainst, verify2.VotesAgainst)
	assert.Equal(verify.VotesTotal, verify2.VotesTotal)

	err = SetVoteSummaryVoteCounts(i.ID, tag.ID, 0, 0, tx)
	assert.Nil(err)

	verify3, err := GetVoteSummary(i.ID, tag.ID, tx)
	assert.Nil(err)
	assert.False(verify3.IsZero())

	assert.Zero(verify3.VotesFor)
	assert.Zero(verify3.VotesAgainst)
	assert.Zero(verify3.VotesTotal)

	err = ReconcileVoteSummaryTotals(i.ID, tag.ID, tx)
	assert.Nil(err)

	verify2, err = GetVoteSummary(i.ID, tag.ID, tx)
	assert.Nil(err)
	assert.False(verify2.IsZero())

	assert.Equal(verify.VotesFor, verify2.VotesFor)
	assert.Equal(verify.VotesAgainst, verify2.VotesAgainst)
	assert.Equal(verify.VotesTotal, verify2.VotesTotal)
}
