package model

import (
	"context"
	"testing"
	"time"

	assert "github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/testutil"
	"github.com/blend/go-sdk/uuid"
)

func TestSlackTeam(t *testing.T) {
	assert := assert.New(t)
	todo := context.TODO()
	tx, err := testutil.DefaultDB().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := NewTestManager(tx)

	newTeam := &SlackTeam{
		TeamID:              uuid.V4().String(),
		TeamName:            uuid.V4().String(),
		CreatedUTC:          time.Now().UTC(),
		IsEnabled:           true,
		ContentRatingFilter: ContentRatingPG13,
		CreatedByID:         uuid.V4().String(),
		CreatedByName:       uuid.V4().String(),
	}
	err = m.Invoke(todo).Create(newTeam)
	assert.Nil(err)

	var verify SlackTeam
	_, err = m.Invoke(todo).Get(&verify, newTeam.TeamID)
	assert.Nil(err)
	assert.False(verify.IsZero())
}
