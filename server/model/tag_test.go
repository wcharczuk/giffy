package model

import (
	"testing"

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
