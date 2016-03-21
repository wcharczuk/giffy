package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/blendlabs/go-util"
	"github.com/blendlabs/httprouter"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/core/auth"
	"github.com/wcharczuk/giffy/server/core/web"
)

const (
	// OAuthProviderGoogle is the only auth provider we use right now.
	OAuthProviderGoogle = "google"
)

func indexAction(session *auth.Session, ctx *web.HTTPContext) web.ControllerResult {
	return ctx.Static("server/_static/index.html")
}

// Init inits the app.
func Init() *httprouter.Router {
	core.DBInit()

	util.StartProcessQueueDispatchers(1)

	web.InitViewCache(
		"server/_views/header.html",
		"server/_views/footer.html",
		"server/_views/upload_image.html",
		"server/_views/upload_image_complete.html",
	)

	router := httprouter.New()

	new(APIController).Register(router)
	new(AuthController).Register(router)
	new(ImageController).Register(router)

	router.GET("/", web.ActionHandler(indexAction))

	//static files
	router.ServeFiles("/static/*filepath", http.Dir("server/_static"))

	router.NotFound = web.NotFoundHandler
	router.PanicHandler = web.PanicHandler

	return router
}

// Start starts the app.
func Start(router *httprouter.Router) {
	bindAddr := fmt.Sprintf(":%s", core.ConfigPort())
	server := &http.Server{
		Addr:    bindAddr,
		Handler: router,
	}
	web.Logf("Giffy Server Started, listening on %s", bindAddr)
	log.Fatal(server.ListenAndServe())
}
