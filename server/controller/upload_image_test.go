package controller

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
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
	app.SetDiagnostics(logger.NewDiagnosticsAgent(logger.NewEventFlagSetNone()))
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
	assert.NotNil(res.Body)

	defer res.Body.Close()

	resContents, err := ioutil.ReadAll(res.Body)
	assert.Nil(err)
	assert.True(strings.Contains(string(resContents), "SUCCESS"))

	imagesByUser, err := model.GetImagesForUserID(session.UserID, tx)
	assert.Nil(err)
	assert.NotEmpty(imagesByUser)
	assert.Equal("image.gif", imagesByUser[0].DisplayName)
	assert.Equal(58285, imagesByUser[0].FileSize)
	assert.Equal(275, imagesByUser[0].Width)
	assert.Equal(364, imagesByUser[0].Height)
	assert.NotEmpty(imagesByUser[0].MD5)
}
