package model

import (
	"context"
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/testutil"
)

func TestDeleteUserSession(t *testing.T) {
	assert := assert.New(t)
	todo := context.TODO()
	tx, err := testutil.DefaultDB().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := NewTestManager(tx)

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	us, err := m.CreateTestUserSession(todo, u.ID)
	assert.Nil(err)

	err = m.DeleteUserSession(todo, u.ID, us.SessionID)
	assert.Nil(err)

	var verify UserSession
	_, err = m.Invoke(todo).Get(&verify, us.SessionID)
	assert.Nil(err)
	assert.True(verify.IsZero())
}
