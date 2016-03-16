package model

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/spiffy"
)

func TestGetUserAuthByToken(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := spiffy.DefaultDb().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := createTestUser(tx)
	assert.Nil(err)

	_, err = createTestUserAuth(u.ID, "test", "password", tx)
	assert.Nil(err)

	verify, err := GetUserAuthByToken("test", tx)
	assert.Nil(err)
	assert.False(verify.IsZero())

	assert.Equal(u.ID, verify.UserID)
	assert.Equal("test", verify.Provider)
}
