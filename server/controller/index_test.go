package controller

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/go-web"
)

func TestIndex(t *testing.T) {
	assert := assert.New(t)

	app := web.New()
	app.Register(Index{})
	res, err := app.Mock().WithPathf("/").FetchResponse()
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

	oldEnv := core.ConfigEnvironment()

	os.Setenv("GIFFY_ENV", "prod")
	defer os.Setenv("GIFFY_ENV", oldEnv)

	app := web.New()
	app.Register(Index{})
	contents, err := app.Mock().WithPathf("/static/js/giffy.min.1231231232313.js").FetchResponseAsBytes()
	assert.Nil(err)
	assert.NotEmpty(contents)

	contents, err = app.Mock().WithPathf("/static/js/giffy.min.js").FetchResponseAsBytes()
	assert.Nil(err)
	assert.NotEmpty(contents)
}
