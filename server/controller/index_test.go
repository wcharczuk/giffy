package controller

import (
	"io/ioutil"
	"net/http"
	"strings"
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
	defer res.Body.Close()

	contents, err := ioutil.ReadAll(res.Body)
	assert.Nil(err)
	assert.NotEmpty(contents)

	assert.True(strings.Contains(string(contents), "giffy.min.js"))
}
