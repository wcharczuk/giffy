package web

// XMLResult is a json result.
type XMLResult struct {
	StatusCode int
	Response   interface{}
}

// Render renders the result
func (ar *XMLResult) Render(ctx *Ctx) error {
	return WriteXML(ctx.Response(), ctx.Request(), ar.StatusCode, ar.Response)
}
