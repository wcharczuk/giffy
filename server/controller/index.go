package controller

import (
	"net/http"

	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/go-web"
)

// Index is the root controller.
type Index struct{}

func (i Index) indexAction(r *web.RequestContext) web.ControllerResult {
	if core.ConfigIsProduction() {
		return r.Static("_client/dist/index.html")
	}
	return r.Static("_client/src/index.html")
}

func (i Index) faviconAction(r *web.RequestContext) web.ControllerResult {
	if core.ConfigIsProduction() {
		return r.Static("_client/dist/images/favicon.ico")
	}
	return r.Static("_client/src/images/favicon.ico")
}

// Register registers the controller
func (i Index) Register(app *web.App) {
	app.GET("/", i.indexAction)
	app.GET("/favicon.ico", i.faviconAction)

	if core.ConfigIsProduction() {
		app.Static("/static/*filepath", http.Dir("_client/dist"))
	} else {
		app.Static("/bower/*filepath", http.Dir("_client/bower"))
		app.Static("/static/*filepath", http.Dir("_client/src"))
	}
}
