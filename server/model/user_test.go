package model

import (
	"context"
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/testutil"
)

func TestGetUsersByCountAndOffset(t *testing.T) {
	assert := assert.New(t)
	todo := context.TODO()
	tx, err := testutil.DefaultDB().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := NewTestManager(tx)

	_, err = m.CreateTestUser(todo)
	assert.Nil(err)

	users, err := m.GetUsersByCountAndOffset(todo, 10, 0)
	assert.Nil(err)
	assert.NotEmpty(users)
}

func TestGetAllUsers(t *testing.T) {
	assert := assert.New(t)
	todo := context.TODO()
	tx, err := testutil.DefaultDB().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := NewTestManager(tx)

	_, err = m.CreateTestUser(todo)
	assert.Nil(err)

	all, err := m.GetAllUsers(todo)
	assert.Nil(err)
	assert.NotEmpty(all)
}
