package auth

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/wcharczuk/giffy/server/model"
)

func TestNewSession(t *testing.T) {
	assert := assert.New(t)

	modelSession := model.NewUserSession(1)
	session := NewSession(1, modelSession.SessionID)
	assert.Equal(modelSession.UserID, session.UserID)
	assert.Equal(modelSession.SessionID, session.SessionID)
}
