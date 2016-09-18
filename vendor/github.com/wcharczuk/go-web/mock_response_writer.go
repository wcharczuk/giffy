package web

import (
	"bytes"
	"io"
	"net/http"
)

// --------------------------------------------------------------------------------
// MockResponseWriter
// --------------------------------------------------------------------------------

// NewMockResponseWriter returns a mocked response writer.
func NewMockResponseWriter(buffer io.Writer) *MockResponseWriter {
	return &MockResponseWriter{
		innerWriter: buffer,
		contents:    bytes.NewBuffer([]byte{}),
		headers:     http.Header{},
	}
}

// MockResponseWriter is an object that satisfies response writer but uses an internal buffer.
type MockResponseWriter struct {
	innerWriter   io.Writer
	contents      *bytes.Buffer
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

// InnerResponse returns the backing httpresponse writer.
func (res *MockResponseWriter) InnerResponse() http.ResponseWriter {
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

// Bytes returns the raw response.
func (res *MockResponseWriter) Bytes() []byte {
	return res.contents.Bytes()
}

// Flush is a no-op.
func (res *MockResponseWriter) Flush() error {
	_, err := res.contents.WriteTo(res.innerWriter)
	return err
}

// Close is a no-op.
func (res *MockResponseWriter) Close() error {
	return nil
}
