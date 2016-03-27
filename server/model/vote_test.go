package model

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/spiffy"
)

func TestGetVotesForUser(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := spiffy.DefaultDb().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)
	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := CreateTestTagForImage(u.ID, i.ID, "winning", tx)
	assert.Nil(err)

	_, err = CreateOrChangeVote(u.ID, i.ID, tag.ID, false, tx)
	assert.Nil(err)

	votes, err := GetVotesForUser(u.ID, tx)
	assert.Nil(err)
	assert.NotEmpty(votes)
}

func TestGetVotesForUserForImage(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := spiffy.DefaultDb().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)
	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := CreateTestTagForImage(u.ID, i.ID, "winning", tx)
	assert.Nil(err)

	_, err = CreateOrChangeVote(u.ID, i.ID, tag.ID, false, tx)
	assert.Nil(err)

	votes, err := GetVotesForUserForImage(u.ID, i.ID, tx)
	assert.Nil(err)
	assert.NotEmpty(votes)
}

func TestGetVotesForUserForTag(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := spiffy.DefaultDb().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)
	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := CreateTestTagForImage(u.ID, i.ID, "winning", tx)
	assert.Nil(err)

	_, err = CreateOrChangeVote(u.ID, i.ID, tag.ID, false, tx)
	assert.Nil(err)

	votes, err := GetVotesForUserForTag(u.ID, tag.ID, tx)
	assert.Nil(err)
	assert.NotEmpty(votes)
}

func TestGetVote(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := spiffy.DefaultDb().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)
	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := CreateTestTagForImage(u.ID, i.ID, "winning", tx)
	assert.Nil(err)

	_, err = CreateOrChangeVote(u.ID, i.ID, tag.ID, false, tx)
	assert.Nil(err)

	vote, err := GetVote(u.ID, i.ID, tag.ID, tx)
	assert.Nil(err)
	assert.False(vote.IsZero())
}
