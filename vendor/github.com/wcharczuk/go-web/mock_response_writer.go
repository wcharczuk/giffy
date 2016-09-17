package web

import (
	"io"
	"net/http"
)

// --------------------------------------------------------------------------------
// MockResponseWriter
// --------------------------------------------------------------------------------

// NewMockResponseWriter returns a mocked response writer.
func NewMockResponseWriter(buffer io.Writer) *MockResponseWriter {
	return &MockResponseWriter{
		contents: buffer,
		headers:  http.Header{},
	}
}

// MockResponseWriter is an object that satisfies response writer but uses an internal buffer.
type MockResponseWriter struct {
	contents      io.Writer
	statusCode    int
	contentLength int
	headers       http.Header
}

// Write writes data and adds to ContentLength.
func (res *MockResponseWriter) Write(buffer []byte) (int, error) {
	bytes, err := res.contents.Write(buffer)
	res.contentLength = res.contentLength + bytes
	return bytes, err
}

// Header returns the response headers.
func (res *MockResponseWriter) Header() http.Header {
	return res.headers
}

// WriteHeader sets the status code.
func (res *MockResponseWriter) WriteHeader(statusCode int) {
	res.statusCode = statusCode
}

// InnerWriter returns the backing httpresponse writer.
func (res *MockResponseWriter) InnerWriter() http.ResponseWriter {
	return res
}

// StatusCode returns the status code.
func (res *MockResponseWriter) StatusCode() int {
	return res.statusCode
}

// ContentLength returns the content length.
func (res *MockResponseWriter) ContentLength() int {
	return res.contentLength
}

// Flush is a no-op.
func (res *MockResponseWriter) Flush() error {
	return nil
}
