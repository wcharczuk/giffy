package web

import "net/http"

// NewResponseWriter creates a new response writer.
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		HTTPResponse: w,
	}
}

// ResponseWriter a better response writer
type ResponseWriter struct {
	HTTPResponse  http.ResponseWriter
	StatusCode    int
	ContentLength int
}

// Write writes the data to the response.
func (sarw *ResponseWriter) Write(b []byte) (int, error) {
	bytesWritten, err := sarw.HTTPResponse.Write(b)
	sarw.ContentLength = sarw.ContentLength + bytesWritten
	return bytesWritten, err
}

// Header accesses the response header collection.
func (sarw *ResponseWriter) Header() http.Header {
	return sarw.HTTPResponse.Header()
}

// WriteHeader is actually a terrible name and this writes the status code.
func (sarw *ResponseWriter) WriteHeader(code int) {
	sarw.StatusCode = code
	sarw.HTTPResponse.WriteHeader(code)
}
