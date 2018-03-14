package controller

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-web"
	"github.com/wcharczuk/giffy/server/config"
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

func TestStaticRewrite(t *testing.T) {
	assert := assert.New(t)

	app := web.New()
	app.Register(Index{Config: &config.Giffy{Environment: config.EnvironmentProd}})
	contents, err := app.Mock().WithPathf("/static/js/giffy.min.1231231232313.js").Bytes()
	assert.Nil(err)
	assert.NotEmpty(contents)

	contents, err = app.Mock().WithPathf("/static/js/giffy.min.js").Bytes()
	assert.Nil(err)
	assert.NotEmpty(contents)
}
