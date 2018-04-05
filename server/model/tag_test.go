package model

import (
	"testing"

	"github.com/blend/go-sdk/assert"
)

func TestGetAllTags(t *testing.T) {
	assert := assert.New(t)
	tx, err := DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	_, err = CreateTestTag(u.ID, "test", tx)
	assert.Nil(err)

	all, err := GetAllTags(tx)
	assert.Nil(err)
	assert.NotEmpty(all)
}

func TestGetTagByID(t *testing.T) {
	assert := assert.New(t)
	tx, err := DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	tag, err := CreateTestTag(u.ID, "test", tx)
	assert.Nil(err)

	verify, err := GetTagByID(tag.ID, tx)
	assert.Nil(err)
	assert.False(verify.IsZero())
}

func TestGetTagByUUID(t *testing.T) {
	assert := assert.New(t)
	tx, err := DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	tag, err := CreateTestTag(u.ID, "test", tx)
	assert.Nil(err)

	verify, err := GetTagByUUID(tag.UUID, tx)
	assert.Nil(err)
	assert.False(verify.IsZero())
}

func TestGetTagByValue(t *testing.T) {
	assert := assert.New(t)
	tx, err := DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	_, err = CreateTestTag(u.ID, "test", tx)
	assert.Nil(err)

	verify, err := GetTagByValue("test", tx)
	assert.Nil(err)
	assert.False(verify.IsZero())
	assert.Equal("test", verify.TagValue)
}

func TestMergeTags(t *testing.T) {
	assert := assert.New(t)
	tx, err := DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u1, err := CreateTestUser(tx)
	assert.Nil(err)
	u2, err := CreateTestUser(tx)
	assert.Nil(err)

	i1, err := CreateTestImage(u1.ID, tx)
	assert.Nil(err)
	t1, err := CreateTestTagForImageWithVote(u1.ID, i1.ID, "__test_tag1", tx)
	assert.Nil(err)

	i2, err := CreateTestImage(u2.ID, tx)
	assert.Nil(err)
	t2, err := CreateTestTagForImageWithVote(u2.ID, i2.ID, "__test_tag2", tx)
	assert.Nil(err)

	mergeErr := MergeTags(t1.ID, t2.ID, tx)
	assert.Nil(mergeErr)

	verify, err := GetVoteSummary(i1.ID, t2.ID, tx)
	assert.Nil(err)
	assert.False(verify.IsZero())
	assert.Equal(1, verify.VotesTotal)

	verify, err = GetVoteSummary(i2.ID, t2.ID, tx)
	assert.Nil(err)
	assert.False(verify.IsZero())
	assert.Equal(1, verify.VotesTotal)
}

func TestMergeTagsWithExisting(t *testing.T) {
	assert := assert.New(t)
	tx, err := DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u1, err := CreateTestUser(tx)
	assert.Nil(err)
	u2, err := CreateTestUser(tx)
	assert.Nil(err)

	i1, err := CreateTestImage(u1.ID, tx)
	assert.Nil(err)
	t1, err := CreateTestTagForImageWithVote(u1.ID, i1.ID, "__test_tag1", tx)
	assert.Nil(err)

	i2, err := CreateTestImage(u2.ID, tx)
	assert.Nil(err)
	t2, err := CreateTestTagForImageWithVote(u2.ID, i2.ID, "__test_tag2", tx)
	assert.Nil(err)

	// user 1 adds tag 1 to image 2
	_, err = CreatTestVoteSummaryWithVote(i2.ID, t1.ID, u1.ID, 1, 0, tx)
	assert.Nil(err)

	mergeErr := MergeTags(t1.ID, t2.ID, tx)
	assert.Nil(mergeErr)

	verify, err := GetVoteSummary(i2.ID, t2.ID, tx)
	assert.Nil(err)
	assert.Equal(2, verify.VotesTotal)
}

func TestDeleteTagAndVotesByID(t *testing.T) {
	assert := assert.New(t)
	tx, err := DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	tag, err := CreateTestTagForImageWithVote(u.ID, i.ID, "__test_tag_value", tx)
	assert.Nil(err)

	err = DeleteTagAndVotesByID(tag.ID, tx)

	verify, err := GetTagByID(tag.ID, tx)
	assert.Nil(err)
	assert.True(verify.IsZero())
}

func TestSetTagValue(t *testing.T) {
	assert := assert.New(t)
	tx, err := DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	tag, err := CreateTestTag(u.ID, "__test_tag_value", tx)
	assert.Nil(err)

	err = SetTagValue(tag.ID, "__test_tag_value_2", tx)
	assert.Nil(err)

	verify, err := GetTagByValue("__test_tag_value", tx)
	assert.Nil(err)
	assert.True(verify.IsZero())

	verify, err = GetTagByValue("__test_tag_value_2", tx)
	assert.Nil(err)
	assert.False(verify.IsZero())
}
