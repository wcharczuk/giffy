package web

import "github.com/blend/go-sdk/logger"

var (
	// BufferPool is a shared sync.Pool of bytes.Buffer instances.
	BufferPool = logger.NewBufferPool(32)
)
