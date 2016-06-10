package web

import (
	"compress/gzip"
	"io"
	"net/http"
)

// ResponseWriter is a super-type of http.ResponseWriter that includes
// the StatusCode and ContentLength for the request
type ResponseWriter interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(int)
	Flush() error
	StatusCode() int
	ContentLength() int
}

// --------------------------------------------------------------------------------
// RawResponseWriter
// --------------------------------------------------------------------------------

// NewRawResponseWriter creates a new response writer.
func NewRawResponseWriter(w http.ResponseWriter) *RawResponseWriter {
	return &RawResponseWriter{
		HTTPResponse: w,
	}
}

// RawResponseWriter a better response writer
type RawResponseWriter struct {
	HTTPResponse  http.ResponseWriter
	statusCode    int
	contentLength int
}

// Write writes the data to the response.
func (rw *RawResponseWriter) Write(b []byte) (int, error) {
	bytesWritten, err := rw.HTTPResponse.Write(b)
	rw.contentLength = rw.contentLength + bytesWritten
	return bytesWritten, err
}

// Header accesses the response header collection.
func (rw *RawResponseWriter) Header() http.Header {
	return rw.HTTPResponse.Header()
}

// WriteHeader is actually a terrible name and this writes the status code.
func (rw *RawResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.HTTPResponse.WriteHeader(code)
}

// Flush is a no op on raw response writers.
func (rw *RawResponseWriter) Flush() error {
	return nil
}

// StatusCode returns the status code.
func (rw *RawResponseWriter) StatusCode() int {
	return rw.statusCode
}

// ContentLength returns the content length
func (rw *RawResponseWriter) ContentLength() int {
	return rw.contentLength
}

// NewCompressedResponseWriter returns a new gzipped response writer.
func NewCompressedResponseWriter(w http.ResponseWriter) *CompressedResponseWriter {
	return &CompressedResponseWriter{HTTPResponse: w}
}

// --------------------------------------------------------------------------------
// CompressedResponseWriter
// --------------------------------------------------------------------------------

// CompressedResponseWriter is a response writer that compresses output.
type CompressedResponseWriter struct {
	GZIPWriter    *gzip.Writer
	HTTPResponse  http.ResponseWriter
	statusCode    int
	contentLength int
}

func (crw *CompressedResponseWriter) ensureCompressedStream() {
	if crw.GZIPWriter == nil {
		crw.GZIPWriter = gzip.NewWriter(crw.HTTPResponse)
	}
}

// Write writes the byes to the stream.
func (crw *CompressedResponseWriter) Write(b []byte) (int, error) {
	crw.ensureCompressedStream()
	bw, err := crw.GZIPWriter.Write(b)
	crw.contentLength = crw.contentLength + bw
	return bw, err
}

// Header returns the headers for the response.
func (crw *CompressedResponseWriter) Header() http.Header {
	return crw.HTTPResponse.Header()
}

// WriteHeader writes a status code.
func (crw *CompressedResponseWriter) WriteHeader(code int) {
	crw.statusCode = code
	crw.HTTPResponse.WriteHeader(code)
}

// Flush pushes any buffered data out to the response.
func (crw *CompressedResponseWriter) Flush() error {
	crw.ensureCompressedStream()
	return crw.GZIPWriter.Flush()
}

// Close closes any underlying resources.
func (crw *CompressedResponseWriter) Close() error {
	if crw.GZIPWriter != nil {
		err := crw.GZIPWriter.Close()
		crw.GZIPWriter = nil
		return err
	}
	return nil
}

// StatusCode returns the status code for the request.
func (crw *CompressedResponseWriter) StatusCode() int {
	return crw.statusCode
}

// ContentLength returns the content length for the request.
func (crw *CompressedResponseWriter) ContentLength() int {
	return crw.contentLength
}

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
