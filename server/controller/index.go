package controller

import (
	"fmt"
	"net/http"

	"github.com/blendlabs/go-exception"
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

func (i Index) methodNotAllowedHandler(r *web.RequestContext) web.ControllerResult {
	return r.View().BadRequest("Method Not Allowed")
}

func (i Index) notFoundHandler(r *web.RequestContext) web.ControllerResult {
	return r.View().NotFound()
}

func (i Index) panicHandler(r *web.RequestContext, err interface{}) web.ControllerResult {
	return r.View().InternalError(exception.Newf("%v", err))
}

// Register registers the controller
func (i Index) Register(app *web.App) {
	app.SetMethodNotAllowedHandler(i.methodNotAllowedHandler)
	app.SetNotFoundHandler(i.notFoundHandler)
	app.SetPanicHandler(i.panicHandler)

	app.GET("/", i.indexAction)
	app.GET("/favicon.ico", i.faviconAction)

	if core.ConfigIsProduction() {
		app.Static("/static/*filepath", http.Dir("_client/dist"))
		app.StaticRewrite("/static/*filepath", `^(.*)\.([0-9]+)\.(css|js)$`, func(path string, parts ...string) string {
			if len(parts) < 3 {
				return path
			}
			return fmt.Sprintf("%s.%s", parts[1], parts[3])
		})
		app.StaticHeader("/static/*filepath", "access-control-allow-origin", "*")
		app.StaticHeader("/static/*filepath", "cache-control", "public,max-age=315360000")
	} else {
		app.Static("/bower/*filepath", http.Dir("_client/bower"))
		app.Static("/static/*filepath", http.Dir("_client/src"))
	}
}
