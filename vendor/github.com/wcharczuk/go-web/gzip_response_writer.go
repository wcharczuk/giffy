package web

import (
	"compress/gzip"
	"net/http"
)

// --------------------------------------------------------------------------------
// CompressedResponseWriter
// --------------------------------------------------------------------------------

// NewCompressedResponseWriter returns a new gzipped response writer.
func NewCompressedResponseWriter(w http.ResponseWriter) *CompressedResponseWriter {
	return &CompressedResponseWriter{HTTPResponse: w}
}

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
	written, err := crw.GZIPWriter.Write(b)
	crw.contentLength += written
	return written, err
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

// InnerWriter returns the backing http response.
func (crw *CompressedResponseWriter) InnerWriter() http.ResponseWriter {
	return crw.HTTPResponse
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
