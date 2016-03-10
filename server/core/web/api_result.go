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
