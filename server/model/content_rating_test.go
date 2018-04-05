package model

import (
	"testing"

	"github.com/blend/go-sdk/assert"
)

func TestAllContentRatings(t *testing.T) {
	assert := assert.New(t)
	var all []ContentRating
	err := DB().GetAll(&all)
	assert.Nil(err)
	assert.NotEmpty(all)
}

func TestGetContentRatingByName(t *testing.T) {
	assert := assert.New(t)
	rating, err := GetContentRatingByName("G", nil)
	assert.Nil(err)
	assert.False(rating.IsZero())
}
