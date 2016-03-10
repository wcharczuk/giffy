package web

import (
	"compress/gzip"
	"net/http"
)

// NewGZippedResponseWriter returns a new gzipped response writer.
func NewGZippedResponseWriter(w http.ResponseWriter) *GZippedResponseWriter {
	return &GZippedResponseWriter{HTTPResponse: w}
}

// GZippedResponseWriter is a response writer that compresses output.
type GZippedResponseWriter struct {
	GZIPWriter   *gzip.Writer
	HTTPResponse http.ResponseWriter
	StatusCode   int
	BytesWritten int
}

func (gzw *GZippedResponseWriter) ensureCompressedStream() {
	if gzw.GZIPWriter == nil {
		gzw.GZIPWriter = gzip.NewWriter(gzw.HTTPResponse)
	}
}

// Write writes the byes to the stream.
func (gzw *GZippedResponseWriter) Write(b []byte) (int, error) {
	gzw.ensureCompressedStream()
	bw, err := gzw.GZIPWriter.Write(b)
	gzw.BytesWritten = gzw.BytesWritten + bw
	return bw, err
}

// Header returns the headers for the response.
func (gzw *GZippedResponseWriter) Header() http.Header {
	return gzw.HTTPResponse.Header()
}

// WriteHeader writes a status code.
func (gzw *GZippedResponseWriter) WriteHeader(code int) {
	gzw.StatusCode = code
	gzw.HTTPResponse.WriteHeader(code)
}

// Flush pushes any buffered data out to the response.
func (gzw *GZippedResponseWriter) Flush() error {
	gzw.ensureCompressedStream()
	return gzw.GZIPWriter.Flush()
}

// Close closes any underlying resources.
func (gzw *GZippedResponseWriter) Close() error {
	if gzw.GZIPWriter != nil {
		err := gzw.GZIPWriter.Close()
		gzw.GZIPWriter = nil
		return err
	}
	return nil
}
