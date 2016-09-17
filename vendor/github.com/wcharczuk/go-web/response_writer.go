package web

import "net/http"

// ResponseWriter is a super-type of http.ResponseWriter that includes
// the StatusCode and ContentLength for the request
type ResponseWriter interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(int)
	InnerWriter() http.ResponseWriter
	Flush() error
	StatusCode() int
	ContentLength() int
}
