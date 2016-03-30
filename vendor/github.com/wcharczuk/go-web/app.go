package web

import (
	"fmt"
	"html/template"
	"net/http"
)

// New returns a new app.
func New() *App {
	return &App{
		router:             NewRouter(),
		name:               "Web",
		apiResultProvider:  NewAPIResultProvider(nil),
		viewResultProvider: NewViewResultProvider(nil, nil),
	}
}

// NewWithLogger returns a new app with a given logger.
func NewWithLogger(logger Logger) *App {
	return &App{
		router:             NewRouter(),
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
	router    Router
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
	a.router.GET(path, handler)
}

// OPTIONS registers a OPTIONS request handler.
func (a *App) OPTIONS(path string, handler ControllerAction) {
	a.router.OPTIONS(path, handler)
}

// HEAD registers a HEAD request handler.
func (a *App) HEAD(path string, handler ControllerAction) {
	a.router.HEAD(path, handler)
}

// PUT registers a PUT request handler.
func (a *App) PUT(path string, handler ControllerAction) {
	a.router.PUT(path, handler)
}

// POST registers a POST request handler.
func (a *App) POST(path string, handler ControllerAction) {
	a.router.POST(path, handler)
}

// DELETE registers a DELETE request handler.
func (a *App) DELETE(path string, handler ControllerAction) {
	a.router.DELETE(path, handler)
}

// Static registers a Static request handler.
func (a *App) Static(path string, root http.FileSystem) {
	a.router.Static(path, root)
}

// SetNotFoundHandler sets the not found handler.
func (a *App) SetNotFoundHandler(handler ControllerAction) {
	a.router.SetNotFoundHandler(handler)
}

// SetMethodNotAllowedHandler sets the not found handler.
func (a *App) SetMethodNotAllowedHandler(handler ControllerAction) {
	a.router.SetMethodNotAllowedHandler(handler)
}

// SetPanicHandler sets the not found handler.
func (a *App) SetPanicHandler(handler PanicControllerAction) {
	a.router.SetPanicHandler(handler)
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
