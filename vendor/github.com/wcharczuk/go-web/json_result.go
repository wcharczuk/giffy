package web

// JSONResult is a json result.
type JSONResult struct {
	StatusCode int
	Response   interface{}
}

// Render turns the response into JSON.
func (ar *JSONResult) Render(ctx *RequestContext) error {
	_, err := WriteJSON(ctx.Response, ctx.Request, ar.StatusCode, ar.Response)
	return err
}
