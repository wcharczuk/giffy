package web

import "github.com/blendlabs/go-exception"

type ServiceResponseMeta struct {
	HttpCode   int                  `json:"http_code"`
	ApiVersion string               `json:"api_version,omitempty"`
	Message    string               `json:"message,omitempty"`
	Exception  *exception.Exception `json:"exception,omitempty"`
}

type ServiceResponse struct {
	Meta     ServiceResponseMeta `json:"meta"`
	Response interface{}         `json:"response"`
}
