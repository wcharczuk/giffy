package model

import (
	"testing"
	"time"

	assert "github.com/blendlabs/go-assert"
	util "github.com/blendlabs/go-util"
)

func TestSlackTeam(t *testing.T) {
	assert := assert.New(t)
	tx, err := DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	newTeam := &SlackTeam{
		TeamID:              util.String.RandomString(32),
		TeamName:            util.String.RandomString(128),
		CreatedUTC:          time.Now().UTC(),
		IsEnabled:           true,
		ContentRatingFilter: ContentRatingPG13,
		CreatedByID:         util.String.RandomString(32),
		CreatedByName:       util.String.RandomString(128),
	}
	err = DB().CreateInTx(newTeam, tx)
	assert.Nil(err)

	var verify SlackTeam
	err = DB().GetByIDInTx(&verify, tx, newTeam.TeamID)
	assert.Nil(err)
	assert.False(verify.IsZero())
}
