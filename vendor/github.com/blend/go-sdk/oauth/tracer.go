package oauth

import "net/http"

// Tracer is a trace shim.
type Tracer interface {
	Start(r *http.Request) TraceFinisher
}

// TraceFinisher is a finisher for a trace.
type TraceFinisher interface {
	Finish(*http.Request, *Result, error)
}
