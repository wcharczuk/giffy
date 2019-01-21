package request

import (
	"github.com/blend/go-sdk/logger"
)

// NewFactory creates a new Factory.
func NewFactory() *Factory {
	return &Factory{}
}

// Factory is a helper to create requests with common metadata.
// It is generally for creating requests to *any* host.
type Factory struct {
	Log                    logger.Log
	MockedResponseProvider MockedResponseProvider
	OnRequest              Handler
	OnResponse             ResponseHandler
	Tracer                 Tracer
}

// WithLogger sets the logger.
func (m *Factory) WithLogger(log logger.Log) *Factory {
	m.Log = log
	return m
}

// WithMockedResponseProvider sets the mocked response provider.
func (m *Factory) WithMockedResponseProvider(mrp MockedResponseProvider) *Factory {
	m.MockedResponseProvider = mrp
	return m
}

// WithOnRequest sets the on request handler..
func (m *Factory) WithOnRequest(handler Handler) *Factory {
	m.OnRequest = handler
	return m
}

// WithOnResponse sets the on response handler.
func (m *Factory) WithOnResponse(handler ResponseHandler) *Factory {
	m.OnResponse = handler
	return m
}

// WithTracer sets the tracer.
func (m *Factory) WithTracer(tracer Tracer) *Factory {
	m.Tracer = tracer
	return m
}

// Create creates a new request.
func (m Factory) Create() *Request {
	return New().
		WithLogger(m.Log).
		WithMockProvider(m.MockedResponseProvider).
		WithRequestHandler(m.OnRequest).
		WithResponseHandler(m.OnResponse).
		WithTracer(m.Tracer)
}

// Get returns a new get request for a given url.
func (m Factory) Get(url string) (*Request, error) {
	return m.Create().AsGet().WithRawURL(url)
}

// Post returns a new post request for a given url.
func (m Factory) Post(url string, body []byte) (*Request, error) {
	return m.Create().AsPost().WithPostBody(body).WithRawURL(url)
}
