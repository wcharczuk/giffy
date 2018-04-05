package model

import (
	"testing"
	"time"

	assert "github.com/blend/go-sdk/assert"
	util "github.com/blend/go-sdk/util"
)

func TestSlackTeam(t *testing.T) {
	assert := assert.New(t)
	tx, err := DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	newTeam := &SlackTeam{
		TeamID:              util.String.Random(32),
		TeamName:            util.String.Random(128),
		CreatedUTC:          time.Now().UTC(),
		IsEnabled:           true,
		ContentRatingFilter: ContentRatingPG13,
		CreatedByID:         util.String.Random(32),
		CreatedByName:       util.String.Random(128),
	}
	err = DB().CreateInTx(newTeam, tx)
	assert.Nil(err)

	var verify SlackTeam
	err = DB().InTx(tx).Get(&verify, newTeam.TeamID)
	assert.Nil(err)
	assert.False(verify.IsZero())
}
