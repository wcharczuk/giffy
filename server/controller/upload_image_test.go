package controller

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/r2"
	"github.com/blend/go-sdk/testutil"
	"github.com/blend/go-sdk/uuid"
	"github.com/blend/go-sdk/web"
	"github.com/blend/go-sdk/webutil"
	"github.com/wcharczuk/giffy/server/awsutil"
	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/filemanager"
	"github.com/wcharczuk/giffy/server/model"
)

func TestUploadImageByPostedFile(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := testutil.DefaultDB().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := model.NewTestManager(tx)

	fm := filemanager.New(uuid.V4().String(), awsutil.Config{})
	fm.Mock()
	defer fm.ReleaseMock()

	auth, session := MockAuth(assert, &m, MockAdminLogin)
	defer MockLogout(assert, &m, auth, session)

	contents, err := ioutil.ReadFile("server/controller/testdata/image.gif")
	assert.Nil(err)

	app := web.MustNew(
		web.OptAuth(*auth, nil),
		web.OptLog(logger.All()),
	)
	app.Register(UploadImage{
		Log:    app.Log,
		Model:  &m,
		Config: config.MustNewFromEnv(),
		Files:  fm,
	})
	app.Views.AddPaths(
		"server/_views/footer.html",
		"server/_views/upload_image.html",
		"server/_views/upload_image_complete.html",
		"server/_views/header.html",
	)

	resContents, res, err := web.MockMethod(app, http.MethodPost, "/images/upload",
		r2.OptCookieValue(auth.CookieDefaults.Name, session.SessionID),
		r2.OptPostedFiles(webutil.PostedFile{
			Key:      "image",
			FileName: "image.gif",
			Contents: contents,
		}),
	).Bytes()
	assert.Nil(err)
	assert.Equal(http.StatusOK, res.StatusCode, string(resContents))
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
