package web

import (
	"database/sql"
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
		requestStartHandlers: []RequestEventHandler{func(r *RequestContext) {
			r.onRequestStart()
		}},
		requestCompleteHandlers: []RequestEventHandler{func(r *RequestContext) {
			r.onRequestEnd()
		}},
		requestErrorHandlers: []RequestEventErrorHandler{func(r *RequestContext, err interface{}) {
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

	requestStartHandlers    []RequestEventHandler
	requestCompleteHandlers []RequestEventHandler
	requestErrorHandlers    []RequestEventErrorHandler

	staticRewriteRules map[string][]*RewriteRule
	staticHeaders      map[string]http.Header

	port string
	tx   *sql.Tx
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

// SetTx sets the app's tx.
func (a *App) SetTx(tx *sql.Tx) {
	a.tx = tx
}

// Tx returns the app's transaction (usually used for testing.
func (a *App) Tx() *sql.Tx {
	return a.tx
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

func (a *App) marshalControllerMiddleware(action ControllerAction, middleware ...ControllerMiddleware) ControllerAction {
	if len(middleware) == 0 {
		return action
	}

	var nest = func(a, b ControllerMiddleware) ControllerMiddleware {
		if b == nil {
			return a
		}
		return func(action ControllerAction) ControllerAction {
			return a(b(action))
		}
	}

	var metaAction ControllerMiddleware
	for _, step := range middleware {
		metaAction = nest(step, metaAction)
	}
	return metaAction(action)
}

// GET registers a GET request handler.
func (a *App) GET(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.GET(path, a.RenderAction(a.marshalControllerMiddleware(action, middleware...)))
}

// OPTIONS registers a OPTIONS request handler.
func (a *App) OPTIONS(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.OPTIONS(path, a.RenderAction(a.marshalControllerMiddleware(action, middleware...)))
}

// HEAD registers a HEAD request handler.
func (a *App) HEAD(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.HEAD(path, a.RenderAction(a.marshalControllerMiddleware(action, middleware...)))
}

// PUT registers a PUT request handler.
func (a *App) PUT(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.PUT(path, a.RenderAction(a.marshalControllerMiddleware(action, middleware...)))
}

// POST registers a POST request actions.
func (a *App) POST(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.POST(path, a.RenderAction(a.marshalControllerMiddleware(action, middleware...)))
}

// DELETE registers a DELETE request handler.
func (a *App) DELETE(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.DELETE(path, a.RenderAction(a.marshalControllerMiddleware(action, middleware...)))
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
			a.onRequestStart(context)
			fileServer.ServeHTTP(gzw, r)
			a.onRequestComplete(context)
			gzw.Flush()
			context.setStatusCode(gzw.StatusCode)
			context.setContentLength(gzw.BytesWritten)
			context.LogRequest()
		} else {
			rw := NewResponseWriter(w)
			context := a.RequestContext(rw, r, parseParams(p))
			a.onRequestStart(context)
			fileServer.ServeHTTP(rw, r)
			context.setStatusCode(rw.StatusCode)
			context.setContentLength(rw.ContentLength)
			a.onRequestComplete(context)
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
			a.onRequestError(r, err)
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
func (a *App) onRequestStart(r *RequestContext) {
	if len(a.requestStartHandlers) > 0 {
		for _, handler := range a.requestStartHandlers {
			handler(r)
		}
	}
}

// OnRequestComplete triggers the onRequestStart event.
func (a *App) onRequestComplete(r *RequestContext) {
	if len(a.requestCompleteHandlers) > 0 {
		for _, handler := range a.requestCompleteHandlers {
			handler(r)
		}
	}
}

// OnRequestError triggers the onRequestStart event.
func (a *App) onRequestError(r *RequestContext, err interface{}) {
	if len(a.requestErrorHandlers) > 0 {
		for _, handler := range a.requestErrorHandlers {
			handler(r, err)
		}
	}
}

// RequestStartHandler fires before an action handler is run.
func (a *App) RequestStartHandler(handler RequestEventHandler) {
	a.requestStartHandlers = append(a.requestStartHandlers, handler)
}

// RequestCompleteHandler fires after an action handler is run.
func (a *App) RequestCompleteHandler(handler RequestEventHandler) {
	a.requestCompleteHandlers = append(a.requestCompleteHandlers, handler)
}

// RequestErrorHandler fires if there is an error logged.
func (a *App) RequestErrorHandler(handler RequestEventErrorHandler) {
	a.requestErrorHandlers = append(a.requestErrorHandlers, handler)
}

// --------------------------------------------------------------------------------
// Testing Methods
// --------------------------------------------------------------------------------

// Mock returns a request bulider to facilitate mocking requests.
func (a *App) Mock() *MockRequestBuilder {
	return NewMockRequestBuilder(a)
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

	a.onRequestStart(context)
	context.Render(action(context))
	context.setStatusCode(rw.StatusCode)
	context.setContentLength(rw.ContentLength)
	a.onRequestComplete(context)
	context.LogRequest()
}

func (a *App) renderCompressed(action ControllerAction, w http.ResponseWriter, r *http.Request, p RouteParameters) {
	w.Header().Set("Content-Encoding", "gzip")

	gzw := NewGZippedResponseWriter(w)
	defer gzw.Close()

	context := a.RequestContext(gzw, r, p)

	a.onRequestStart(context)
	result := action(context)
	context.Render(result)
	gzw.Flush()
	context.setStatusCode(gzw.StatusCode)
	context.setContentLength(gzw.BytesWritten)
	a.onRequestComplete(context)
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
