package web

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/blend/go-sdk/env"
	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/logger"
)

const (
	// VarzStarted is a common variable.
	VarzStarted = "startedUTC"
	// VarzRequests is a common variable.
	VarzRequests = "http_requests"
	// VarzRequests2xx is a common variable.
	VarzRequests2xx = "http_requests2xx"
	// VarzRequests3xx is a common variable.
	VarzRequests3xx = "http_requests3xx"
	// VarzRequests4xx is a common variable.
	VarzRequests4xx = "http_requests4xx"
	// VarzRequests5xx is a common variable.
	VarzRequests5xx = "http_requests5xx"
	// VarzErrors is a common variable.
	VarzErrors = "errors_total"
	// VarzFatals is a common variable.
	VarzFatals = "fatals_total"

	// ListenerHealthz is the uid of the healthz logger listeners.
	ListenerHealthz = "healthz"

	// ErrHealthzAppUnset is a common error.
	ErrHealthzAppUnset Error = "healthz app unset"
)

// NewHealthz returns a new healthz.
func NewHealthz(app *App) *Healthz {
	return &Healthz{
		app:            app,
		defaultHeaders: map[string]string{},
		state:          State{},
		vars: map[string]interface{}{
			VarzRequests:    int64(0),
			VarzRequests2xx: int64(0),
			VarzRequests3xx: int64(0),
			VarzRequests4xx: int64(0),
			VarzRequests5xx: int64(0),
			VarzErrors:      int64(0),
			VarzFatals:      int64(0),
		},
	}
}

// NewHealthzFromEnv returns a new healthz from the env.
func NewHealthzFromEnv(app *App) *Healthz {
	return NewHealthzFromConfig(app, NewHealthzConfigFromEnv())
}

// NewHealthzFromConfig returns a new healthz sidecar from a config.
func NewHealthzFromConfig(app *App, cfg *HealthzConfig) *Healthz {
	hz := NewHealthz(app)
	hz = hz.WithBindAddr(cfg.GetBindAddr())
	hz = hz.WithRecoverPanics(cfg.GetRecoverPanics())
	hz = hz.WithMaxHeaderBytes(cfg.GetMaxHeaderBytes())
	hz = hz.WithReadHeaderTimeout(cfg.GetReadHeaderTimeout())
	hz = hz.WithReadTimeout(cfg.GetReadTimeout())
	hz = hz.WithWriteTimeout(cfg.GetWriteTimeout())
	hz = hz.WithIdleTimeout(cfg.GetIdleTimeout())
	return hz
}

// HealthzHost hosts an app with a healthz, starting both servers.
func HealthzHost(app *App, hz *Healthz) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = exception.Newf("%v", r)
			return
		}
	}()

	appQuit := make(chan struct{})
	hzQuit := make(chan struct{})
	go func() {
		err = app.Start()
		close(appQuit)
	}()

	go func() {
		err = hz.Start()
		close(hzQuit)
	}()
	select {
	case <-appQuit:
		return
	case <-hzQuit:
		return
	}
}

// Healthz is a sentinel / healthcheck sidecar that can run on a different
// port to the main app.
// It typically implements the following routes:
// 	/healthz - overall health endpoint, 200 on healthy, 5xx on not.
// 	/varz    - basic stats and metrics since start
//	/debug/vars - `pkg/expvar` output.
type Healthz struct {
	app        *App
	startedUTC time.Time
	bindAddr   string
	log        *logger.Logger

	defaultHeaders map[string]string
	server         *http.Server
	listener       *net.TCPListener

	maxHeaderBytes    int
	readTimeout       time.Duration
	readHeaderTimeout time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration

	state State

	varsLock sync.Mutex
	vars     map[string]interface{}

	recoverPanics bool
	err           error
}

// App returns the underlying app.
func (hz *Healthz) App() *App {
	return hz.app
}

// Vars returns the underlying vars collection.
func (hz *Healthz) Vars() State {
	return hz.vars
}

// Server returns the underlying server.
func (hz *Healthz) Server() *http.Server {
	return hz.server
}

// Listener returns the underlying listener.
func (hz *Healthz) Listener() *net.TCPListener {
	return hz.listener
}

// WithServer sets the underlying server.
func (hz *Healthz) WithServer(server *http.Server) *Healthz {
	hz.server = server
	return hz
}

// WithErr sets the err that will abort app start.
func (hz *Healthz) WithErr(err error) *Healthz {
	hz.err = err
	return hz
}

// Err returns any errors that are generated before app start.
func (hz *Healthz) Err() error {
	return hz.err
}

// WithDefaultHeaders sets the default headers
func (hz *Healthz) WithDefaultHeaders(headers map[string]string) *Healthz {
	hz.defaultHeaders = headers
	return hz
}

// WithDefaultHeader adds a default header.
func (hz *Healthz) WithDefaultHeader(key string, value string) *Healthz {
	hz.defaultHeaders[key] = value
	return hz
}

// DefaultHeaders returns the default headers.
func (hz *Healthz) DefaultHeaders() map[string]string {
	return hz.defaultHeaders
}

// WithState sets app state and returns a reference to the app for building apps with a fluent api.
func (hz *Healthz) WithState(key string, value interface{}) *Healthz {
	hz.state[key] = value
	return hz
}

// GetState gets app state element by key.
func (hz *Healthz) GetState(key string) interface{} {
	if value, hasValue := hz.state[key]; hasValue {
		return value
	}
	return nil
}

// SetState sets app state.
func (hz *Healthz) SetState(key string, value interface{}) {
	hz.state[key] = value
}

// State is a bag for common app state.
func (hz *Healthz) State() State {
	return hz.state
}

// RecoverPanics returns if the app recovers panics.
func (hz *Healthz) RecoverPanics() bool {
	return hz.recoverPanics
}

// WithRecoverPanics sets if the app should recover panics.
func (hz *Healthz) WithRecoverPanics(value bool) *Healthz {
	hz.recoverPanics = value
	return hz
}

// MaxHeaderBytes returns the app max header bytes.
func (hz *Healthz) MaxHeaderBytes() int {
	return hz.maxHeaderBytes
}

// WithMaxHeaderBytes sets the max header bytes value and returns a reference.
func (hz *Healthz) WithMaxHeaderBytes(byteCount int) *Healthz {
	hz.maxHeaderBytes = byteCount
	return hz
}

// ReadHeaderTimeout returns the read header timeout for the server.
func (hz *Healthz) ReadHeaderTimeout() time.Duration {
	return hz.readHeaderTimeout
}

// WithReadHeaderTimeout returns the read header timeout for the server.
func (hz *Healthz) WithReadHeaderTimeout(timeout time.Duration) *Healthz {
	hz.readHeaderTimeout = timeout
	return hz
}

// ReadTimeout returns the read timeout for the server.
func (hz *Healthz) ReadTimeout() time.Duration {
	return hz.readTimeout
}

// WithReadTimeout sets the read timeout for the server and returns a reference to the app for building apps with a fluent api.
func (hz *Healthz) WithReadTimeout(timeout time.Duration) *Healthz {
	hz.readTimeout = timeout
	return hz
}

// IdleTimeout is the time before we close a connection.
func (hz *Healthz) IdleTimeout() time.Duration {
	return hz.idleTimeout
}

// WithIdleTimeout sets the idle timeout.
func (hz *Healthz) WithIdleTimeout(timeout time.Duration) *Healthz {
	hz.idleTimeout = timeout
	return hz
}

// WriteTimeout returns the write timeout for the server.
func (hz *Healthz) WriteTimeout() time.Duration {
	return hz.writeTimeout
}

// WithWriteTimeout sets the write timeout for the server and returns a reference to the app for building apps with a fluent api.
func (hz *Healthz) WithWriteTimeout(timeout time.Duration) *Healthz {
	hz.writeTimeout = timeout
	return hz
}

// WithPort sets the port for the bind address of the app, and returns a reference to the app.
func (hz *Healthz) WithPort(port int32) *Healthz {
	hz.SetPort(port)
	return hz
}

// SetPort sets the port the app listens on, typically to `:%d` which indicates listen on any interface.
func (hz *Healthz) SetPort(port int32) {
	hz.bindAddr = fmt.Sprintf(":%v", port)
}

// WithPortFromEnv sets the port from an environment variable, and returns a reference to the app.
func (hz *Healthz) WithPortFromEnv() *Healthz {
	hz.SetPortFromEnv()
	return hz
}

// SetPortFromEnv sets the port from an environment variable, and returns a reference to the app.
func (hz *Healthz) SetPortFromEnv() {
	if env.Env().Has(EnvironmentVariablePort) {
		port, err := env.Env().Int32(EnvironmentVariablePort)
		if err != nil {
			hz.err = err
		}
		hz.bindAddr = fmt.Sprintf(":%v", port)
	}
}

// BindAddr returns the address the server will bind to.
func (hz *Healthz) BindAddr() string {
	return hz.bindAddr
}

// WithBindAddr sets the address the app listens on, and returns a reference to the app.
func (hz *Healthz) WithBindAddr(bindAddr string) *Healthz {
	hz.bindAddr = bindAddr
	return hz
}

// WithBindAddrFromEnv sets the address the app listens on, and returns a reference to the app.
func (hz *Healthz) WithBindAddrFromEnv() *Healthz {
	hz.bindAddr = env.Env().String(EnvironmentVariableBindAddr)
	return hz
}

// Logger returns the diagnostics agent for the app.
func (hz *Healthz) Logger() *logger.Logger {
	return hz.log
}

// WithLogger sets the app logger agent and returns a reference to the app.
// It also sets underlying loggers in any child resources like providers and the auth manager.
func (hz *Healthz) WithLogger(log *logger.Logger) *Healthz {
	hz.log = log
	return hz
}

// Start starts the server.
func (hz *Healthz) Start() (err error) {
	if hz.app == nil {
		err = exception.Wrap(ErrHealthzAppUnset)
		return
	}
	start := time.Now()
	if hz.log != nil {
		hz.log.SyncTrigger(NewAppEvent(HealthzStart).WithHealthz(hz))
		defer hz.log.SyncTrigger(NewAppEvent(HealthzExit).WithHealthz(hz).WithErr(err))
	}

	if hz.server == nil {
		hz.server = hz.CreateServer()
	}
	hz.vars[VarzStarted] = time.Now().UTC()

	if hz.app.log != nil {
		hz.app.log.Listen(logger.WebRequest, ListenerHealthz, logger.NewWebRequestEventListener(hz.appWebRequestListener))
		hz.app.log.Listen(logger.Error, ListenerHealthz, logger.NewErrorEventListener(hz.appErrorListener))
		hz.app.log.Listen(logger.Fatal, ListenerHealthz, logger.NewErrorEventListener(hz.appErrorListener))
	}

	if hz.log != nil {
		hz.log.SyncInfof("healthz server started, listening on %s", hz.server.Addr)
		if hz.log.Flags() != nil {
			hz.log.SyncInfof("healthz server logging flags %s", hz.log.Flags().String())
		}
	}

	var listener net.Listener
	listener, err = net.Listen("tcp", hz.bindAddr)
	if err != nil {
		err = exception.Wrap(err)
		return
	}
	hz.listener = listener.(*net.TCPListener)

	if hz.log != nil {
		hz.log.SyncTrigger(NewAppEvent(HealthzStartComplete).WithHealthz(hz).WithElapsed(time.Since(start)))
	}

	return hz.server.Serve(TCPKeepAliveListener{hz.listener})
}

// Shutdown stops the server.
func (hz *Healthz) Shutdown() error {
	if hz.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if hz.log != nil {
		hz.log.SyncInfof("healthz server shutting down")
	}
	hz.server.SetKeepAlivesEnabled(false)
	return exception.Wrap(hz.server.Shutdown(ctx))
}

// CreateServer returns the basic http.Server for the app.
func (hz *Healthz) CreateServer() *http.Server {
	return &http.Server{
		Addr:              hz.BindAddr(),
		Handler:           hz,
		MaxHeaderBytes:    hz.maxHeaderBytes,
		ReadTimeout:       hz.readTimeout,
		ReadHeaderTimeout: hz.readHeaderTimeout,
		WriteTimeout:      hz.writeTimeout,
		IdleTimeout:       hz.idleTimeout,
	}
}

// ServeHTTP makes the router implement the http.Handler interface.
func (hz *Healthz) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if hz.recoverPanics {
		defer hz.recover(w, r)
	}

	res := NewRawResponseWriter(w)
	res.Header().Set(HeaderContentEncoding, ContentEncodingIdentity)

	route := strings.ToLower(r.URL.Path)

	start := time.Now()
	if hz.log != nil {
		hz.log.Trigger(logger.NewWebRequestStartEvent(r).WithState(hz.state).WithRoute(route))

		defer func() {
			hz.log.Trigger(logger.NewWebRequestEvent(r).
				WithStatusCode(res.StatusCode()).
				WithElapsed(time.Since(start)).
				WithContentLength(int64(res.ContentLength())).
				WithState(hz.state))
		}()
	}

	if len(hz.defaultHeaders) > 0 {
		for key, value := range hz.defaultHeaders {
			res.Header().Set(key, value)
		}
	}

	switch route {
	case "/healthz":
		hz.healthzHandler(res, r)
	case "/varz":
		hz.varzHandler(res, r)
	default:
		http.NotFound(res, r)
	}

	if err := res.Close(); err != nil && err != http.ErrBodyNotAllowed && hz.log != nil {
		hz.log.Error(err)
	}
}

func (hz *Healthz) recover(w http.ResponseWriter, req *http.Request) {
	if rcv := recover(); rcv != nil {
		if hz.log != nil {
			hz.log.Fatalf("%v", rcv)
		}

		http.Error(w, fmt.Sprintf("%v", rcv), http.StatusInternalServerError)
		return
	}
}

func (hz *Healthz) healthzHandler(w ResponseWriter, r *http.Request) {
	if hz.app.Running() {
		w.WriteHeader(http.StatusOK)
		w.Header().Set(HeaderContentType, ContentTypeText)
		fmt.Fprintf(w, "OK!\n")
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set(HeaderContentType, ContentTypeText)
	fmt.Fprintf(w, "Failure!\n")
	return
}

// /varz
// writes out the current stats
func (hz *Healthz) varzHandler(w ResponseWriter, r *http.Request) {
	hz.varsLock.Lock()
	defer hz.varsLock.Unlock()

	keys := make([]string, len(hz.vars))

	var index int
	for key := range hz.vars {
		keys[index] = key
		index++
	}

	sort.Strings(keys)

	w.WriteHeader(http.StatusOK)
	w.Header().Set(HeaderContentType, ContentTypeText)
	for _, key := range keys {
		fmt.Fprintf(w, "%s: %v\n", key, hz.vars[key])
	}
}

func (hz *Healthz) appWebRequestListener(wre *logger.WebRequestEvent) {
	hz.varsLock.Lock()
	defer hz.varsLock.Unlock()

	hz.incrementVarUnsafe(VarzRequests)
	if wre.StatusCode() >= http.StatusInternalServerError {
		hz.incrementVarUnsafe(VarzRequests5xx)
	} else if wre.StatusCode() >= http.StatusBadRequest {
		hz.incrementVarUnsafe(VarzRequests4xx)
	} else if wre.StatusCode() >= http.StatusMultipleChoices {
		hz.incrementVarUnsafe(VarzRequests3xx)
	} else {
		hz.incrementVarUnsafe(VarzRequests2xx)
	}
}

func (hz *Healthz) appErrorListener(e *logger.ErrorEvent) {
	hz.varsLock.Lock()
	defer hz.varsLock.Unlock()

	switch e.Flag() {
	case logger.Error:
		hz.incrementVarUnsafe(VarzErrors)
		return
	case logger.Fatal:
		hz.incrementVarUnsafe(VarzFatals)
		return
	}
}

func (hz *Healthz) incrementVarUnsafe(key string) {
	if value, hasValue := hz.vars[key]; hasValue {
		if typed, isTyped := value.(int64); isTyped {
			hz.vars[key] = typed + 1
		}
	} else {
		hz.vars[key] = int64(1)
	}
}
