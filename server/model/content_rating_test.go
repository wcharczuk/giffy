package model

import (
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/db/dbutil"
	"github.com/blend/go-sdk/testutil"
)

func TestGetContentRatingByName(t *testing.T) {
	assert := assert.New(t)
	todo := todo()
	tx, err := testutil.DefaultDB().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{BaseManager: dbutil.NewBaseManager(testutil.DefaultDB(), db.OptTx(tx))}

	rating, err := m.GetContentRatingByName(todo, "G")
	assert.Nil(err)
	assert.False(rating.IsZero())
}
