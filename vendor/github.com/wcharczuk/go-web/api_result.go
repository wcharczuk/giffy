package web

import "github.com/blendlabs/go-exception"

// APIResultMeta is the meta component of a service response.
type APIResultMeta struct {
	HTTPCode   int                  `json:"http_code"`
	APIVersion string               `json:"api_version,omitempty"`
	Message    string               `json:"message,omitempty"`
	Exception  *exception.Exception `json:"exception,omitempty"`
}

// APIResult is the standard API response format.
type APIResult struct {
	Meta     *APIResultMeta `json:"meta"`
	Response interface{}    `json:"response"`
}

// Render turns the response into JSON.
func (ar *APIResult) Render(ctx *RequestContext) error {
	_, err := WriteJSON(ctx.Response, ctx.Request, ar.Meta.HTTPCode, ar)
	return err
}
