package controller

import (
	"fmt"
	"net/http"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-web"
	"github.com/wcharczuk/giffy/server/core"
)

// Index is the root controller.
type Index struct{}

func (i Index) indexAction(r *web.Ctx) web.Result {
	if core.ConfigIsProduction() {
		return r.Static("_client/dist/index.html")
	}
	return r.Static("_client/src/index.html")
}

func (i Index) faviconAction(r *web.Ctx) web.Result {
	if core.ConfigIsProduction() {
		return r.Static("_client/dist/images/favicon.ico")
	}
	return r.Static("_client/src/images/favicon.ico")
}

func (i Index) methodNotAllowedHandler(r *web.Ctx) web.Result {
	return r.View().BadRequest("Method Not Allowed")
}

func (i Index) notFoundHandler(r *web.Ctx) web.Result {
	return i.indexAction(r)
}

func (i Index) panicHandler(r *web.Ctx, err interface{}) web.Result {
	return r.View().InternalError(exception.Newf("%v", err))
}

// Register registers the controller
func (i Index) Register(app *web.App) {
	app.SetMethodNotAllowedHandler(i.methodNotAllowedHandler)
	app.SetNotFoundHandler(i.notFoundHandler)
	app.SetPanicHandler(i.panicHandler)

	app.GET("/", i.indexAction)
	app.GET("/index.html", i.indexAction)
	app.GET("/favicon.ico", i.faviconAction)

	if core.ConfigIsProduction() {
		app.Static("/static/*filepath", http.Dir("_client/dist"))
		app.AddStaticRewriteRule("/static/*filepath", `^(.*)\.([0-9]+)\.(css|js)$`, func(path string, parts ...string) string {
			if len(parts) < 4 {
				return path
			}
			return fmt.Sprintf("%s.%s", parts[1], parts[3])
		})
		app.AddStaticHeader("/static/*filepath", "access-control-allow-origin", "*")
		app.AddStaticHeader("/static/*filepath", "cache-control", "public,max-age=315360000")
	} else {
		app.Static("/bower/*filepath", http.Dir("_client/bower"))
		app.Static("/static/*filepath", http.Dir("_client/src"))
	}
}
