package model

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/spiffy"
)

func TestAllContentRatings(t *testing.T) {
	assert := assert.New(t)
	var all []ContentRating
	err := spiffy.DefaultDb().GetAll(&all)
	assert.Nil(err)
	assert.NotEmpty(all)
}
