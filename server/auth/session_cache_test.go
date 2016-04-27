package auth

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-util"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/model"
)

func TestSessionCacheSingleton(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(SessionState())
}

func TestSessionCache(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := model.CreateTestUser(tx)
	assert.Nil(err)

	sessionID := util.RandomString(32)
	cache := newSessionCache()
	cache.Add(u.ID, sessionID, tx)
	assert.NotEmpty(cache.Sessions)
	session := cache.Get(sessionID)
	assert.NotNil(session)
	assert.NotNil(session.User)
	assert.Equal(sessionID, session.SessionID)
	assert.True(cache.IsActive(sessionID))
	cache.Expire(sessionID)
	assert.False(cache.IsActive(sessionID))
}
