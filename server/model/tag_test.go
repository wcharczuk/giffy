package model

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/spiffy"
)

func TestGetTagByValue(t *testing.T) {
	assert := assert.New(t)
	tx, err := spiffy.DefaultDb().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u, err := createTestUser(tx)
	assert.Nil(err)
}
