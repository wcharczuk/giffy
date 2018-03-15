package controller

import (
	"fmt"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-web"
	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/model"
)

// Index is the root controller.
type Index struct {
	Config *config.Giffy
}

func (i Index) indexAction(r *web.Ctx) web.Result {
	if i.Config.IsProduction() {
		return r.Static("_client/dist/index.html")
	}
	return r.Static("_client/src/index.html")
}

func (i Index) faviconAction(r *web.Ctx) web.Result {
	if i.Config.IsProduction() {
		return r.Static("_client/dist/images/favicon.ico")
	}
	return r.Static("_client/src/images/favicon.ico")
}

func (i Index) methodNotAllowedHandler(r *web.Ctx) web.Result {
	return r.View().BadRequest(fmt.Errorf("method Not Allowed"))
}

func (i Index) notFoundHandler(r *web.Ctx) web.Result {
	return i.indexAction(r)
}

func (i Index) panicHandler(r *web.Ctx, err interface{}) web.Result {
	return r.View().InternalError(exception.Newf("%v", err))
}

func (i Index) statusAction(r *web.Ctx) web.Result {
	_, err := model.DB().Query("select 'ok!'").Any()
	if err != nil {
		r.Logger().Error(err)
		return r.JSON().Result(map[string]interface{}{"status": false})
	}
	return r.JSON().Result(map[string]interface{}{"status": true})
}

// Register registers the controller
func (i Index) Register(app *web.App) {
	app.WithMethodNotAllowedHandler(i.methodNotAllowedHandler)
	app.WithNotFoundHandler(i.notFoundHandler)
	app.WithPanicAction(i.panicHandler)

	app.GET("/", i.indexAction)
	app.GET("/index.html", i.indexAction)
	app.GET("/favicon.ico", i.faviconAction)

	if i.Config.IsProduction() {
		app.Static("/static/*filepath", "_client/dist")
		app.AddStaticRewriteRule("/static/*filepath", `^(.*)\.([0-9]+)\.(css|js)$`, func(path string, parts ...string) string {
			if len(parts) < 4 {
				return path
			}
			return fmt.Sprintf("%s.%s", parts[1], parts[3])
		})
		app.AddStaticHeader("/static/*filepath", "access-control-allow-origin", "*")
		app.AddStaticHeader("/static/*filepath", "cache-control", "public,max-age=315360000")
	} else {
		app.Static("/bower/*filepath", "_client/bower")
		app.Static("/static/*filepath", "_client/src")
	}

	app.GET("/status", i.statusAction)
}
