package web

import "net/http"

// NoContentResult returns a no content response.
type NoContentResult struct{}

// Render renders a static result.
func (ncr *NoContentResult) Render(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusNoContent)
	_, err := w.Write([]byte{})
	return err
}
