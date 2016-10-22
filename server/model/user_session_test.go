package model

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestDeleteUserSession(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	us, err := CreateTestUserSession(u.ID, tx)
	assert.Nil(err)

	err = DeleteUserSession(u.ID, us.SessionID, tx)
	assert.Nil(err)

	var verify UserSession
	err = DB().GetByIDInTx(&verify, tx, us.SessionID)
	assert.Nil(err)
	assert.True(verify.IsZero())
}
