package web

import (
	"compress/gzip"
	"io"
	"net/http"
)

// NewGZippedResponseWriter returns a new gzipped response writer.
func NewGZippedResponseWriter(writer io.Writer) *GZippedResponseWriter {
	gz := gzip.NewWriter(writer)
	return &GZippedResponseWriter{Writer: gz, Headers: http.Header{}}
}

// GZippedResponseWriter is a response writer that compresses output.
type GZippedResponseWriter struct {
	Writer       *gzip.Writer
	ResponseMeta http.ResponseWriter
	Headers      http.Header
	StatusCode   int
}

// Write writes the byes to the stream.
func (gzw GZippedResponseWriter) Write(b []byte) (int, error) {
	return gzw.Writer.Write(b)
}

// Header returns the headers for the response.
func (gzw GZippedResponseWriter) Header() http.Header {
	return gzw.Headers
}

// WriteHeader writes a status code.
func (gzw *GZippedResponseWriter) WriteHeader(code int) {
	gzw.StatusCode = code
}
