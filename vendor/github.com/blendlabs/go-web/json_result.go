package web

// JSONResult is a json result.
type JSONResult struct {
	StatusCode int
	Response   interface{}
}

// Render renders the result
func (ar *JSONResult) Render(ctx *Ctx) error {
	return WriteJSON(ctx.Response(), ctx.Request(), ar.StatusCode, ar.Response)
}
