package model

import (
	"testing"
	"time"

	assert "github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/db"
	util "github.com/blend/go-sdk/util"
)

func TestSlackTeam(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	newTeam := &SlackTeam{
		TeamID:              util.String.Random(32),
		TeamName:            util.String.Random(128),
		CreatedUTC:          time.Now().UTC(),
		IsEnabled:           true,
		ContentRatingFilter: ContentRatingPG13,
		CreatedByID:         util.String.Random(32),
		CreatedByName:       util.String.Random(128),
	}
	err = m.Invoke(todo).Create(newTeam)
	assert.Nil(err)

	var verify SlackTeam
	err = m.Invoke(todo).Get(&verify, newTeam.TeamID)
	assert.Nil(err)
	assert.False(verify.IsZero())
}
