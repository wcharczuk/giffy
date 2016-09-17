package web

import "net/http"

// JSONResult is a json result.
type JSONResult struct {
	StatusCode int
	Response   interface{}
}

// Render renders the result
func (ar *JSONResult) Render(w http.ResponseWriter, r *http.Request) error {
	return WriteJSON(w, r, ar.StatusCode, ar.Response)
}
