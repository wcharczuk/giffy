package controller

import (
	"net/http"
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/wcharczuk/go-web"
)

func TestIndex(t *testing.T) {
	assert := assert.New(t)

	app := web.New()
	app.Register(Index{})
	res, err := app.Mock().WithPathf("/").Response()
	assert.Nil(err)
	assert.True(res.StatusCode == http.StatusOK || res.StatusCode == http.StatusNotModified, res.StatusCode)
}
