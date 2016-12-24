package web

import "net/http"

// NoContentResult returns a no content response.
type NoContentResult struct{}

// Render renders a static result.
func (ncr *NoContentResult) Render(rc *RequestContext) error {
	rc.Response.WriteHeader(http.StatusNoContent)
	_, err := rc.Response.Write([]byte{})
	return err
}
