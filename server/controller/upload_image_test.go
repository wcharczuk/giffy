package controller

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-web"
	"github.com/wcharczuk/giffy/server/filecache"
	"github.com/wcharczuk/giffy/server/model"
)

func TestUploadImageByPostedFile(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	filecache.Mock()
	defer filecache.ReleaseMock()

	auth, session := MockAuth(assert, tx, MockModeratorLogin)
	defer auth.Logout(session, nil)

	f, err := os.Open("server/controller/testdata/image.gif")

	contents, err := ioutil.ReadAll(f)
	assert.Nil(err)

	app := web.New()
	app.SetAuth(auth)
	app.SetDiagnostics(logger.NewDiagnosticsAgent(logger.NewEventFlagSetNone()))
	app.IsolateTo(tx)
	app.Register(UploadImage{})
	app.View().AddPaths(
		"server/_views/footer.html",
		"server/_views/upload_image.html",
		"server/_views/upload_image_complete.html",
		"server/_views/header.html",
	)
	err = app.View().Initialize()
	assert.Nil(err)
	res, err := app.Mock().WithVerb("POST").WithPathf("/images/upload").WithPostedFile(web.PostedFile{
		Key:      "image",
		FileName: "image.gif",
		Contents: contents,
	}).WithHeader(auth.SessionParamName(), session.SessionID).FetchResponse()
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
