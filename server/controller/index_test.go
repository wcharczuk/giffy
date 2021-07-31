package controller

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/web"
	"github.com/wcharczuk/giffy/server/config"
)

func TestIndex(t *testing.T) {
	assert := assert.New(t)

	app := web.MustNew()
	app.Register(Index{Config: config.MustNewFromEnv()})
	res, err := web.MockGet(app, "/").Do()
	assert.Nil(err)
	assert.True(res.StatusCode == http.StatusOK || res.StatusCode == http.StatusNotModified, res.StatusCode)
	defer res.Body.Close()

	contents, err := ioutil.ReadAll(res.Body)
	assert.Nil(err)
	assert.NotEmpty(contents)

	assert.Matches("giffy.min.(.*).js", string(contents))
}

func TestStaticRewrite(t *testing.T) {
	assert := assert.New(t)

	app := web.MustNew()
	app.Register(Index{Config: new(config.Giffy)})
	contents, _, err := web.MockGet(app, "/static/js/giffy.min.1231231232313.js").Bytes()
	assert.Nil(err)
	assert.NotEmpty(contents)

	contents, _, err = web.MockGet(app, "/static/js/giffy.min.js").Bytes()
	assert.Nil(err)
	assert.NotEmpty(contents)
}
