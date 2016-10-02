package model

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/wcharczuk/giffy/server/core"
)

func TestGetVotesForUser(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)
	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := CreateTestTagForImageWithVote(u.ID, i.ID, core.UUIDv4().ToShortString(), tx)
	assert.Nil(err)

	_, err = CreateOrUpdateVote(u.ID, i.ID, tag.ID, false, tx)
	assert.Nil(err)

	votes, err := GetVotesForUser(u.ID, tx)
	assert.Nil(err)
	assert.NotEmpty(votes)
}

func TestGetVotesForImage(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)
	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := CreateTestTagForImageWithVote(u.ID, i.ID, core.UUIDv4().ToShortString(), tx)
	assert.Nil(err)

	_, err = CreateOrUpdateVote(u.ID, i.ID, tag.ID, false, tx)
	assert.Nil(err)

	votes, err := GetVotesForImage(i.ID, tx)
	assert.Nil(err)
	assert.NotEmpty(votes)
}

func TestGetVotesForTag(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)
	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := CreateTestTagForImageWithVote(u.ID, i.ID, core.UUIDv4().ToShortString(), tx)
	assert.Nil(err)

	_, err = CreateOrUpdateVote(u.ID, i.ID, tag.ID, false, tx)
	assert.Nil(err)

	votes, err := GetVotesForTag(tag.ID, tx)
	assert.Nil(err)
	assert.NotEmpty(votes)
}

func TestGetVotesForUserForImage(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)
	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := CreateTestTagForImageWithVote(u.ID, i.ID, core.UUIDv4().ToShortString(), tx)
	assert.Nil(err)

	_, err = CreateOrUpdateVote(u.ID, i.ID, tag.ID, false, tx)
	assert.Nil(err)

	votes, err := GetVotesForUserForImage(u.ID, i.ID, tx)
	assert.Nil(err)
	assert.NotEmpty(votes)
}

func TestGetVotesForUserForTag(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)
	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := CreateTestTagForImageWithVote(u.ID, i.ID, core.UUIDv4().ToShortString(), tx)
	assert.Nil(err)

	_, err = CreateOrUpdateVote(u.ID, i.ID, tag.ID, false, tx)
	assert.Nil(err)

	votes, err := GetVotesForUserForTag(u.ID, tag.ID, tx)
	assert.Nil(err)
	assert.NotEmpty(votes)
}

func TestGetVote(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)
	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)
	tag, err := CreateTestTagForImageWithVote(u.ID, i.ID, core.UUIDv4().ToShortString(), tx)
	assert.Nil(err)

	_, err = CreateOrUpdateVote(u.ID, i.ID, tag.ID, false, tx)
	assert.Nil(err)

	vote, err := GetVote(u.ID, i.ID, tag.ID, tx)
	assert.Nil(err)
	assert.False(vote.IsZero())
}
