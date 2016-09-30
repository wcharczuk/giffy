package web

import "github.com/blendlabs/go-logger"

var (
	// BufferPool is a shared sync.Pool of bytes.Buffer instances.
	BufferPool = logger.NewBufferPool(32)
)
