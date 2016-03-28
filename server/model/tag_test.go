package model

import (
	"testing"
	"time"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/spiffy"
)

func TestGetAllTags(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
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
	tx, err := spiffy.DefaultDb().Begin()
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
	tx, err := spiffy.DefaultDb().Begin()
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
	tx, err := spiffy.DefaultDb().Begin()
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
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	i1, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	t1, err := CreateTestTagForImage(u.ID, i1.ID, "__test_tag1", tx)
	assert.Nil(err)

	i2, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	t2, err := CreateTestTagForImage(u.ID, i2.ID, "__test_tag2", tx)
	assert.Nil(err)

	newLink := NewVoteSummary(i2.ID, t1.ID, u.ID, time.Now().UTC())
	newLink.VotesFor = 1
	newLink.VotesTotal = 1
	err = spiffy.DefaultDb().CreateInTransaction(newLink, tx)

	mergeErr := MergeTags(t1.ID, t2.ID, tx)
	assert.Nil(mergeErr)

	verify, err := GetVoteSummary(i2.ID, t2.ID, tx)
	assert.Nil(err)
	assert.Equal(2, verify.VotesTotal)
}

func TestDeleteTagByID(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	tag, err := CreateTestTag(u.ID, "test", tx)
	assert.Nil(err)

	err = DeleteTagByID(tag.ID, tx)

	verify, err := GetTagByID(tag.ID, tx)
	assert.Nil(err)
	assert.True(verify.IsZero())
}
