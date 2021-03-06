package controller

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/uuid"
	"github.com/blend/go-sdk/web"
	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/filemanager"
	"github.com/wcharczuk/giffy/server/model"
)

func TestUploadImageByPostedFile(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.Manager{DB: db.Default(), Tx: tx}

	fm := filemanager.New(uuid.V4().String(), config.NewAwsFromEnv())
	fm.Mock()
	defer fm.ReleaseMock()

	auth, session := MockAuth(assert, &m, MockAdminLogin)
	defer MockLogout(assert, &m, auth, session)

	f, err := os.Open("server/controller/testdata/image.gif")

	contents, err := ioutil.ReadAll(f)
	assert.Nil(err)

	app := web.New()
	app.WithAuth(auth)
	app.WithLogger(logger.None())
	app.Register(UploadImage{Model: &m, Config: config.NewFromEnv(), Files: fm})
	app.Views().AddPaths(
		"server/_views/footer.html",
		"server/_views/upload_image.html",
		"server/_views/upload_image_complete.html",
		"server/_views/header.html",
	)
	err = app.Views().Initialize()
	assert.Nil(err)
	res, err := app.Mock().WithVerb("POST").WithPathf("/images/upload").WithPostedFile(web.PostedFile{
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

	imagesByUser, err := m.GetImagesForUserID(todo, parseInt64(session.UserID))
	assert.Nil(err)
	assert.NotEmpty(imagesByUser)
	assert.Equal("image.gif", imagesByUser[0].DisplayName)
	assert.Equal(58285, imagesByUser[0].FileSize)
	assert.Equal(275, imagesByUser[0].Width)
	assert.Equal(364, imagesByUser[0].Height)
	assert.NotEmpty(imagesByUser[0].MD5)
}
