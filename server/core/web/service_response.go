package web

import "github.com/blendlabs/go-exception"

// ServiceResponseMeta is the meta component of a service response.
type ServiceResponseMeta struct {
	HTTPCode   int                  `json:"http_code"`
	APIVersion string               `json:"api_version,omitempty"`
	Message    string               `json:"message,omitempty"`
	Exception  *exception.Exception `json:"exception,omitempty"`
}

// ServiceResponse is the standard API response format.
type ServiceResponse struct {
	Meta     ServiceResponseMeta `json:"meta"`
	Response interface{}         `json:"response"`
}
