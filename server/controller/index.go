package controller

import (
	"fmt"

	exception "github.com/blend/go-sdk/ex"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/web"
	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/model"
)

// Index is the root controller.
type Index struct {
	Log    logger.Log
	Config *config.Giffy
	Model  *model.Manager
}

func (i Index) indexAction(r *web.Ctx) web.Result {
	return web.Static("_client/dist/index.html")
}

func (i Index) privacyAction(r *web.Ctx) web.Result {
	return r.Views.View("privacy", nil)
}

func (i Index) tosAction(r *web.Ctx) web.Result {
	return r.Views.View("tos", nil)
}

func (i Index) faviconAction(r *web.Ctx) web.Result {
	return web.Static("_client/dist/images/favicon.ico")
}

func (i Index) methodNotAllowedHandler(r *web.Ctx) web.Result {
	return r.Views.BadRequest(fmt.Errorf("method Not Allowed"))
}

func (i Index) notFoundHandler(r *web.Ctx) web.Result {
	return i.indexAction(r)
}

func (i Index) panicHandler(r *web.Ctx, err interface{}) web.Result {
	return r.Views.InternalError(exception.New(err))
}

func (i Index) statusAction(r *web.Ctx) web.Result {
	_, err := i.Model.Invoke(r.Context()).Query("select 'ok!'").Any()
	if err != nil {
		logger.MaybeError(i.Log, err)
		return web.JSON.Result(map[string]interface{}{"status": false})
	}
	return web.JSON.Result(map[string]interface{}{"status": true})
}

// Register registers the controller
func (i Index) Register(app *web.App) {
	app.MethodNotAllowedHandler = app.RenderAction(i.methodNotAllowedHandler)
	app.NotFoundHandler = app.RenderAction(i.notFoundHandler)
	app.PanicAction = i.panicHandler

	app.GET("/", i.indexAction)
	app.GET("/index.html", i.indexAction)
	app.GET("/favicon.ico", i.faviconAction)

	app.GET("/privacy", i.privacyAction)
	app.GET("/tos", i.tosAction)

	app.ServeStaticCached("/static", []string{"_client/dist"})
	app.SetStaticRewriteRule("/static", `^(.*)\.([0-9]+)\.(css|js)$`, func(path string, parts ...string) string {
		if len(parts) < 4 {
			return path
		}
		return fmt.Sprintf("%s.%s", parts[1], parts[3])
	})
	app.SetStaticHeader("/static", "access-control-allow-origin", "*")
	app.SetStaticHeader("/static", "cache-control", "public,max-age=315360000")

	app.GET("/status", i.statusAction)
}
