package controller

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-logger"
	util "github.com/blendlabs/go-util"
	"github.com/blendlabs/go-util/uuid"
	"github.com/blendlabs/go-web"
	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/filemanager"
	"github.com/wcharczuk/giffy/server/model"
)

func TestUploadImageByPostedFile(t *testing.T) {
	assert := assert.New(t)
	tx, err := model.DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	fm := filemanager.New(uuid.V4().String(), config.NewAwsFromEnv())
	fm.Mock()
	defer fm.ReleaseMock()

	auth, session := MockAuth(assert, tx, MockAdminLogin)
	defer MockLogout(assert, auth, session, tx)

	f, err := os.Open("server/controller/testdata/image.gif")

	contents, err := ioutil.ReadAll(f)
	assert.Nil(err)

	app := web.New()
	app.WithAuth(auth)
	app.WithLogger(logger.None())
	app.Register(UploadImage{Config: config.NewFromEnv(), Files: fm})
	app.Views().AddPaths(
		"server/_views/footer.html",
		"server/_views/upload_image.html",
		"server/_views/upload_image_complete.html",
		"server/_views/header.html",
	)
	err = app.Views().Initialize()
	assert.Nil(err)
	res, err := app.Mock().WithTx(tx).WithVerb("POST").WithPathf("/images/upload").WithPostedFile(web.PostedFile{
		Key:      "image",
		FileName: "image.gif",
		Contents: contents,
	}).WithCookieValue(auth.CookieName(), session.SessionID).Response()
	assert.Nil(err)
	assert.Equal(http.StatusOK, res.StatusCode)
	assert.NotNil(res.Body)

	defer res.Body.Close()

	resContents, err := ioutil.ReadAll(res.Body)
	assert.Nil(err)
	assert.True(strings.Contains(string(resContents), "SUCCESS"))

	imagesByUser, err := model.GetImagesForUserID(util.Parse.Int64(session.UserID), tx)
	assert.Nil(err)
	assert.NotEmpty(imagesByUser)
	assert.Equal("image.gif", imagesByUser[0].DisplayName)
	assert.Equal(58285, imagesByUser[0].FileSize)
	assert.Equal(275, imagesByUser[0].Width)
	assert.Equal(364, imagesByUser[0].Height)
	assert.NotEmpty(imagesByUser[0].MD5)
}
