package model

import (
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/db"
	util "github.com/blend/go-sdk/util"
)

func TestGetUserAuthByToken(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	key, err := util.Crypto.CreateKey(32)
	assert.Nil(err)

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	_, err = m.CreateTestUserAuth(todo, u.ID, "test", "password", key)
	assert.Nil(err)

	verify, err := m.GetUserAuthByToken(todo, "test", key)
	assert.Nil(err)
	assert.False(verify.IsZero())

	assert.Equal(u.ID, verify.UserID)
	assert.Equal("test", verify.Provider)
}

func TestDeleteUserAuthForProvider(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	key, err := util.Crypto.CreateKey(32)
	assert.Nil(err)

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	_, err = m.CreateTestUserAuth(todo, u.ID, "test", "password", key)
	assert.Nil(err)

	err = m.DeleteUserAuthForProvider(todo, u.ID, "test")
	assert.Nil(err)

	verify, err := m.GetUserAuthByToken(todo, "test", key)
	assert.Nil(err)
	assert.True(verify.IsZero())
}
