package controller

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/go-web"
)

const (
	TestUserUUID = "a68aac8196e444d4a3e570192a20f369"
)

type testUserResponse struct {
	Meta     *web.APIResponseMeta `json:"meta"`
	Response *model.User          `json:"response"`
}

type testUsersResponse struct {
	Meta     *web.APIResponseMeta `json:"meta"`
	Response []model.User         `json:"response"`
}

type testImagesResponse struct {
	Meta     *web.APIResponseMeta `json:"meta"`
	Response []model.Image        `json:"response"`
}

func TestAPIUser(t *testing.T) {
	assert := assert.New(t)

	app := web.New()
	app.Register(API{})

	var res testUserResponse
	err := app.Mock().WithPathf("/api/user/%s", TestUserUUID).JSON(&res)
	assert.Nil(err)
	assert.NotNil(res)
	assert.NotNil(res.Response)
	assert.Equal(TestUserUUID, res.Response.UUID)
}

func TestAPIImages(t *testing.T) {
	assert := assert.New(t)

	app := web.New()
	app.Register(API{})

	var res testImagesResponse
	err := app.Mock().WithPathf("/api/images").JSON(&res)
	assert.Nil(err)
	assert.NotNil(res)
	assert.NotEmpty(res.Response)
}
