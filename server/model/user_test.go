package model

import (
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/db"
)

func TestGetUsersByCountAndOffset(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	_, err = m.CreateTestUser(todo)
	assert.Nil(err)

	users, err := m.GetUsersByCountAndOffset(todo, 10, 0)
	assert.Nil(err)
	assert.NotEmpty(users)
}

func TestGetAllUsers(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	_, err = m.CreateTestUser(todo)
	assert.Nil(err)

	all, err := m.GetAllUsers(todo)
	assert.Nil(err)
	assert.NotEmpty(all)
}
