package model

import (
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/util"
)

func TestGetSearchHistory(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
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
