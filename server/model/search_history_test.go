package model

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/spiffy"
)

func TestGetSearchHistory(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := spiffy.DefaultDb().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	t, err := CreateTestTagForImageWithVote(u.ID, i.ID, "__test_search_history", tx)
	assert.Nil(err)
}
