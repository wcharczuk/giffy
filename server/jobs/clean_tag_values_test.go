package jobs

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/model"
)

func TestCleanTagValues(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := spiffy.DefaultDb().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := model.CreateTestUser(tx)
	assert.Nil(err)
	i, err := model.CreateTestImage(u.ID, tx)
	assert.Nil(err)

	_, err = model.CreateTestTagForImageWithVote(u.ID, i.ID, "winning's", tx)
	assert.Nil(err)

	_, err = model.CreateTestTagForImageWithVote(u.ID, i.ID, "they're", tx)
	assert.Nil(err)

	_, err = model.CreateTestTagForImageWithVote(u.ID, i.ID, "theyre", tx)
	assert.Nil(err)

	_, err = model.CreateTestTagForImageWithVote(u.ID, i.ID, "crushing it", tx)
	assert.Nil(err)

	ct := chronometer.NewCancellationToken()
	job := &CleanTagValues{}
	err = job.ExecuteInTransaction(ct, tx)

	assert.Nil(err)

	verify, err := model.GetTagByValue("winning's", tx)
	assert.Nil(err)
	assert.True(verify.IsZero())

	verify, err = model.GetTagByValue("they're", tx)
	assert.Nil(err)
	assert.True(verify.IsZero())

	verify, err = model.GetTagByValue("theyre", tx)
	assert.Nil(err)
	assert.False(verify.IsZero())

	verify, err = model.GetTagByValue("crushing it", tx)
	assert.Nil(err)
	assert.False(verify.IsZero())
}
