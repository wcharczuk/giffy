package model

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/spiffy"
)

func TestDeleteUserSession(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := spiffy.DefaultDb().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := createTestUser(tx)
	assert.Nil(err)

	us, err := createTestUserSession(u.ID, tx)
	assert.Nil(err)

	err = DeleteUserSession(u.ID, us.SessionID, tx)
	assert.Nil(err)

	var verify UserSession
	err = spiffy.DefaultDb().GetByIDInTransaction(&verify, tx, us.SessionID)
	assert.Nil(err)
	assert.True(verify.IsZero())
}
