package web

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	logger "github.com/blendlabs/go-logger"
	"github.com/julienschmidt/httprouter"
)

// New returns a new app.
func New() *App {
	app := &App{
		router:             httprouter.New(),
		staticRewriteRules: map[string][]*RewriteRule{},
		staticHeaders:      map[string]http.Header{},
		tlsCertLock:        &sync.Mutex{},
	}
	app.SetDiagnostics(logger.NewDiagnosticsAgent(logger.NewEventFlagSetNone()))
	return app
}

// AppStartDelegate is a function that is run on start. Typically you use this to initialize the app.
type AppStartDelegate func(app *App) error

// App is the server for the app.
type App struct {
	name string

	diagnostics *logger.DiagnosticsAgent
	config      interface{}
	router      *httprouter.Router
	viewCache   *template.Template

	tlsCertBytes, tlsKeyBytes []byte
	tlsCertLock               *sync.Mutex
	tlsCert                   *tls.Certificate

	apiResultProvider  *APIResultProvider
	viewResultProvider *ViewResultProvider

	startDelegate AppStartDelegate

	staticRewriteRules map[string][]*RewriteRule
	staticHeaders      map[string]http.Header

	panicHandler PanicControllerAction

	tx *sql.Tx

	port string
}

// AppName returns the app name.
func (a *App) AppName() string {
	return a.diagnostics.Writer().Label()
}

// SetAppName sets a log label for the app.
func (a *App) SetAppName(appName string) {
	a.diagnostics.Writer().SetLabel(appName)
}

// UseTLS sets the app to use TLS.
func (a *App) UseTLS(tlsCert, tlsKey []byte) {
	a.tlsCertBytes = tlsCert
	a.tlsKeyBytes = tlsKey
}

// Diagnostics returns the diagnostics agent for the app.
func (a *App) Diagnostics() *logger.DiagnosticsAgent {
	return a.diagnostics
}

// SetDiagnostics sets the diagnostics agent.
func (a *App) SetDiagnostics(da *logger.DiagnosticsAgent) {
	a.diagnostics = da
	if a.diagnostics != nil {
		a.diagnostics.AddEventListener(logger.EventWebRequestStart, a.onRequestStart)
		a.diagnostics.AddEventListener(logger.EventWebRequest, a.onRequestComplete)
		a.diagnostics.AddEventListener(logger.EventWebResponse, a.onResponse)
	}
}

// Config returns the app config object.
func (a *App) Config() interface{} {
	return a.config
}

// SetConfig sets the app config object.
func (a *App) SetConfig(config interface{}) {
	a.config = config
}

// InitializeConfig reads a config prototype from the environment.
func (a *App) InitializeConfig(configPrototype interface{}) error {
	config, err := ReadConfigFromEnvironment(configPrototype)
	if err != nil {
		return err
	}
	a.config = config
	return nil
}

func (a *App) onRequestStart(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
	if len(state) < 1 {
		return
	}
	context, isContext := state[0].(*RequestContext)
	if !isContext {
		return
	}
	logger.WriteRequestStart(writer, ts, context.Request)
}

func (a *App) onRequestComplete(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
	if len(state) < 1 {
		return
	}
	context, isContext := state[0].(*RequestContext)
	if !isContext {
		return
	}
	logger.WriteRequest(writer, ts, context.Request, context.Response.StatusCode(), context.Response.ContentLength(), context.Elapsed())
}

func (a *App) onResponse(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
	if len(state) < 1 {
		return
	}
	body, stateIsBody := state[0].([]byte)
	if !stateIsBody {
		return
	}
	logger.WriteResponseBody(writer, ts, body)
}

// ViewCache gets the view cache for the app.
func (a *App) ViewCache() *template.Template {
	return a.viewCache
}

// SetViewCache sets the view cache for the app.
func (a *App) SetViewCache(viewCache *template.Template) {
	a.viewCache = viewCache
}

// IsolateTo sets the app to use a transaction for *all* requests.
// Caveat: only use during testing.
func (a *App) IsolateTo(tx *sql.Tx) {
	a.tx = tx
}

// Port returns the port for the app.
func (a *App) Port() string {
	if len(a.port) != 0 {
		return a.port
	}
	envVar := os.Getenv("PORT")
	if len(envVar) != 0 {
		return envVar
	}

	return "8080"
}

// SetPort sets the port the app listens on.
func (a *App) SetPort(port string) {
	a.port = port
}

// OnStart lets you register a task that is run before the server starts.
// Typically this delegate sets up the database connection and other init items.
func (a *App) OnStart(action AppStartDelegate) {
	a.startDelegate = action
}

// Start starts the server and binds to the given address.
func (a *App) Start() error {
	bindAddr := fmt.Sprintf(":%s", a.Port())
	server := &http.Server{
		Addr:    bindAddr,
		Handler: a,
	}

	return a.StartWithServer(server)
}

func (a *App) getCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if a.tlsCert == nil {
		tlsCert, err := tls.X509KeyPair(a.tlsCertBytes, a.tlsKeyBytes)
		if err != nil {
			return nil, err
		}
		a.tlsCert = &tlsCert
	}
	return a.tlsCert, nil
}

// StartWithServer starts the app on a custom server.
// This lets you configure things like TLS keys and
// other options.
func (a *App) StartWithServer(server *http.Server) error {
	if a.startDelegate != nil {
		a.diagnostics.Infof("Startup tasks starting")
		err := a.startDelegate(a)
		if err != nil {
			a.diagnostics.Errorf("Startup tasks error: %v", err)
			return err
		}
		a.diagnostics.Infof("Startup tasks complete")
	}

	// this is the only property we will set of the server
	// i.e. the server handler (which is this app)
	server.Handler = a
	a.diagnostics.Infof("Started, listening on %s", server.Addr)
	a.diagnostics.Infof("Diagnostics Verbosity %s", a.diagnostics.Events().String())

	if len(a.tlsCertBytes) > 0 && len(a.tlsKeyBytes) > 0 {
		server.TLSConfig = &tls.Config{
			GetCertificate: a.getCertificate,
		}

		return server.ListenAndServeTLS("", "")
	}

	return server.ListenAndServe()
}

// Register registers a controller with the app's router.
func (a *App) Register(c Controller) {
	c.Register(a)
}

// InitializeViewCache caches templates by path.
func (a *App) InitializeViewCache(paths ...string) error {
	views, err := template.ParseFiles(paths...)
	if err != nil {
		return err
	}
	a.viewCache = template.Must(views, nil)
	return nil
}

// GET registers a GET request handler.
func (a *App) GET(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.GET(path, a.renderAction(NestMiddleware(action, middleware...)))
}

// OPTIONS registers a OPTIONS request handler.
func (a *App) OPTIONS(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.OPTIONS(path, a.renderAction(NestMiddleware(action, middleware...)))
}

// HEAD registers a HEAD request handler.
func (a *App) HEAD(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.HEAD(path, a.renderAction(NestMiddleware(action, middleware...)))
}

// PUT registers a PUT request handler.
func (a *App) PUT(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.PUT(path, a.renderAction(NestMiddleware(action, middleware...)))
}

// PATCH registers a PATCH request handler.
func (a *App) PATCH(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.PATCH(path, a.renderAction(NestMiddleware(action, middleware...)))
}

// POST registers a POST request actions.
func (a *App) POST(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.POST(path, a.renderAction(NestMiddleware(action, middleware...)))
}

// DELETE registers a DELETE request handler.
func (a *App) DELETE(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.DELETE(path, a.renderAction(NestMiddleware(action, middleware...)))
}

// --------------------------------------------------------------------------------
// Static Result Methods
// --------------------------------------------------------------------------------

// AddStaticRewriteRule adds a rewrite rule for a specific statically served path.
// Make sure to serve the static path with (app).Static(path, root).
func (a *App) AddStaticRewriteRule(path, match string, action RewriteAction) error {
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

// AddStaticHeader adds a header for the given static path.
func (a *App) AddStaticHeader(path, key, value string) {
	if _, hasHeaders := a.staticHeaders[path]; !hasHeaders {
		a.staticHeaders[path] = http.Header{}
	}
	a.staticHeaders[path].Add(key, value)
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

	a.router.GET(path, a.renderAction(a.staticAction(path, root)))
}

// staticAction returns a ControllerAction for a given static path and root.
func (a *App) staticAction(path string, root http.FileSystem) ControllerAction {
	fileServer := http.FileServer(root)

	return func(r *RequestContext) ControllerResult {

		var staticRewriteRules []*RewriteRule
		var staticHeaders http.Header

		if rules, hasRules := a.staticRewriteRules[path]; hasRules {
			staticRewriteRules = rules
		}

		if headers, hasHeaders := a.staticHeaders[path]; hasHeaders {
			staticHeaders = headers
		}

		filePath, _ := r.RouteParam("filepath")

		return &StaticResult{
			FilePath:     filePath,
			FileServer:   fileServer,
			RewriteRules: staticRewriteRules,
			Headers:      staticHeaders,
		}
	}
}

// --------------------------------------------------------------------------------
// Router internal methods
// --------------------------------------------------------------------------------

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
	a.panicHandler = handler
	a.router.PanicHandler = func(w http.ResponseWriter, r *http.Request, err interface{}) {
		a.renderAction(func(r *RequestContext) ControllerResult {
			a.diagnostics.Fatal(err)
			return handler(r, err)
		})(w, r, httprouter.Params{})
	}
}

func (a *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	a.router.ServeHTTP(w, req)
}

// --------------------------------------------------------------------------------
// Testing Methods
// --------------------------------------------------------------------------------

// Mock returns a request bulider to facilitate mocking requests.
func (a *App) Mock() *MockRequestBuilder {
	return NewMockRequestBuilder(a)
}

// --------------------------------------------------------------------------------
// Request Pipeline
// --------------------------------------------------------------------------------

// renderAction is the translation step from ControllerAction to httprouter.Handle.
// this is where the bulk of the "pipeline" happens.
func (a *App) renderAction(action ControllerAction) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.setCommonResponseHeaders(w)
		response := a.newResponse(w, r)
		context := a.pipelineInit(response, r, NewRouteParameters(p))
		a.renderResult(action, context)
		a.pipelineComplete(context)
	}
}

func (a *App) setCommonResponseHeaders(w http.ResponseWriter) {
	w.Header().Set("Vary", "Accept-Encoding")
	w.Header().Set("X-Served-By", "github.com/wcharczuk/go-web")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("X-Xss-Protection", "1; mode=block")
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

func (a *App) newResponse(w http.ResponseWriter, r *http.Request) ResponseWriter {
	var response ResponseWriter
	if a.shouldCompressOutput(r) {
		w.Header().Set("Content-Encoding", "gzip")
		if a.diagnostics.IsEnabled(logger.EventWebResponse) {
			response = NewBufferedCompressedResponseWriter(w)
		} else {
			response = NewCompressedResponseWriter(w)
		}
	} else {
		if a.diagnostics.IsEnabled(logger.EventWebResponse) {
			response = NewBufferedResponseWriter(w)
		} else {
			response = NewResponseWriter(w)
		}
	}
	return response
}

func (a *App) shouldCompressOutput(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
}

func (a *App) pipelineInit(w ResponseWriter, r *http.Request, p RouteParameters) *RequestContext {
	context := a.requestContext(w, r, p)
	context.onRequestStart()
	a.diagnostics.OnEvent(logger.EventWebRequestStart, context)
	return context
}

// RequestContext creates an http context.
func (a *App) requestContext(w ResponseWriter, r *http.Request, p RouteParameters) *RequestContext {
	rc := NewRequestContext(w, r, p)
	rc.app = a
	rc.tx = a.tx
	rc.diagnostics = a.diagnostics
	rc.config = a.config
	return rc
}

func (a *App) renderResult(action ControllerAction, context *RequestContext) {
	result := action(context)
	if result != nil {
		err := result.Render(context.Response, context.Request)
		if err != nil {
			a.diagnostics.Error(err)
		}
	}
}

func (a *App) pipelineComplete(context *RequestContext) {
	err := context.Response.Flush()
	if err != nil && err != http.ErrBodyNotAllowed {
		a.diagnostics.Error(err)
	}
	context.onRequestEnd()
	context.setLoggedStatusCode(context.Response.StatusCode())
	context.setLoggedContentLength(context.Response.ContentLength())
	if a.diagnostics.IsEnabled(logger.EventWebResponse) {
		a.diagnostics.OnEvent(logger.EventWebResponse, context.Response.Bytes())
	}
	a.diagnostics.OnEvent(logger.EventWebRequest, context)

	err = context.Response.Close()
	if err != nil && err != http.ErrBodyNotAllowed {
		a.diagnostics.Error(err)
	}
}
