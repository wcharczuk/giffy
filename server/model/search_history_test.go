package model

import (
	"context"
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/testutil"
)

func TestGetSearchHistory(t *testing.T) {
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

	tag, err := m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, "__test_search_history")
	assert.Nil(err)

	_, err = m.CreateTestSearchHistory(todo, "unit test", "test search", &i.ID, &tag.ID)
	assert.Nil(err)

	history, err := m.GetSearchHistory(todo)
	assert.Nil(err)
	assert.NotEmpty(history)
}
