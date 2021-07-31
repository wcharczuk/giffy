package model

import (
	"context"
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/testutil"
)

func TestSetVoteSummaryVoteCounts(t *testing.T) {
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
	tag, err := m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, "__test_winning")
	assert.Nil(err)

	err = m.SetVoteSummaryVoteCounts(todo, i.ID, tag.ID, 101, 100)
	assert.Nil(err)

	itv, err := m.GetVoteSummary(todo, i.ID, tag.ID)
	assert.Nil(err)
	assert.False(itv.IsZero())

	assert.Equal(101, itv.VotesFor)
	assert.Equal(100, itv.VotesAgainst)
	assert.Equal(1, itv.VotesTotal)
}

func TestSetVoteSummaryTagID(t *testing.T) {
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
	tag, err := m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, "__test_winning")
	assert.Nil(err)

	tag2, err := m.CreateTestTag(todo, u.ID, "__test_winning_two")
	assert.Nil(err)

	err = m.SetVoteSummaryTagID(todo, i.ID, tag.ID, tag2.ID)
	assert.Nil(err)

	votesForOldTag, err := m.GetVoteSummariesForTag(todo, tag.ID)
	assert.Nil(err)
	assert.Empty(votesForOldTag)

	votesForNewTag, err := m.GetVoteSummariesForTag(todo, tag2.ID)
	assert.Nil(err)
	assert.NotEmpty(votesForNewTag)
}

func TestCreateOrUpdateVote(t *testing.T) {
	assert := assert.New(t)
	todo := context.TODO()
	tx, err := testutil.DefaultDB().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := NewTestManager(tx)

	votingUser, votingUserErr := m.CreateTestUser(todo)
	assert.Nil(votingUserErr)

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)
	i, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)
	tag, err := m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, "__test_winning")
	assert.Nil(err)

	_, voteErr := m.CreateOrUpdateVote(todo, votingUser.ID, i.ID, tag.ID, false)
	assert.Nil(voteErr)

	voteRecord, voteRecordErr := m.GetVoteSummary(todo, i.ID, tag.ID)
	assert.Nil(voteRecordErr)
	assert.NotNil(voteRecord)
	assert.Zero(voteRecord.VotesTotal)
}

func TestGetImagesForTagID(t *testing.T) {
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
	tag, err := m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, "__test_winning")
	assert.Nil(err)

	imagesForTag, err := m.GetImagesForTagID(todo, tag.ID)
	assert.Nil(err)
	assert.NotEmpty(imagesForTag)
}

func TestGetTagsForImageID(t *testing.T) {
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
	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, "__test_winning")
	assert.Nil(err)

	tagsForImage, err := m.GetTagsForImageID(todo, i.ID)
	assert.Nil(err)
	assert.NotEmpty(tagsForImage)
}

func TestGetSummariesForImage(t *testing.T) {
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
	tag, err := m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, "__test_winning")
	assert.Nil(err)

	err = m.SetVoteSummaryVoteCounts(todo, i.ID, tag.ID, 101, 100)
	assert.Nil(err)

	summaries, err := m.GetVoteSummariesForImage(todo, i.ID)
	assert.Nil(err)
	assert.NotEmpty(summaries)
}

func TestGetSummariesForTag(t *testing.T) {
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
	tag, err := m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, "__test_winning")
	assert.Nil(err)

	err = m.SetVoteSummaryVoteCounts(todo, i.ID, tag.ID, 101, 100)
	assert.Nil(err)

	summaries, err := m.GetVoteSummariesForTag(todo, tag.ID)
	assert.Nil(err)
	assert.NotEmpty(summaries)
}

func TestReconcileVoteSummaryTotals(t *testing.T) {
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
	tag, err := m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, "__test_winning")
	assert.Nil(err)

	u2, err := m.CreateTestUser(todo)
	assert.Nil(err)

	_, err = m.CreateOrUpdateVote(todo, u2.ID, i.ID, tag.ID, true)
	assert.Nil(err)

	verify, err := m.GetVoteSummary(todo, i.ID, tag.ID)
	assert.Nil(err)
	assert.False(verify.IsZero())

	err = m.ReconcileVoteSummaryTotals(todo, i.ID, tag.ID)
	assert.Nil(err)

	verify2, err := m.GetVoteSummary(todo, i.ID, tag.ID)
	assert.Nil(err)
	assert.False(verify2.IsZero())

	assert.Equal(verify.VotesFor, verify2.VotesFor)
	assert.Equal(verify.VotesAgainst, verify2.VotesAgainst)
	assert.Equal(verify.VotesTotal, verify2.VotesTotal)

	err = m.SetVoteSummaryVoteCounts(todo, i.ID, tag.ID, 0, 0)
	assert.Nil(err)

	verify3, err := m.GetVoteSummary(todo, i.ID, tag.ID)
	assert.Nil(err)
	assert.False(verify3.IsZero())

	assert.Zero(verify3.VotesFor)
	assert.Zero(verify3.VotesAgainst)
	assert.Zero(verify3.VotesTotal)

	err = m.ReconcileVoteSummaryTotals(todo, i.ID, tag.ID)
	assert.Nil(err)

	verify2, err = m.GetVoteSummary(todo, i.ID, tag.ID)
	assert.Nil(err)
	assert.False(verify2.IsZero())

	assert.Equal(verify.VotesFor, verify2.VotesFor)
	assert.Equal(verify.VotesAgainst, verify2.VotesAgainst)
	assert.Equal(verify.VotesTotal, verify2.VotesTotal)
}
