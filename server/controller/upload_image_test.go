package controller

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	assert "github.com/blendlabs/go-assert"
	logger "github.com/blendlabs/go-logger"
	"github.com/wcharczuk/giffy/server/auth"
	"github.com/wcharczuk/giffy/server/filecache"
	"github.com/wcharczuk/giffy/server/model"
	web "github.com/wcharczuk/go-web"
)

func TestUploadImageByPostedFile(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	filecache.Mock()
	defer filecache.ReleaseMock()

	session := MockAuth(assert, tx, MockModeratorLogin)
	defer auth.Logout(session.UserID, session.SessionID, nil, tx)

	f, err := os.Open("server/controller/testdata/image.gif")

	contents, err := ioutil.ReadAll(f)
	assert.Nil(err)

	app := web.New()
	app.SetDiagnostics(logger.NewDiagnosticsAgent(logger.EventAll))
	app.IsolateTo(tx)
	app.Register(UploadImage{})
	err = app.InitializeViewCache([]string{
		"server/_views/footer.html",
		"server/_views/upload_image.html",
		"server/_views/upload_image_complete.html",
		"server/_views/header.html",
	}...)
	assert.Nil(err)
	res, err := app.Mock().WithVerb("POST").WithPathf("/images/upload").WithPostedFile(web.PostedFile{
		Key:      "image",
		FileName: "image.gif",
		Contents: contents,
	}).WithHeader(auth.SessionParamName, session.SessionID).FetchResponse()
	assert.Nil(err)
	assert.Equal(http.StatusOK, res.StatusCode)
}
