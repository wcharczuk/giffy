package model

import (
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/db"
)

func TestDeleteUserSession(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	us, err := m.CreateTestUserSession(todo, u.ID)
	assert.Nil(err)

	err = m.DeleteUserSession(todo, u.ID, us.SessionID)
	assert.Nil(err)

	var verify UserSession
	err = m.Invoke(todo).Get(&verify, us.SessionID)
	assert.Nil(err)
	assert.True(verify.IsZero())
}
