package model

import (
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/db"
)

func TestGetContentRatingByName(t *testing.T) {
	assert := assert.New(t)
	todo := todo()
	tx, err := defaultDB().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	rating, err := m.GetContentRatingByName(todo, "G")
	assert.Nil(err)
	assert.False(rating.IsZero())
}
