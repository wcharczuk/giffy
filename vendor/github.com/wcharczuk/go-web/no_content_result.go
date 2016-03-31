package web

import "net/http"

// NoContentResult returns a no content response.
type NoContentResult struct{}

// Provider returns the result provider if there is one.
func (ncr NoContentResult) Provider() ControllerResultProvider {
	return nil
}

// Render renders a static result.
func (ncr NoContentResult) Render(ctx *RequestContext) error {
	ctx.Response.WriteHeader(http.StatusNoContent)
	_, err := ctx.Response.Write([]byte{})
	return err
}
