package web

import "github.com/blendlabs/go-exception"

// APIResponseMeta is the meta component of a service response.
type APIResponseMeta struct {
	HTTPCode   int                  `json:"http_code"`
	APIVersion string               `json:"api_version,omitempty"`
	Message    string               `json:"message,omitempty"`
	Exception  *exception.Exception `json:"exception,omitempty"`
}

// APIResponse is the standard API response format.
type APIResponse struct {
	Meta     *APIResponseMeta `json:"meta"`
	Response interface{}      `json:"response"`
}
