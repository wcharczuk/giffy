package web

import (
	"bytes"
	"compress/gzip"
	"net/http"
)

// --------------------------------------------------------------------------------
// CompressedResponseWriter
// --------------------------------------------------------------------------------

// NewCompressedResponseWriter returns a new gzipped response writer.
func NewCompressedResponseWriter(w http.ResponseWriter) ResponseWriter {
	return &CompressedResponseWriter{
		innerResponse: w,
	}
}

// NewBufferedCompressedResponseWriter returns a new gzipped response writer.
func NewBufferedCompressedResponseWriter(w http.ResponseWriter) ResponseWriter {
	return &CompressedResponseWriter{
		innerResponse:  w,
		responseBuffer: bytes.NewBuffer([]byte{}),
	}
}

// CompressedResponseWriter is a response writer that compresses output.
type CompressedResponseWriter struct {
	gzipWriter     *gzip.Writer
	innerResponse  http.ResponseWriter
	responseBuffer *bytes.Buffer
	statusCode     int
	contentLength  int
}

func (crw *CompressedResponseWriter) ensureCompressedStream() {
	if crw.gzipWriter == nil {
		crw.gzipWriter = gzip.NewWriter(crw.innerResponse)
	}
}

// Write writes the byes to the stream.
func (crw *CompressedResponseWriter) Write(b []byte) (int, error) {
	if crw.responseBuffer == nil {
		crw.ensureCompressedStream()
		written, err := crw.gzipWriter.Write(b)
		crw.contentLength += written
		return written, err
	}
	written, err := crw.responseBuffer.Write(b)
	crw.contentLength += written
	return written, err
}

// Header returns the headers for the response.
func (crw *CompressedResponseWriter) Header() http.Header {
	return crw.innerResponse.Header()
}

// WriteHeader writes a status code.
func (crw *CompressedResponseWriter) WriteHeader(code int) {
	crw.statusCode = code
	crw.innerResponse.WriteHeader(code)
}

// InnerResponse returns the backing http response.
func (crw *CompressedResponseWriter) InnerResponse() http.ResponseWriter {
	return crw.innerResponse
}

// StatusCode returns the status code for the request.
func (crw *CompressedResponseWriter) StatusCode() int {
	return crw.statusCode
}

// ContentLength returns the content length for the request.
func (crw *CompressedResponseWriter) ContentLength() int {
	return crw.contentLength
}

// Bytes returns the raw response.
func (crw *CompressedResponseWriter) Bytes() []byte {
	if crw.responseBuffer == nil {
		return []byte{}
	}
	return crw.responseBuffer.Bytes()
}

// Flush pushes any buffered data out to the response.
func (crw *CompressedResponseWriter) Flush() error {
	crw.ensureCompressedStream()
	if crw.responseBuffer != nil {
		written, err := crw.gzipWriter.Write(crw.responseBuffer.Bytes())
		crw.contentLength = written
		if err != nil {
			return err
		}
	}
	return crw.gzipWriter.Flush()
}

// Close closes any underlying resources.
func (crw *CompressedResponseWriter) Close() error {
	crw.responseBuffer = nil
	if crw.gzipWriter != nil {
		err := crw.gzipWriter.Close()
		crw.gzipWriter = nil
		return err
	}
	return nil
}
