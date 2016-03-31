package model

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-util"
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

	tag, err := CreateTestTagForImageWithVote(u.ID, i.ID, "__test_search_history", tx)
	assert.Nil(err)

	_, err = CreateTestSearchHistory("unit test", "test search", util.OptionalInt64(i.ID), util.OptionalInt64(tag.ID), tx)
	assert.Nil(err)

	history, err := GetSearchHistory(tx)
	assert.Nil(err)
	assert.NotEmpty(history)
}
