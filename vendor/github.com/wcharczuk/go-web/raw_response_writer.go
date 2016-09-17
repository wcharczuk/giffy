package web

import "net/http"

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

// InnerWriter returns the backing writer.
func (rw *RawResponseWriter) InnerWriter() http.ResponseWriter {
	return rw.HTTPResponse
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
