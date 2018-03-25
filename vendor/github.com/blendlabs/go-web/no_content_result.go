package web

import "net/http"

// NoContentResult returns a no content response.
type NoContentResult struct{}

// Render renders a static result.
func (ncr *NoContentResult) Render(ctx *Ctx) error {
	ctx.Response().WriteHeader(http.StatusNoContent)
	_, err := ctx.Response().Write([]byte{})
	return err
}
