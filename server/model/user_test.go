package model

import (
	"testing"

	"github.com/blend/go-sdk/assert"
)

func TestGetUsersByCountAndOffset(t *testing.T) {
	assert := assert.New(t)
	tx, err := DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	_, err = CreateTestUser(tx)
	assert.Nil(err)

	users, err := GetUsersByCountAndOffset(10, 0, tx)
	assert.Nil(err)
	assert.NotEmpty(users)
}

func TestGetAllUsers(t *testing.T) {
	assert := assert.New(t)
	tx, err := DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	_, err = CreateTestUser(tx)
	assert.Nil(err)

	all, err := GetAllUsers(tx)
	assert.Nil(err)
	assert.NotEmpty(all)
}
