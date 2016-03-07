package web

import (
	"compress/gzip"
	"io"
	"net/http"
)

func NewGZippedResponseWriter(writer io.Writer) *GZippedResponseWriter {
	gz := gzip.NewWriter(writer)
	return &GZippedResponseWriter{Writer: gz, Headers: http.Header{}}
}

type GZippedResponseWriter struct {
	Writer       *gzip.Writer
	ResponseMeta http.ResponseWriter
	Headers      http.Header
	StatusCode   int
}

func (gzw GZippedResponseWriter) Write(b []byte) (int, error) {
	bytes_written, bytes_err := gzw.Writer.Write(b)
	return bytes_written, bytes_err
}

func (gzw GZippedResponseWriter) Header() http.Header {
	return gzw.Headers
}

func (gzw *GZippedResponseWriter) WriteHeader(code int) {
	gzw.StatusCode = code
}
