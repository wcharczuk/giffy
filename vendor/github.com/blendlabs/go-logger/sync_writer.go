package logger

import (
	"io"
	"sync"
)

// NewSyncWriter returns a new interlocked writer.
func NewSyncWriter(innerWriter io.Writer) io.Writer {
	if innerWriter == nil {
		return nil
	}
	return &SyncWriter{
		innerWriter: innerWriter,
		syncRoot:    &sync.Mutex{},
	}
}

// SyncWriter is a writer that serializes access to the Write() method.
type SyncWriter struct {
	innerWriter io.Writer
	syncRoot    *sync.Mutex
}

// Write writes the given bytes to the inner writer.
func (sw *SyncWriter) Write(buffer []byte) (int, error) {
	sw.syncRoot.Lock()
	defer sw.syncRoot.Unlock()
	return sw.innerWriter.Write(buffer)
}
