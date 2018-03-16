package web

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"net/url"

	exception "github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
	env "github.com/blendlabs/go-util/env"
)

// New returns a new app.
func New() *App {
	views := NewViewCache()
	vrp := &ViewResultProvider{views: views}
	return &App{
		auth:                  NewAuthManager(),
		state:                 map[string]interface{}{},
		views:                 views,
		statics:               map[string]Fileserver{},
		readTimeout:           DefaultReadTimeout,
		redirectTrailingSlash: true,
		recoverPanics:         true,

		defaultHeaders: DefaultHeaders,

		viewProvider: vrp,
		jsonProvider: &JSONResultProvider{},
		xmlProvider:  &XMLResultProvider{},
		textProvider: &TextResultProvider{},
	}
}

// NewFromEnv returns a new app from the environment.
func NewFromEnv() *App {
	return NewFromConfig(NewConfigFromEnv())
}

// NewFromConfig returns a new app from a given config.
func NewFromConfig(cfg *Config) *App {
	views, err := NewViewCacheFromConfig(&cfg.Views)
	if err != nil {
		return &App{err: err}
	}
	vrp := &ViewResultProvider{views: views}

	tlsConfig, err := cfg.TLS.GetConfig()
	if err != nil {
		return &App{err: err}
	}

	baseURL := cfg.GetBaseURL()
	var base *url.URL
	if len(baseURL) > 0 {
		base, err = url.Parse(baseURL)
		if err != nil {
			return &App{err: err}
		}
	}

	return &App{
		auth:                   NewAuthManagerFromConfig(cfg),
		state:                  map[string]interface{}{},
		views:                  views,
		statics:                map[string]Fileserver{},
		tlsConfig:              tlsConfig,
		redirectTrailingSlash:  cfg.GetRedirectTrailingSlash(),
		handleMethodNotAllowed: cfg.GetHandleMethodNotAllowed(),
		handleOptions:          cfg.GetHandleOptions(),
		recoverPanics:          cfg.GetRecoverPanics(),

		bindAddr: cfg.GetBindAddr(),
		baseURL:  base,

		defaultHeaders: cfg.GetDefaultHeaders(DefaultHeaders),

		hsts:                  cfg.GetHSTS(),
		hstsMaxAgeSeconds:     cfg.GetHSTSMaxAgeSeconds(),
		hstsIncludeSubdomains: cfg.GetHSTSIncludeSubDomains(),
		hstsPreload:           cfg.GetHSTSPreload(),

		maxHeaderBytes:    cfg.GetMaxHeaderBytes(),
		readHeaderTimeout: cfg.GetReadHeaderTimeout(),
		readTimeout:       cfg.GetReadTimeout(),
		writeTimeout:      cfg.GetWriteTimeout(),
		idleTimeout:       cfg.GetIdleTimeout(),

		viewProvider: vrp,
		jsonProvider: &JSONResultProvider{},
		xmlProvider:  &XMLResultProvider{},
		textProvider: &TextResultProvider{},
	}
}

// App is the server for the app.
type App struct {
	baseURL  *url.URL
	bindAddr string

	log   *logger.Logger
	auth  *AuthManager
	views *ViewCache

	hsts                  bool
	hstsMaxAgeSeconds     int
	hstsIncludeSubdomains bool
	hstsPreload           bool
	tlsConfig             *tls.Config

	defaultHeaders map[string]string

	startDelegate AppStartDelegate
	server        *http.Server

	// statics serve files at various routes
	statics map[string]Fileserver

	routes                  map[string]*node
	notFoundHandler         Handler
	methodNotAllowedHandler Handler
	panicHandler            PanicHandler
	panicAction             PanicAction
	redirectTrailingSlash   bool
	handleOptions           bool
	handleMethodNotAllowed  bool

	defaultMiddleware []Middleware

	viewProvider *ViewResultProvider
	jsonProvider *JSONResultProvider
	xmlProvider  *XMLResultProvider
	textProvider *TextResultProvider

	maxHeaderBytes    int
	readTimeout       time.Duration
	readHeaderTimeout time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration

	state map[string]interface{}

	recoverPanics bool

	err error
}

// Err returns an initialization error.
func (a *App) Err() error {
	return a.err
}

// State is a bag for common app state.
func (a *App) State() map[string]interface{} {
	return a.state
}

// WithDefaultHeaders sets the default headers
func (a *App) WithDefaultHeaders(headers map[string]string) *App {
	a.defaultHeaders = headers
	return a
}

// WithDefaultHeader adds a default header.
func (a *App) WithDefaultHeader(key string, value string) *App {
	if a.defaultHeaders == nil {
		a.defaultHeaders = map[string]string{}
	}
	a.defaultHeaders[key] = value
	return a
}

// DefaultHeaders returns the default headers.
func (a *App) DefaultHeaders() map[string]string {
	return a.defaultHeaders
}

// WithState sets app state and returns a reference to the app for building apps with a fluent api.
func (a *App) WithState(key string, value interface{}) *App {
	a.state[key] = value
	return a
}

// SetState sets app state.
func (a *App) SetState(key string, value interface{}) {
	a.state[key] = value
}

// HandleMethodNotAllowed returns if we should handle unhandled verbs.
func (a *App) HandleMethodNotAllowed() bool {
	return a.handleMethodNotAllowed
}

// WithHandleMethodNotAllowed sets if we should handlem ethod not allowed.
func (a *App) WithHandleMethodNotAllowed(handle bool) *App {
	a.handleMethodNotAllowed = handle
	return a
}

// HandleOptions returns if we should handle OPTIONS requests.
func (a *App) HandleOptions() bool {
	return a.handleOptions
}

// WithHandleOptions returns if we should handle OPTIONS requests.
func (a *App) WithHandleOptions(handle bool) *App {
	a.handleOptions = handle
	return a
}

// RecoverPanics returns if the app recovers panics.
func (a *App) RecoverPanics() bool {
	return a.recoverPanics
}

// WithRecoverPanics sets if the app should recover panics.
func (a *App) WithRecoverPanics(value bool) *App {
	a.recoverPanics = value
	return a
}

// BaseURL returns the domain for the app.
func (a *App) BaseURL() *url.URL {
	return a.baseURL
}

// WithBaseURL sets the `BaseURL` field and returns a reference to the app for building apps with a fluent api.
func (a *App) WithBaseURL(baseURL string) *App {
	if err := a.SetBaseURL(baseURL); err != nil {
		a.err = err
	}
	return a
}

// SetBaseURL sets the base url for the app.
func (a *App) SetBaseURL(baseURL string) error {
	u, err := url.Parse(baseURL)
	if err != nil {
		return exception.Wrap(err)
	}
	a.baseURL = u
	return nil
}

// MaxHeaderBytes returns the app max header bytes.
func (a *App) MaxHeaderBytes() int {
	return a.maxHeaderBytes
}

// WithMaxHeaderBytes sets the max header bytes value and returns a reference.
func (a *App) WithMaxHeaderBytes(byteCount int) *App {
	a.maxHeaderBytes = byteCount
	return a
}

// ReadHeaderTimeout returns the read header timeout for the server.
func (a *App) ReadHeaderTimeout() time.Duration {
	return a.readHeaderTimeout
}

// WithReadHeaderTimeout returns the read header timeout for the server.
func (a *App) WithReadHeaderTimeout(timeout time.Duration) *App {
	a.readHeaderTimeout = timeout
	return a
}

// ReadTimeout returns the read timeout for the server.
func (a *App) ReadTimeout() time.Duration {
	return a.readTimeout
}

// WithReadTimeout sets the read timeout for the server and returns a reference to the app for building apps with a fluent api.
func (a *App) WithReadTimeout(timeout time.Duration) *App {
	a.readTimeout = timeout
	return a
}

// IdleTimeout is the time before we close a connection.
func (a *App) IdleTimeout() time.Duration {
	return a.idleTimeout
}

// WithIdleTimeout sets the idle timeout.
func (a *App) WithIdleTimeout(timeout time.Duration) *App {
	a.idleTimeout = timeout
	return a
}

// WriteTimeout returns the write timeout for the server.
func (a *App) WriteTimeout() time.Duration {
	return a.writeTimeout
}

// WithWriteTimeout sets the write timeout for the server and returns a reference to the app for building apps with a fluent api.
func (a *App) WithWriteTimeout(timeout time.Duration) *App {
	a.writeTimeout = timeout
	return a
}

// WithHSTS enables or disables issuing the strict transport security header.
func (a *App) WithHSTS(enabled bool) *App {
	a.hsts = enabled
	return a
}

// HSTS returns if strict transport security is enabled.
func (a *App) HSTS() bool {
	return a.hsts
}

// WithHSTSMaxAgeSeconds sets the hsts max age seconds.
func (a *App) WithHSTSMaxAgeSeconds(ageSeconds int) *App {
	a.hstsMaxAgeSeconds = ageSeconds
	return a
}

// HSTSMaxAgeSeconds is the maximum lifetime browsers should honor the secure transport header.
func (a *App) HSTSMaxAgeSeconds() int {
	return a.hstsMaxAgeSeconds
}

// WithHSTSIncludeSubdomains sets if we should include subdomains in hsts.
func (a *App) WithHSTSIncludeSubdomains(includeSubdomains bool) *App {
	a.hstsIncludeSubdomains = includeSubdomains
	return a
}

// HSTSIncludeSubdomains returns if we should include subdomains in hsts.
func (a *App) HSTSIncludeSubdomains() bool {
	return a.hstsIncludeSubdomains
}

// WithHSTSPreload sets if we preload hsts.
func (a *App) WithHSTSPreload(preload bool) *App {
	a.hstsPreload = preload
	return a
}

// HSTSPreload returns if we should preload hsts.
func (a *App) HSTSPreload() bool {
	return a.hstsPreload
}

// WithTLSConfig sets the tls config for the app.
func (a *App) WithTLSConfig(config *tls.Config) *App {
	a.SetTLSConfig(config)
	return a
}

// SetTLSConfig sets the tls config.
func (a *App) SetTLSConfig(config *tls.Config) {
	a.tlsConfig = config
}

// TLSConfig returns the app tls config.
func (a *App) TLSConfig() *tls.Config {
	return a.tlsConfig
}

// WithTLSCertPair sets the app to use TLS when listening, and returns a reference to the app for building apps with a fluent api.
func (a *App) WithTLSCertPair(tlsCert, tlsKey []byte) *App {
	if err := a.SetTLSCertPair(tlsCert, tlsKey); err != nil {
		a.err = err
	}
	return a
}

// SetTLSCertPair sets the app to use TLS with a given cert.
func (a *App) SetTLSCertPair(tlsCert, tlsKey []byte) error {
	cert, err := tls.X509KeyPair(tlsCert, tlsKey)
	if err != nil {
		return err
	}
	if a.tlsConfig == nil {
		a.tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	} else {
		a.tlsConfig.Certificates = []tls.Certificate{cert}
	}
	return nil
}

// WithTLSCertPairFromFiles sets the tls key pair from a given set of paths to files, and returns a reference to the app.
func (a *App) WithTLSCertPairFromFiles(tlsCertPath, tlsKeyPath string) *App {
	a.err = a.SetTLSCertPairFromFiles(tlsCertPath, tlsKeyPath)
	return a
}

// SetTLSCertPairFromFiles reads a tls key pair from a given set of paths.
func (a *App) SetTLSCertPairFromFiles(tlsCertPath, tlsKeyPath string) error {
	cert, err := ioutil.ReadFile(tlsCertPath)
	if err != nil {
		return exception.Wrap(err)
	}

	key, err := ioutil.ReadFile(tlsKeyPath)
	if err != nil {
		return exception.Wrap(err)
	}

	return a.SetTLSCertPair(cert, key)
}

// WithTLSFromEnv reads TLS settings from the environment, and returns a reference to the app for building apps with a fluent api.
func (a *App) WithTLSFromEnv() *App {
	if err := a.SetTLSFromEnv(); err != nil {
		a.err = err
	}
	return a
}

// SetTLSFromEnv reads TLS settings from the environment.
func (a *App) SetTLSFromEnv() error {
	tlsCert := env.Env().Bytes(EnvironmentVariableTLSCert)
	tlsKey := env.Env().Bytes(EnvironmentVariableTLSKey)
	tlsCertPath := env.Env().String(EnvironmentVariableTLSCertFile)
	tlsKeyPath := env.Env().String(EnvironmentVariableTLSKeyFile)

	if len(tlsCert) > 0 && len(tlsKey) > 0 {
		return a.SetTLSCertPair(tlsCert, tlsKey)
	} else if len(tlsCertPath) > 0 && len(tlsKeyPath) > 0 {
		return a.SetTLSCertPairFromFiles(tlsCertPath, tlsKeyPath)
	}
	return nil
}

// WithTLSClientCertPool sets the client cert pool and returns a reference to the app.
func (a *App) WithTLSClientCertPool(certs ...[]byte) *App {
	if err := a.SetTLSClientCertPool(certs...); err != nil {
		a.err = err
	}
	return a
}

// SetTLSClientCertPool set the client cert pool from a given set of pems.
func (a *App) SetTLSClientCertPool(certs ...[]byte) error {
	if a.tlsConfig == nil {
		a.tlsConfig = &tls.Config{}
	}
	a.tlsConfig.ClientCAs = x509.NewCertPool()
	for _, cert := range certs {
		ok := a.tlsConfig.ClientCAs.AppendCertsFromPEM(cert)
		if !ok {
			return exception.New("invalid ca cert for client cert pool")
		}
	}
	a.tlsConfig.BuildNameToCertificate()
	// This is a solution to enforce the server fetch the new config when a new
	// request comes in. The server would use the old ClientCAs pool if this is
	// not called.
	a.tlsConfig.GetConfigForClient = func(_ *tls.ClientHelloInfo) (*tls.Config, error) {
		return a.tlsConfig, nil
	}
	return nil
}

// WithTLSClientCertVerification sets the verification level for client certs.
func (a *App) WithTLSClientCertVerification(verification tls.ClientAuthType) *App {
	if a.tlsConfig == nil {
		a.tlsConfig = &tls.Config{}
	}
	a.tlsConfig.ClientAuth = verification
	return a
}

// Logger returns the diagnostics agent for the app.
func (a *App) Logger() *logger.Logger {
	return a.log
}

// WithLogger sets the app logger agent and returns a reference to the app.
// It also sets underlying loggers in any child resources like providers and the auth manager.
func (a *App) WithLogger(log *logger.Logger) *App {
	a.log = log
	if a.viewProvider != nil {
		a.viewProvider.log = log
	}
	if a.jsonProvider != nil {
		a.jsonProvider.log = log
	}
	if a.xmlProvider != nil {
		a.xmlProvider.log = log
	}
	if a.textProvider != nil {
		a.textProvider.log = log
	}
	if a.auth != nil {
		a.auth.log = log
	}
	return a
}

// WithAuth sets the auth manager.
func (a *App) WithAuth(am *AuthManager) *App {
	a.auth = am
	return a
}

// Auth returns the session manager.
func (a *App) Auth() *AuthManager {
	return a.auth
}

// WithPort sets the port for the bind address of the app, and returns a reference to the app.
func (a *App) WithPort(port int32) *App {
	a.SetPort(port)
	return a
}

// SetPort sets the port the app listens on, typically to `:%d` which indicates listen on any interface.
func (a *App) SetPort(port int32) {
	a.bindAddr = fmt.Sprintf(":%v", port)
}

// WithPortFromEnv sets the port from an environment variable, and returns a reference to the app.
func (a *App) WithPortFromEnv() *App {
	a.SetPortFromEnv()
	return a
}

// SetPortFromEnv sets the port from an environment variable, and returns a reference to the app.
func (a *App) SetPortFromEnv() {
	if env.Env().Has(EnvironmentVariablePort) {
		port, err := env.Env().Int32(EnvironmentVariablePort)
		if err != nil {
			a.err = err
		}
		a.bindAddr = fmt.Sprintf(":%v", port)
	}
}

// BindAddr returns the address the server will bind to.
func (a *App) BindAddr() string {
	return a.bindAddr
}

// WithBindAddr sets the address the app listens on, and returns a reference to the app.
func (a *App) WithBindAddr(bindAddr string) *App {
	a.bindAddr = bindAddr
	return a
}

// WithBindAddrFromEnv sets the address the app listens on, and returns a reference to the app.
func (a *App) WithBindAddrFromEnv() *App {
	a.bindAddr = env.Env().String(EnvironmentVariableBindAddr)
	return a
}

// WithDefaultMiddleware sets the application wide default middleware.
func (a *App) WithDefaultMiddleware(middleware ...Middleware) *App {
	a.defaultMiddleware = middleware
	return a
}

// DefaultMiddleware returns the default middleware.
func (a *App) DefaultMiddleware() []Middleware {
	return a.defaultMiddleware
}

// OnStart lets you register a task that is run before the server starts.
// Typically this delegate sets up the database connection and other init items.
func (a *App) OnStart(action AppStartDelegate) {
	a.startDelegate = action
}

// Server returns the basic http.Server for the app.
func (a *App) Server() *http.Server {
	return &http.Server{
		Addr:              a.BindAddr(),
		Handler:           a,
		MaxHeaderBytes:    a.maxHeaderBytes,
		ReadTimeout:       a.readTimeout,
		ReadHeaderTimeout: a.readHeaderTimeout,
		WriteTimeout:      a.writeTimeout,
		IdleTimeout:       a.idleTimeout,
		TLSConfig:         a.tlsConfig,
	}
}

// Start starts the server and binds to the given address.
func (a *App) Start() error {
	return a.StartWithServer(a.Server())
}

// StartWithServer starts the app on a custom server.
// This lets you configure things like TLS keys and
// other options.
func (a *App) StartWithServer(server *http.Server) (err error) {
	start := time.Now()
	if a.log != nil {
		a.log.Trigger(NewAppStartEvent(a))
		defer a.log.Trigger(NewAppExitEvent(a, err))
	}

	// early exit if we already had an issue.
	if a.err != nil {
		err = a.err
		return
	}

	if a.startDelegate != nil {
		err = a.startDelegate(a)
		if err != nil {
			return
		}
	}

	err = a.commonStartupTasks()
	if err != nil {
		return
	}

	serverProtocol := "http"
	if server.TLSConfig != nil {
		serverProtocol = "https (tls)"
	}

	a.syncInfof("%s server started, listening on %s", serverProtocol, server.Addr)
	if a.log != nil {
		if a.log.Flags() != nil {
			a.syncInfof("%s server logging flags %s", serverProtocol, a.log.Flags().String())
		}
		a.log.Trigger(NewAppStartCompleteEvent(a, time.Since(start), err))
	}

	if server.TLSConfig != nil && server.TLSConfig.ClientCAs != nil {
		a.syncInfof("%s using client cert pool with (%d) client certs", serverProtocol, len(a.tlsConfig.ClientCAs.Subjects()))
	}

	if server.TLSConfig != nil {
		err = exception.Wrap(server.ListenAndServeTLS("", ""))
		return
	}

	a.server = server
	err = exception.Wrap(server.ListenAndServe())
	return
}

// Shutdown stops the server.
func (a *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	serverProtocol := "http"
	if a.server.TLSConfig != nil {
		serverProtocol = "https (tls)"
	}

	a.syncInfof("%s server shutting down", serverProtocol)
	a.server.SetKeepAlivesEnabled(false)
	return exception.Wrap(a.server.Shutdown(ctx))
}

// WithControllers registers given controllers and returns a reference to the app.
func (a *App) WithControllers(controllers ...Controller) *App {
	for _, c := range controllers {
		a.Register(c)
	}
	return a
}

// Register registers a controller with the app's router.
func (a *App) Register(c Controller) {
	c.Register(a)
}

// --------------------------------------------------------------------------------
// Route Registration / HTTP Methods
// --------------------------------------------------------------------------------

// GET registers a GET request handler.
func (a *App) GET(path string, action Action, middleware ...Middleware) {
	a.Handle("GET", path, a.renderAction(a.middlewarePipeline(action, middleware...)))
}

// OPTIONS registers a OPTIONS request handler.
func (a *App) OPTIONS(path string, action Action, middleware ...Middleware) {
	a.Handle("OPTIONS", path, a.renderAction(a.middlewarePipeline(action, middleware...)))
}

// HEAD registers a HEAD request handler.
func (a *App) HEAD(path string, action Action, middleware ...Middleware) {
	a.Handle("HEAD", path, a.renderAction(a.middlewarePipeline(action, middleware...)))
}

// PUT registers a PUT request handler.
func (a *App) PUT(path string, action Action, middleware ...Middleware) {
	a.Handle("PUT", path, a.renderAction(a.middlewarePipeline(action, middleware...)))
}

// PATCH registers a PATCH request handler.
func (a *App) PATCH(path string, action Action, middleware ...Middleware) {
	a.Handle("PATCH", path, a.renderAction(a.middlewarePipeline(action, middleware...)))
}

// POST registers a POST request actions.
func (a *App) POST(path string, action Action, middleware ...Middleware) {
	a.Handle("POST", path, a.renderAction(a.middlewarePipeline(action, middleware...)))
}

// DELETE registers a DELETE request handler.
func (a *App) DELETE(path string, action Action, middleware ...Middleware) {
	a.Handle("DELETE", path, a.renderAction(a.middlewarePipeline(action, middleware...)))
}

// Handle adds a raw handler at a given method and path.
func (a *App) Handle(method, path string, handler Handler) {
	if len(path) == 0 {
		panic("path must not be empty")
	}
	if path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}
	if a.routes == nil {
		a.routes = make(map[string]*node)
	}

	root := a.routes[method]
	if root == nil {
		root = new(node)
		a.routes[method] = root
	}

	root.addRoute(method, path, handler)
}

// Lookup finds the route data for a given method and path.
func (a *App) Lookup(method, path string) (route *Route, params RouteParameters, slashRedirect bool) {
	if root := a.routes[method]; root != nil {
		return root.getValue(path)
	}
	return nil, nil, false
}

// ServeHTTP makes the router implement the http.Handler interface.
func (a *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer a.recover(w, req)

	path := req.URL.Path

	if root := a.routes[req.Method]; root != nil {
		if route, params, tsr := root.getValue(path); route != nil {
			route.Handler(w, req, route, params, nil)
			return
		} else if req.Method != "CONNECT" && path != "/" {
			code := 301 // Permanent redirect, request with GET method
			if req.Method != "GET" {
				code = 307
			}

			if tsr && a.redirectTrailingSlash {
				if len(path) > 1 && path[len(path)-1] == '/' {
					req.URL.Path = path[:len(path)-1]
				} else {
					req.URL.Path = path + "/"
				}
				http.Redirect(w, req, req.URL.String(), code)
				return
			}
		}
	}

	if req.Method == "OPTIONS" {
		// Handle OPTIONS requests
		if a.handleOptions {
			if allow := a.allowed(path, req.Method); len(allow) > 0 {
				w.Header().Set("Allow", allow)
				return
			}
		}
	} else {
		// Handle 405
		if a.handleMethodNotAllowed {
			if allow := a.allowed(path, req.Method); len(allow) > 0 {
				w.Header().Set("Allow", allow)
				if a.methodNotAllowedHandler != nil {
					a.methodNotAllowedHandler(w, req, nil, nil, nil)
				} else {
					http.Error(w,
						http.StatusText(http.StatusMethodNotAllowed),
						http.StatusMethodNotAllowed,
					)
				}
				return
			}
		}
	}

	// Handle 404
	if a.notFoundHandler != nil {
		a.notFoundHandler(w, req, nil, nil, nil)
	} else {
		http.NotFound(w, req)
	}
}

// --------------------------------------------------------------------------------
// Views
// --------------------------------------------------------------------------------

// Views returns the view cache.
func (a *App) Views() *ViewCache {
	return a.views
}

// --------------------------------------------------------------------------------
// Static Result Methods
// --------------------------------------------------------------------------------

// WithStaticRewriteRule adds a rewrite rule for a specific statically served path.
// It mutates the path for the incoming static file request to the fileserver according to the action.
func (a *App) WithStaticRewriteRule(route, match string, action RewriteAction) error {
	mountedRoute := a.createStaticMountRoute(route)
	if static, hasRoute := a.statics[mountedRoute]; hasRoute {
		return static.AddRewriteRule(mountedRoute, match, action)
	}
	return exception.Newf("no static fileserver mounted at route").WithMessagef("route: %s", route)
}

// WithStaticHeader adds a header for the given static path.
// These headers are automatically added to any result that the static path fileserver sends.
func (a *App) WithStaticHeader(route, key, value string) error {
	mountedRoute := a.createStaticMountRoute(route)
	if static, hasRoute := a.statics[mountedRoute]; hasRoute {
		return static.AddHeader(key, value)
	}
	return exception.Newf("no static fileserver mounted at route").WithMessagef("route: %s", mountedRoute)
}

// ServeStatic serves files from the given file system root.
// If the path does not end with "/*filepath" that suffix will be added for you internally.
// For example if root is "/etc" and *filepath is "passwd", the local file
// "/etc/passwd" would be served.
func (a *App) ServeStatic(route, filepath string) {
	sfs := NewStaticFileServer(http.Dir(filepath))
	mountedRoute := a.createStaticMountRoute(route)
	a.statics[mountedRoute] = sfs
	a.Handle("GET", mountedRoute, a.renderAction(a.middlewarePipeline(sfs.Action)))
}

// ServeStaticCached serves files from the given file system root.
// If the path does not end with "/*filepath" that suffix will be added for you internally.
func (a *App) ServeStaticCached(route, filepath string) {
	sfs := NewCachedStaticFileServer(http.Dir(filepath))
	mountedRoute := a.createStaticMountRoute(route)
	a.statics[mountedRoute] = sfs
	a.Handle("GET", mountedRoute, a.renderAction(a.middlewarePipeline(sfs.Action)))
}

func (a *App) createStaticMountRoute(route string) string {
	mountedRoute := route
	if !strings.HasSuffix(mountedRoute, "*filepath") {
		if strings.HasSuffix(mountedRoute, "/") {
			mountedRoute = mountedRoute + "*filepath"
		} else {
			mountedRoute = mountedRoute + "/*filepath"
		}
	}
	return mountedRoute
}

// --------------------------------------------------------------------------------
// Router internal methods
// --------------------------------------------------------------------------------

// WithNotFoundHandler sets the not found handler.
func (a *App) WithNotFoundHandler(handler Action) *App {
	a.notFoundHandler = a.renderAction(handler)
	return a
}

// WithMethodNotAllowedHandler sets the not allowed handler.
func (a *App) WithMethodNotAllowedHandler(handler Action) *App {
	a.methodNotAllowedHandler = a.renderAction(handler)
	return a
}

// WithPanicAction sets the panic action.
func (a *App) WithPanicAction(action PanicAction) *App {
	a.panicAction = action
	return a
}

// --------------------------------------------------------------------------------
// Testing Methods
// --------------------------------------------------------------------------------

// Mock returns a request bulider to facilitate mocking requests.
func (a *App) Mock() *MockRequestBuilder {
	return NewMockRequestBuilder(a)
}

// --------------------------------------------------------------------------------
// App Lifecycle
// --------------------------------------------------------------------------------

func (a *App) commonStartupTasks() error {
	return a.views.Initialize()
}

// --------------------------------------------------------------------------------
// Request Pipeline
// --------------------------------------------------------------------------------

// renderAction is the translation step from Action to Handler.
// this is where the bulk of the "pipeline" happens.
func (a *App) renderAction(action Action) Handler {
	return func(w http.ResponseWriter, r *http.Request, route *Route, p RouteParameters, state State) {
		var err error

		for key, value := range a.defaultHeaders {
			w.Header().Set(key, value)
		}

		if a.hsts {
			a.addHSTSHeader(w)
		}

		var response ResponseWriter
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set(HeaderContentEncoding, ContentEncodingGZIP)
			response = NewCompressedResponseWriter(w)
		} else {
			w.Header().Set(HeaderContentEncoding, ContentEncodingIdentity)
			response = NewRawResponseWriter(w)
		}
		context := a.createCtx(response, r, route, p, state)
		context.onRequestStart()
		if a.log != nil {
			a.log.Trigger(a.loggerRequestStartEvent(context))
		}

		result := action(context)
		if result != nil {
			err = result.Render(context)
			if err != nil {
				a.logError(err)
			}
		}

		context.onRequestEnd()
		context.setLoggedStatusCode(response.StatusCode())
		context.setLoggedContentLength(response.ContentLength())

		err = response.Close()
		if err != nil && err != http.ErrBodyNotAllowed {
			a.logError(err)
		}

		// call the cancel func if it's set.
		if context.cancel != nil {
			context.cancel()
		}

		// effectively "request complete"
		if a.log != nil {
			a.log.Trigger(a.loggerRequestEvent(context))
		}
	}
}

func (a *App) addHSTSHeader(w http.ResponseWriter) {
	parts := []string{fmt.Sprintf(HSTSMaxAgeFormat, a.hstsMaxAgeSeconds)}
	if a.hstsIncludeSubdomains {
		parts = append(parts, HSTSIncludeSubDomains)
	}
	if a.hstsPreload {
		parts = append(parts, HSTSPreload)
	}
	w.Header().Set(HeaderStrictTransportSecurity, strings.Join(parts, "; "))
}

func (a *App) loggerRequestStartEvent(ctx *Ctx) *logger.WebRequestEvent {
	event := logger.NewWebRequestStartEvent(ctx.Request).
		WithState(ctx.state)

	if ctx.Route() != nil {
		event = event.WithRoute(ctx.Route().String())
	}
	return event
}

func (a *App) loggerRequestEvent(ctx *Ctx) *logger.WebRequestEvent {
	event := logger.NewWebRequestEvent(ctx.Request).
		WithStatusCode(ctx.statusCode).
		WithElapsed(ctx.Elapsed()).
		WithContentLength(int64(ctx.contentLength)).
		WithState(ctx.state)

	if ctx.Route() != nil {
		event = event.WithRoute(ctx.Route().String())
	}

	if ctx.Response.Header() != nil {
		event = event.WithContentType(ctx.Response.Header().Get(HeaderContentType))
		event = event.WithContentEncoding(ctx.Response.Header().Get(HeaderContentEncoding))
	}
	return event
}

func (a *App) recover(w http.ResponseWriter, req *http.Request) {
	if a.recoverPanics {
		if rcv := recover(); rcv != nil {
			if a.panicAction != nil {
				a.handlePanic(w, req, rcv)
			} else if a.log != nil {
				a.log.Fatalf("%v", rcv)
			}
		}
	}
}

func (a *App) handlePanic(w http.ResponseWriter, r *http.Request, err interface{}) {
	a.renderAction(func(ctx *Ctx) Result {
		if a.log != nil {
			a.log.Fatalf("%v", err)
		}
		return a.panicAction(ctx, err)
	})(w, r, nil, nil, nil)
}

func (a *App) createCtx(w ResponseWriter, r *http.Request, route *Route, p RouteParameters, s State) *Ctx {
	ctx := &Ctx{
		Response:        w,
		Request:         r,
		app:             a,
		route:           route,
		routeParameters: p,
		state:           s,
		auth:            a.auth,
		log:             a.log,
		view:            a.viewProvider,
		json:            a.jsonProvider,
		xml:             a.xmlProvider,
		text:            a.textProvider,
		defaultResultProvider: a.textProvider,
	}
	if ctx.state == nil {
		ctx.state = State{}
	}
	if a.state != nil && len(a.state) > 0 {
		for key, value := range a.state {
			ctx.state[key] = value
		}
	}

	return ctx
}

func (a *App) middlewarePipeline(action Action, middleware ...Middleware) Action {
	if len(middleware) == 0 && len(a.defaultMiddleware) == 0 {
		return action
	}

	finalMiddleware := make([]Middleware, len(middleware)+len(a.defaultMiddleware))
	cursor := len(finalMiddleware) - 1
	for i := len(a.defaultMiddleware) - 1; i >= 0; i-- {
		finalMiddleware[cursor] = a.defaultMiddleware[i]
		cursor--
	}

	for i := len(middleware) - 1; i >= 0; i-- {
		finalMiddleware[cursor] = middleware[i]
		cursor--
	}

	return NestMiddleware(action, finalMiddleware...)
}

func (a *App) allowed(path, reqMethod string) (allow string) {
	if path == "*" { // server-wide
		for method := range a.routes {
			if method == "OPTIONS" {
				continue
			}

			// add request method to list of allowed methods
			if len(allow) == 0 {
				allow = method
			} else {
				allow += ", " + method
			}
		}
		return
	}
	for method := range a.routes {
		// Skip the requested method - we already tried this one
		if method == reqMethod || method == "OPTIONS" {
			continue
		}

		handle, _, _ := a.routes[method].getValue(path)
		if handle != nil {
			// add request method to list of allowed methods
			if len(allow) == 0 {
				allow = method
			} else {
				allow += ", " + method
			}
		}
	}
	if len(allow) > 0 {
		allow += ", OPTIONS"
	}
	return
}

func (a *App) logError(err error) {
	if a.log == nil {
		return
	}

	a.log.Error(err)
}

func (a *App) syncInfof(format string, args ...interface{}) {
	if a.log == nil {
		return
	}
	a.log.SyncInfof(format, args...)
}

func (a *App) syncFatalf(format string, args ...interface{}) {
	if a.log == nil {
		return
	}
	a.log.SyncFatalf(format, args...)
}
