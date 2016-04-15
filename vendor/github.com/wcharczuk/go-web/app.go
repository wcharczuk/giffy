package web

import (
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// New returns a new app.
func New() *App {
	return &App{
		router:             httprouter.New(),
		name:               "Web",
		staticRewriteRules: map[string][]*RewriteRule{},
		staticHeaders:      map[string]http.Header{},
		onRequestStart: []RequestEventHandler{func(r *RequestContext) {
			r.onRequestStart()
		}},
		onRequestComplete: []RequestEventHandler{func(r *RequestContext) {
			r.onRequestEnd()
		}},
		onRequestError: []RequestEventErrorHandler{func(r *RequestContext, err interface{}) {
			if r != nil && r.logger != nil {
				r.logger.Error(err)
			}
		}},
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

	onRequestStart    []RequestEventHandler
	onRequestComplete []RequestEventHandler
	onRequestError    []RequestEventErrorHandler

	staticRewriteRules map[string][]*RewriteRule
	staticHeaders      map[string]http.Header

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
}

// ViewCache gets the view cache for the app.
func (a *App) ViewCache() *template.Template {
	return a.viewCache
}

// SetViewCache sets the view cache for the app.
func (a *App) SetViewCache(viewCache *template.Template) {
	a.viewCache = viewCache
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
	return a.StartWithServer(server)
}

// StartWithServer starts the app on a custom server.
// This lets you configure things like TLS keys and
// other options.
func (a *App) StartWithServer(server *http.Server) error {
	// this is the only property we will set of the server
	// i.e. the server handler (which is this app)
	server.Handler = a
	if a.logger != nil {
		a.logger.Logf("%s Started, listening on %s", a.Name(), server.Addr)
	}
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

// Static serves files from the given file system root.
// The path must end with "/*filepath", files are then served from the local
// path /defined/root/dir/*filepath.
// For example if root is "/etc" and *filepath is "passwd", the local file
// "/etc/passwd" would be served.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Router's NotFound handler.
// To use the operating system's file system implementation,
// use http.Dir:
//     router.ServeFiles("/src/*filepath", http.Dir("/var/www"))
func (a *App) Static(path string, root http.FileSystem) {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}

	fileServer := http.FileServer(root)

	a.router.GET(path, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		filePath := p.ByName("filepath")
		if rules, hasRules := a.staticRewriteRules[path]; hasRules {
			for _, rule := range rules {
				if matched, newFilePath := rule.Apply(filePath); matched {
					filePath = newFilePath
				}
			}
		}

		if headers, hasHeaders := a.staticHeaders[path]; hasHeaders {
			for key, values := range headers {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}
		}

		r.URL.Path = filePath

		w.Header().Set("Vary", "Accept-Encoding")
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gzw := NewGZippedResponseWriter(w)
			defer gzw.Close()

			context := a.RequestContext(gzw, r, parseParams(p))
			a.OnRequestStart(context)
			fileServer.ServeHTTP(gzw, r)
			a.OnRequestComplete(context)
			gzw.Flush()
			context.setStatusCode(gzw.StatusCode)
			context.setContentLength(gzw.BytesWritten)
			context.LogRequest()
		} else {
			rw := NewResponseWriter(w)
			context := a.RequestContext(rw, r, parseParams(p))
			a.OnRequestStart(context)
			fileServer.ServeHTTP(rw, r)
			context.setStatusCode(rw.StatusCode)
			context.setContentLength(rw.ContentLength)
			a.OnRequestComplete(context)
			context.LogRequest()
		}
	})
}

// StaticRewrite adds a rewrite rule for a specific statically served path.
// Make sure to serve the static path with (app).Static(path, root).
func (a *App) StaticRewrite(path, match string, action RewriteAction) error {
	expr, err := regexp.Compile(match)
	if err != nil {
		return err
	}
	a.staticRewriteRules[path] = append(a.staticRewriteRules[path], &RewriteRule{
		MatchExpression: match,
		expr:            expr,
		Action:          action,
	})

	return nil
}

// StaticHeader adds a header for the given static path.
func (a *App) StaticHeader(path, key, value string) {
	if _, hasHeaders := a.staticHeaders[path]; !hasHeaders {
		a.staticHeaders[path] = http.Header{}
	}
	a.staticHeaders[path].Add(key, value)
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
			a.OnRequestError(r, err)

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
	hc.api = NewAPIResultProvider(a, hc)
	hc.view = NewViewResultProvider(a, hc)
	return hc
}

// --------------------------------------------------------------------------------
// Events
// --------------------------------------------------------------------------------

// OnRequestStart triggers the onRequestStart event.
func (a *App) OnRequestStart(r *RequestContext) {
	if len(a.onRequestStart) > 0 {
		for _, handler := range a.onRequestStart {
			handler(r)
		}
	}
}

// OnRequestComplete triggers the onRequestStart event.
func (a *App) OnRequestComplete(r *RequestContext) {
	if len(a.onRequestComplete) > 0 {
		for _, handler := range a.onRequestComplete {
			handler(r)
		}
	}
}

// OnRequestError triggers the onRequestStart event.
func (a *App) OnRequestError(r *RequestContext, err interface{}) {
	if len(a.onRequestError) > 0 {
		for _, handler := range a.onRequestError {
			handler(r, err)
		}
	}
}

// RequestStartHandler fires before an action handler is run.
func (a *App) RequestStartHandler(handler RequestEventHandler) {
	a.onRequestStart = append(a.onRequestStart, handler)
}

// RequestCompleteHandler fires after an action handler is run.
func (a *App) RequestCompleteHandler(handler RequestEventHandler) {
	a.onRequestComplete = append(a.onRequestComplete, handler)
}

// RequestErrorHandler fires if there is an error logged.
func (a *App) RequestErrorHandler(handler RequestEventErrorHandler) {
	a.onRequestError = append(a.onRequestError, handler)
}

// --------------------------------------------------------------------------------
// Render Methods
// --------------------------------------------------------------------------------

// RenderAction is the translation step from APIControllerAction to httprouter.Handle.
func (a *App) RenderAction(action ControllerAction) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Set("X-Served-By", "github.com/wcharczuk/go-web")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("X-Xss-Protection", "1; mode=block")
		w.Header().Set("X-Content-Type-Options", "nosniff")
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
	rw := NewResponseWriter(w)
	context := a.RequestContext(rw, r, p)

	a.OnRequestStart(context)
	context.Render(action(context))
	context.setStatusCode(rw.StatusCode)
	context.setContentLength(rw.ContentLength)
	a.OnRequestComplete(context)
	context.LogRequest()
}

func (a *App) renderCompressed(action ControllerAction, w http.ResponseWriter, r *http.Request, p RouteParameters) {
	w.Header().Set("Content-Encoding", "gzip")

	gzw := NewGZippedResponseWriter(w)
	defer gzw.Close()

	context := a.RequestContext(gzw, r, p)

	a.OnRequestStart(context)
	result := action(context)
	context.Render(result)
	gzw.Flush()
	context.setStatusCode(gzw.StatusCode)
	context.setContentLength(gzw.BytesWritten)
	a.OnRequestComplete(context)
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
