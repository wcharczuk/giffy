package web

import (
	"bytes"
	"net/http"
)

// --------------------------------------------------------------------------------
// UncompressedResponseWriter
// --------------------------------------------------------------------------------

// NewResponseWriter creates a new uncompressed response writer.
func NewResponseWriter(w http.ResponseWriter) *UncompressedResponseWriter {
	return &UncompressedResponseWriter{
		innerResponse: w,
	}
}

// NewBufferedResponseWriter creates a new uncompressed response writer.
func NewBufferedResponseWriter(w http.ResponseWriter) *UncompressedResponseWriter {
	return &UncompressedResponseWriter{
		innerResponse:  w,
		responseBuffer: bytes.NewBuffer([]byte{}),
	}
}

// UncompressedResponseWriter a better response writer
type UncompressedResponseWriter struct {
	innerResponse http.ResponseWriter
	contentLength int
	statusCode    int

	responseBuffer *bytes.Buffer
}

// Write writes the data to the response.
func (rw *UncompressedResponseWriter) Write(b []byte) (int, error) {
	if rw.responseBuffer == nil {
		written, err := rw.innerResponse.Write(b)
		rw.contentLength += written
		return written, err
	}
	written, err := rw.responseBuffer.Write(b)
	rw.contentLength += written
	return written, err
}

// Header accesses the response header collection.
func (rw *UncompressedResponseWriter) Header() http.Header {
	return rw.innerResponse.Header()
}

// WriteHeader is actually a terrible name and this writes the status code.
func (rw *UncompressedResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.innerResponse.WriteHeader(code)
}

// InnerResponse returns the backing writer.
func (rw *UncompressedResponseWriter) InnerResponse() http.ResponseWriter {
	return rw.innerResponse
}

// StatusCode returns the status code.
func (rw *UncompressedResponseWriter) StatusCode() int {
	return rw.statusCode
}

// ContentLength returns the content length
func (rw *UncompressedResponseWriter) ContentLength() int {
	return rw.contentLength
}

// Bytes returns the raw data returned.
func (rw *UncompressedResponseWriter) Bytes() []byte {
	if rw.responseBuffer == nil {
		return []byte{}
	}
	return rw.responseBuffer.Bytes()
}

// Flush is a no op on raw response writers.
func (rw *UncompressedResponseWriter) Flush() error {
	if rw.responseBuffer == nil {
		return nil
	}
	_, err := rw.responseBuffer.WriteTo(rw.innerResponse)
	return err
}

// Close disposes of the response writer.
func (rw *UncompressedResponseWriter) Close() error {

	return nil
}
