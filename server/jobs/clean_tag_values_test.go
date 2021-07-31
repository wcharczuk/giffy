package jobs

import (
	"context"
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/testutil"
	"github.com/wcharczuk/giffy/server/model"
)

func TestCleanTagValues(t *testing.T) {
	assert := assert.New(t)
	tx, err := testutil.DefaultDB().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.NewTestManager(tx)

	ctx := context.TODO()

	u, err := m.CreateTestUser(ctx)
	assert.Nil(err)
	i, err := m.CreateTestImage(ctx, u.ID)
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(ctx, u.ID, i.ID, "winning's")
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(ctx, u.ID, i.ID, "they're")
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(ctx, u.ID, i.ID, "theyre")
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(ctx, u.ID, i.ID, "crushing it")
	assert.Nil(err)

	job := &CleanTagValues{Model: &m}
	err = job.Execute(ctx)

	assert.Nil(err)

	verify, err := m.GetTagByValue(ctx, "winning's")
	assert.Nil(err)
	assert.True(verify.IsZero())

	verify, err = m.GetTagByValue(ctx, "they're")
	assert.Nil(err)
	assert.True(verify.IsZero())

	verify, err = m.GetTagByValue(ctx, "theyre")
	assert.Nil(err)
	assert.False(verify.IsZero())

	verify, err = m.GetTagByValue(ctx, "crushing it")
	assert.Nil(err)
	assert.False(verify.IsZero())
}
