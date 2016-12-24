package web

// JSONResult is a json result.
type JSONResult struct {
	StatusCode int
	Response   interface{}
}

// Render renders the result
func (ar *JSONResult) Render(rc *RequestContext) error {
	return WriteJSON(rc.Response, rc.Request, ar.StatusCode, ar.Response)
}
