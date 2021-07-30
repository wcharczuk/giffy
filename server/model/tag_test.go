package model

import (
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/db"
)

func TestGetAllTags(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	_, err = m.CreateTestTag(todo, u.ID, "test")
	assert.Nil(err)

	all, err := m.GetAllTags(todo)
	assert.Nil(err)
	assert.NotEmpty(all)
}

func TestGetTagByID(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	tag, err := m.CreateTestTag(todo, u.ID, "test")
	assert.Nil(err)

	verify, err := m.GetTagByID(todo, tag.ID)
	assert.Nil(err)
	assert.False(verify.IsZero())
}

func TestGetTagByUUID(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	tag, err := m.CreateTestTag(todo, u.ID, "test")
	assert.Nil(err)

	verify, err := m.GetTagByUUID(todo, tag.UUID)
	assert.Nil(err)
	assert.False(verify.IsZero())
}

func TestGetTagByValue(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	_, err = m.CreateTestTag(todo, u.ID, "test")
	assert.Nil(err)

	verify, err := m.GetTagByValue(todo, "test")
	assert.Nil(err)
	assert.False(verify.IsZero())
	assert.Equal("test", verify.TagValue)
}

func TestMergeTags(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u1, err := m.CreateTestUser(todo)
	assert.Nil(err)
	u2, err := m.CreateTestUser(todo)
	assert.Nil(err)

	i1, err := m.CreateTestImage(todo, u1.ID)
	assert.Nil(err)
	t1, err := m.CreateTestTagForImageWithVote(todo, u1.ID, i1.ID, "__test_tag1")
	assert.Nil(err)

	i2, err := m.CreateTestImage(todo, u2.ID)
	assert.Nil(err)
	t2, err := m.CreateTestTagForImageWithVote(todo, u2.ID, i2.ID, "__test_tag2")
	assert.Nil(err)

	mergeErr := m.MergeTags(todo, t1.ID, t2.ID)
	assert.Nil(mergeErr)

	verify, err := m.GetVoteSummary(todo, i1.ID, t2.ID)
	assert.Nil(err)
	assert.False(verify.IsZero())
	assert.Equal(1, verify.VotesTotal)

	verify, err = m.GetVoteSummary(todo, i2.ID, t2.ID)
	assert.Nil(err)
	assert.False(verify.IsZero())
	assert.Equal(1, verify.VotesTotal)
}

func TestMergeTagsWithExisting(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u1, err := m.CreateTestUser(todo)
	assert.Nil(err)
	u2, err := m.CreateTestUser(todo)
	assert.Nil(err)

	i1, err := m.CreateTestImage(todo, u1.ID)
	assert.Nil(err)
	t1, err := m.CreateTestTagForImageWithVote(todo, u1.ID, i1.ID, "__test_tag1")
	assert.Nil(err)

	i2, err := m.CreateTestImage(todo, u2.ID)
	assert.Nil(err)
	t2, err := m.CreateTestTagForImageWithVote(todo, u2.ID, i2.ID, "__test_tag2")
	assert.Nil(err)

	// user 1 adds tag 1 to image 2
	_, err = m.CreatTestVoteSummaryWithVote(todo, i2.ID, t1.ID, u1.ID, 1, 0)
	assert.Nil(err)

	mergeErr := m.MergeTags(todo, t1.ID, t2.ID)
	assert.Nil(mergeErr)

	verify, err := m.GetVoteSummary(todo, i2.ID, t2.ID)
	assert.Nil(err)
	assert.Equal(2, verify.VotesTotal)
}

func TestDeleteTagAndVotesByID(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	i, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	tag, err := m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, "__test_tag_value")
	assert.Nil(err)

	err = m.DeleteTagAndVotesByID(todo, tag.ID)
	assert.Nil(err)

	verify, err := m.GetTagByID(todo, tag.ID)
	assert.Nil(err)
	assert.True(verify.IsZero())
}

func TestSetTagValue(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	tag, err := m.CreateTestTag(todo, u.ID, "__test_tag_value")
	assert.Nil(err)

	err = m.SetTagValue(todo, tag.ID, "__test_tag_value_2")
	assert.Nil(err)

	verify, err := m.GetTagByValue(todo, "__test_tag_value")
	assert.Nil(err)
	assert.True(verify.IsZero())
}
