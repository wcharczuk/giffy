package model

import (
	"context"
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/testutil"
	"github.com/blend/go-sdk/uuid"
)

func TestGetVotesForUser(t *testing.T) {
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
	tag, err := m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, uuid.V4().String())
	assert.Nil(err)

	_, err = m.CreateOrUpdateVote(todo, u.ID, i.ID, tag.ID, false)
	assert.Nil(err)

	votes, err := m.GetVotesForUser(todo, u.ID)
	assert.Nil(err)
	assert.NotEmpty(votes)
}

func TestGetVotesForImage(t *testing.T) {
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
	tag, err := m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, uuid.V4().String())
	assert.Nil(err)

	_, err = m.CreateOrUpdateVote(todo, u.ID, i.ID, tag.ID, false)
	assert.Nil(err)

	votes, err := m.GetVotesForImage(todo, i.ID)
	assert.Nil(err)
	assert.NotEmpty(votes)
}

func TestGetVotesForTag(t *testing.T) {
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
	tag, err := m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, uuid.V4().String())
	assert.Nil(err)

	_, err = m.CreateOrUpdateVote(todo, u.ID, i.ID, tag.ID, false)
	assert.Nil(err)

	votes, err := m.GetVotesForTag(todo, tag.ID)
	assert.Nil(err)
	assert.NotEmpty(votes)
}

func TestGetVotesForUserForImage(t *testing.T) {
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
	tag, err := m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, uuid.V4().String())
	assert.Nil(err)

	_, err = m.CreateOrUpdateVote(todo, u.ID, i.ID, tag.ID, false)
	assert.Nil(err)

	votes, err := m.GetVotesForUserForImage(todo, u.ID, i.ID)
	assert.Nil(err)
	assert.NotEmpty(votes)
}

func TestGetVotesForUserForTag(t *testing.T) {
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
	tag, err := m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, uuid.V4().String())
	assert.Nil(err)

	_, err = m.CreateOrUpdateVote(todo, u.ID, i.ID, tag.ID, false)
	assert.Nil(err)

	votes, err := m.GetVotesForUserForTag(todo, u.ID, tag.ID)
	assert.Nil(err)
	assert.NotEmpty(votes)
}

func TestGetVote(t *testing.T) {
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
	tag, err := m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, uuid.V4().String())
	assert.Nil(err)

	_, err = m.CreateOrUpdateVote(todo, u.ID, i.ID, tag.ID, false)
	assert.Nil(err)

	vote, err := m.GetVote(todo, u.ID, i.ID, tag.ID)
	assert.Nil(err)
	assert.False(vote.IsZero())
}
