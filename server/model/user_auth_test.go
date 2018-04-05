package model

import (
	"testing"

	"github.com/blend/go-sdk/assert"
	util "github.com/blend/go-sdk/util"
)

func TestGetUserAuthByToken(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	key, err := util.Crypto.CreateKey(32)
	assert.Nil(err)

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	_, err = CreateTestUserAuth(u.ID, "test", "password", key, tx)
	assert.Nil(err)

	verify, err := GetUserAuthByToken("test", key, tx)
	assert.Nil(err)
	assert.False(verify.IsZero())

	assert.Equal(u.ID, verify.UserID)
	assert.Equal("test", verify.Provider)
}

func TestDeleteUserAuthForProvider(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	key, err := util.Crypto.CreateKey(32)
	assert.Nil(err)

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	_, err = CreateTestUserAuth(u.ID, "test", "password", key, tx)
	assert.Nil(err)

	err = DeleteUserAuthForProvider(u.ID, "test", tx)
	assert.Nil(err)

	verify, err := GetUserAuthByToken("test", key, tx)
	assert.Nil(err)
	assert.True(verify.IsZero())
}
