package web

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/blend/go-sdk/env"
	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/logger"
)

// NewHTTPSUpgrader returns a new HTTPSUpgrader which redirects HTTP to HTTPS
func NewHTTPSUpgrader() *HTTPSUpgrader {
	return &HTTPSUpgrader{}
}

// NewHTTPSUpgraderFromEnv returns a new https upgrader from enviroment variables.
func NewHTTPSUpgraderFromEnv() *HTTPSUpgrader {
	return NewHTTPSUpgraderFromConfig(NewHTTPSUpgraderConfigFromEnv())
}

// NewHTTPSUpgraderFromConfig creates a new https upgrader from a config.
func NewHTTPSUpgraderFromConfig(cfg *HTTPSUpgraderConfig) *HTTPSUpgrader {
	return &HTTPSUpgrader{
		bindAddr:          cfg.GetBindAddr(),
		targetPort:        cfg.GetTargetPort(),
		maxHeaderBytes:    cfg.GetMaxHeaderBytes(),
		readTimeout:       cfg.GetReadTimeout(),
		readHeaderTimeout: cfg.GetReadHeaderTimeout(),
		writeTimeout:      cfg.GetWriteTimeout(),
		idleTimeout:       cfg.GetIdleTimeout(),
	}
}

// HTTPSUpgrader redirects HTTP to HTTPS
type HTTPSUpgrader struct {
	bindAddr   string
	targetPort int32

	server *http.Server

	maxHeaderBytes    int
	readTimeout       time.Duration
	readHeaderTimeout time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration

	err error
	log *logger.Logger
}

// WithTargetPort sets the target upgrade port.
// It defaults to unset, or the https default of 443.
func (hu *HTTPSUpgrader) WithTargetPort(port int32) *HTTPSUpgrader {
	hu.targetPort = port
	return hu
}

// TargetPort returns the target upgrade port.
// It defaults to unset, or the https default of 443.
func (hu *HTTPSUpgrader) TargetPort() int32 {
	return hu.targetPort
}

// WithBindAddr sets the address the app listens on, and returns a reference to the app.
func (hu *HTTPSUpgrader) WithBindAddr(bindAddr string) *HTTPSUpgrader {
	hu.bindAddr = bindAddr
	return hu
}

// WithBindAddrFromEnv sets the address the app listens on, and returns a reference to the app.
func (hu *HTTPSUpgrader) WithBindAddrFromEnv() *HTTPSUpgrader {
	hu.bindAddr = env.Env().String(EnvironmentVariableBindAddr)
	return hu
}

// BindAddr returns the address the server will bind to.
func (hu *HTTPSUpgrader) BindAddr() string {
	return hu.bindAddr
}

// WithPort sets the port for the bind address of the app, and returns a reference to the app.
func (hu *HTTPSUpgrader) WithPort(port int32) *HTTPSUpgrader {
	hu.SetPort(port)
	return hu
}

// SetPort sets the port the app listens on, typically to `:%d` which indicates listen on any interface.
func (hu *HTTPSUpgrader) SetPort(port int32) {
	hu.bindAddr = fmt.Sprintf(":%v", port)
}

// WithPortFromEnv sets the port from an environment variable, and returns a reference to the app.
func (hu *HTTPSUpgrader) WithPortFromEnv() *HTTPSUpgrader {
	hu.SetPortFromEnv()
	return hu
}

// SetPortFromEnv sets the port from an environment variable, and returns a reference to the app.
func (hu *HTTPSUpgrader) SetPortFromEnv() {
	if env.Env().Has(EnvironmentVariablePort) {
		port, err := env.Env().Int32(EnvironmentVariablePort)
		if err != nil {
			hu.err = err
		}
		hu.bindAddr = fmt.Sprintf(":%v", port)
	}
}

// WithLogger sets the underlying logger.
func (hu *HTTPSUpgrader) WithLogger(log *logger.Logger) *HTTPSUpgrader {
	hu.log = log
	return hu
}

// Logger returns the underlying logger.
func (hu *HTTPSUpgrader) Logger() *logger.Logger {
	return hu.log
}

// MaxHeaderBytes returns the app max header bytes.
func (hu *HTTPSUpgrader) MaxHeaderBytes() int {
	return hu.maxHeaderBytes
}

// WithMaxHeaderBytes sets the max header bytes value and returns a reference.
func (hu *HTTPSUpgrader) WithMaxHeaderBytes(byteCount int) *HTTPSUpgrader {
	hu.maxHeaderBytes = byteCount
	return hu
}

// ReadHeaderTimeout returns the read header timeout for the server.
func (hu *HTTPSUpgrader) ReadHeaderTimeout() time.Duration {
	return hu.readHeaderTimeout
}

// WithReadHeaderTimeout returns the read header timeout for the server.
func (hu *HTTPSUpgrader) WithReadHeaderTimeout(timeout time.Duration) *HTTPSUpgrader {
	hu.readHeaderTimeout = timeout
	return hu
}

// ReadTimeout returns the read timeout for the server.
func (hu *HTTPSUpgrader) ReadTimeout() time.Duration {
	return hu.readTimeout
}

// WithReadTimeout sets the read timeout for the server and returns a reference to the app for building apps with a fluent api.
func (hu *HTTPSUpgrader) WithReadTimeout(timeout time.Duration) *HTTPSUpgrader {
	hu.readTimeout = timeout
	return hu
}

// IdleTimeout is the time before we close a connection.
func (hu *HTTPSUpgrader) IdleTimeout() time.Duration {
	return hu.idleTimeout
}

// WithIdleTimeout sets the idle timeout.
func (hu *HTTPSUpgrader) WithIdleTimeout(timeout time.Duration) *HTTPSUpgrader {
	hu.idleTimeout = timeout
	return hu
}

// WriteTimeout returns the write timeout for the server.
func (hu *HTTPSUpgrader) WriteTimeout() time.Duration {
	return hu.writeTimeout
}

// WithWriteTimeout sets the write timeout for the server and returns a reference to the app for building apps with a fluent api.
func (hu *HTTPSUpgrader) WithWriteTimeout(timeout time.Duration) *HTTPSUpgrader {
	hu.writeTimeout = timeout
	return hu
}

// Server returns the basic http.Server for the app.
func (hu *HTTPSUpgrader) Server() *http.Server {
	return &http.Server{
		Addr:              hu.BindAddr(),
		Handler:           hu,
		MaxHeaderBytes:    hu.maxHeaderBytes,
		ReadTimeout:       hu.readTimeout,
		ReadHeaderTimeout: hu.readHeaderTimeout,
		WriteTimeout:      hu.writeTimeout,
		IdleTimeout:       hu.idleTimeout,
	}
}

// ServeHTTP redirects HTTP to HTTPS
func (hu *HTTPSUpgrader) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	start := time.Now()
	response := []byte("Upgrade Required")
	if hu.log != nil {
		defer hu.log.Trigger(logger.NewWebRequestEvent(req).
			WithStatusCode(http.StatusMovedPermanently).
			WithContentLength(int64(len(response))).
			WithContentType(ContentTypeText).
			WithElapsed(time.Since(start)))
	}

	newURL := *req.URL
	newURL.Scheme = SchemeHTTPS
	if len(newURL.Host) == 0 {
		newURL.Host = req.Host
	}
	if hu.targetPort > 0 {
		if strings.Contains(newURL.Host, ":") {
			newURL.Host = fmt.Sprintf("%s:%d", strings.SplitN(newURL.Host, ":", 2)[0], hu.targetPort)
		} else {
			newURL.Host = fmt.Sprintf("%s:%d", newURL.Host, hu.targetPort)
		}
	}

	http.Redirect(rw, req, newURL.String(), http.StatusMovedPermanently)
}

// Start starts the server and binds to the given address.
func (hu *HTTPSUpgrader) Start() error {
	return hu.StartWithServer(hu.Server())
}

// StartWithServer starts the app on a custom server.
// This lets you configure things like TLS keys and
// other options.
func (hu *HTTPSUpgrader) StartWithServer(server *http.Server) (err error) {
	// early exit if we already had an issue.
	if hu.err != nil {
		err = hu.err
		return
	}
	if hu.targetPort > 0 {
		hu.log.SyncInfof("https upgrade server started, listening on %s, upgrading to :%d", server.Addr, hu.targetPort)
	} else {
		hu.log.SyncInfof("https upgrade server started, listening on %s", server.Addr)
	}
	hu.server = server
	err = exception.Wrap(server.ListenAndServe())
	return
}

// Shutdown stops the server.
func (hu *HTTPSUpgrader) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hu.log.SyncInfof("https upgrade server shutting down")
	hu.server.SetKeepAlivesEnabled(false)
	return exception.Wrap(hu.server.Shutdown(ctx))
}
