package viewmodel

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestGetSiteStats(t *testing.T) {
	assert := assert.New(t)

	stats, err := GetSiteStats()
	assert.Nil(err)
	assert.NotNil(stats)
}
