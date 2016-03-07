package main

import (
	"fmt"
	"net/http"

	"github.com/blendlabs/httprouter"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/core/web"
	"github.com/wcharczuk/giffy/server/model"
)

func getImagesAction(ctx *web.APIContext) *web.ServiceResponse {
	images, err := model.GetAllImages(nil)
	return ctx.Content(images)
}

func getTagsAction(ctx *web.APIContext) *web.ServiceResponse {
	tags, err := model.GetAllTags(nil)
	return ctx.Content(tags)
}

func getUsersAction(ctx *web.APIContext) *web.ServiceResponse {
	users, err := model.GetAllUsers(nil)
	return ctx.Content(users)
}

func initRouter(router *httprouter.Router) {
	router.GET("/images", web.APIActionHandler(getImagesAction))
	router.GET("/tags", web.APIActionHandler(getTagsAction))
	router.GET("/users", web.APIActionHandler(getUsersAction))
}

func main() {
	core.DBInit()

	router := httprouter.New()
	initRouter(router)

	bindAddr := fmt.Sprintf(":%s", core.ConfigPort())
	server := &http.Server{
		Addr:    bindAddr,
		Handler: router,
	}
	server.ListenAndServe()
}
