package web

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// New returns a new app.
func New() *App {
	return &App{
		router:             httprouter.New(),
		name:               "Web",
		apiResultProvider:  NewAPIResultProvider(nil),
		viewResultProvider: NewViewResultProvider(nil, nil),
	}
}

// NewWithLogger returns a new app with a given logger.
func NewWithLogger(logger Logger) *App {
	return &App{
		router:             httprouter.New(),
		name:               "Web",
		logger:             logger,
		apiResultProvider:  NewAPIResultProvider(logger),
		viewResultProvider: NewViewResultProvider(logger, nil),
	}
}

// App is the server for the app.
type App struct {
	name string

	logger    Logger
	router    *httprouter.Router
	viewCache *template.Template

	apiResultProvider  *APIResultProvider
	viewResultProvider *ViewResultProvider

	port string
}

// Name returns the app name.``
func (a *App) Name() string {
	return a.name
}

// SetName sets the app name
func (a *App) SetName(name string) {
	a.name = name
}

// Logger returns the logger for the app.
func (a *App) Logger() Logger {
	return a.logger
}

// SetLogger sets the logger.
func (a *App) SetLogger(l Logger) {
	a.logger = l
	if a.apiResultProvider != nil {
		a.apiResultProvider.logger = l
	}
	if a.viewResultProvider != nil {
		a.viewResultProvider.logger = l
	}
}

// ViewCache gets the view cache for the app.
func (a *App) ViewCache() *template.Template {
	return a.viewCache
}

// SetViewCache sets the view cache for the app.
func (a *App) SetViewCache(viewCache *template.Template) {
	a.viewCache = viewCache
	if a.viewResultProvider != nil {
		a.viewResultProvider.viewCache = viewCache
	}
}

// Port returns the port for the app.
func (a *App) Port() string {
	return a.port
}

// SetPort sets the port the app listens on.
func (a *App) SetPort(port string) {
	a.port = port
}

// Start starts the server and binds to the given address.
func (a *App) Start() error {
	bindAddr := fmt.Sprintf(":%s", a.port)
	server := &http.Server{
		Addr:    bindAddr,
		Handler: a,
	}
	a.logger.Logf("%s Started, listening on %s", a.Name(), bindAddr)
	return server.ListenAndServe()
}

// StartWithServer starts the app on a custom server.
func (a *App) StartWithServer(server *http.Server) error {
	server.Handler = a
	a.logger.Logf("%s Started, listening on %s", a.Name(), server.Addr)
	return server.ListenAndServe()
}

// Register registers a controller with the app's router.
func (a *App) Register(c Controller) {
	c.Register(a)
}

// InitViewCache caches templates by path.
func (a *App) InitViewCache(paths ...string) error {
	views, err := template.ParseFiles(paths...)
	if err != nil {
		return err
	}
	a.viewCache = template.Must(views, nil)
	a.viewResultProvider.viewCache = a.viewCache
	return nil
}

// GET registers a GET request handler.
func (a *App) GET(path string, handler ControllerAction) {
	a.router.GET(path, a.RenderAction(handler))
}

// OPTIONS registers a OPTIONS request handler.
func (a *App) OPTIONS(path string, handler ControllerAction) {
	a.router.OPTIONS(path, a.RenderAction(handler))
}

// HEAD registers a HEAD request handler.
func (a *App) HEAD(path string, handler ControllerAction) {
	a.router.HEAD(path, a.RenderAction(handler))
}

// PUT registers a PUT request handler.
func (a *App) PUT(path string, handler ControllerAction) {
	a.router.PUT(path, a.RenderAction(handler))
}

// POST registers a POST request handler.
func (a *App) POST(path string, handler ControllerAction) {
	a.router.POST(path, a.RenderAction(handler))
}

// DELETE registers a DELETE request handler.
func (a *App) DELETE(path string, handler ControllerAction) {
	a.router.DELETE(path, a.RenderAction(handler))
}

// Static registers a Static request handler.
func (a *App) Static(path string, root http.FileSystem) {
	a.router.ServeFiles(path, root)
}

// SetNotFoundHandler sets the not found handler.
func (a *App) SetNotFoundHandler(handler ControllerAction) {
	a.router.NotFound = newHandleShim(a, handler)
}

// SetMethodNotAllowedHandler sets the not found handler.
func (a *App) SetMethodNotAllowedHandler(handler ControllerAction) {
	a.router.MethodNotAllowed = newHandleShim(a, handler)
}

// SetPanicHandler sets the not found handler.
func (a *App) SetPanicHandler(handler PanicControllerAction) {
	a.router.PanicHandler = func(w http.ResponseWriter, r *http.Request, err interface{}) {
		a.RenderAction(func(r *RequestContext) ControllerResult {
			return handler(r, err)
		})(w, r, httprouter.Params{})
	}
}

func (a *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	a.router.ServeHTTP(w, req)
}

// RequestContext creates an http context.
func (a *App) RequestContext(w http.ResponseWriter, r *http.Request, p RouteParameters) *RequestContext {
	hc := NewRequestContext(w, r, p)
	hc.logger = a.logger
	hc.view = a.viewResultProvider
	hc.api = a.apiResultProvider
	return hc
}

// --------------------------------------------------------------------------------
// Render Methods
// --------------------------------------------------------------------------------

// RenderAction is the translation step from APIControllerAction to httprouter.Handle.
func (a *App) RenderAction(action ControllerAction) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			a.renderUncompressed(action, w, r, parseParams(p))
		} else {
			a.renderCompressed(action, w, r, parseParams(p))
		}
	}
}

func parseParams(p httprouter.Params) RouteParameters {
	rp := NewRouteParameters()
	for _, pv := range p {
		rp[pv.Key] = pv.Value
	}
	return rp
}

func (a *App) renderUncompressed(action ControllerAction, w http.ResponseWriter, r *http.Request, p RouteParameters) {
	w.Header().Set("Vary", "Accept-Encoding")
	rw := NewResponseWriter(w)
	context := a.RequestContext(rw, r, p)
	context.onRequestStart()
	context.Render(action(context))
	context.setStatusCode(rw.StatusCode)
	context.setContentLength(rw.ContentLength)
	context.onRequestEnd()
	context.LogRequest()
}

func (a *App) renderCompressed(action ControllerAction, w http.ResponseWriter, r *http.Request, p RouteParameters) {
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Vary", "Accept-Encoding")

	gzw := NewGZippedResponseWriter(w)
	defer gzw.Close()
	context := a.RequestContext(gzw, r, p)
	context.onRequestStart()
	result := action(context)
	context.Render(result)
	gzw.Flush()
	context.setStatusCode(gzw.StatusCode)
	context.setContentLength(gzw.BytesWritten)
	context.onRequestEnd()
	context.LogRequest()
}

func newHandleShim(app *App, handler ControllerAction) http.Handler {
	return &handleShim{action: handler, app: app}
}

type handleShim struct {
	action ControllerAction
	app    *App
}

func (hs handleShim) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hs.app.RenderAction(hs.action)(w, r, httprouter.Params{})
}
