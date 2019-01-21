package db

import "context"

// Tracer is a type that can implement traces.
// If any of the methods return a nil finisher, they will be skipped.
type Tracer interface {
	Ping(context.Context, *Connection) TraceFinisher
	Prepare(context.Context, *Connection, string) TraceFinisher
	Query(context.Context, *Connection, *Invocation, string) TraceFinisher
}

// TraceFinisher is a type that can finish traces.
type TraceFinisher interface {
	Finish(error)
}
