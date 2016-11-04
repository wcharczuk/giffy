package logger

import (
	"io"
	"os"
	"sync"
)

// NewStdOutMultiWriterFromEnvironment creates a new multiplexed stdout writer.
func NewStdOutMultiWriterFromEnvironment() io.WriteCloser {
	primary := os.Stdout
	filePath := os.Getenv(EnvironmentVariableLogOutFile)
	if len(filePath) > 0 {
		secondary, err := NewFileWriterFromEnvironmentVars(
			EnvironmentVariableLogOutFile,
			EnvironmentVariableLogOutArchiveCompress,
			EnvironmentVariableLogOutMaxSizeBytes,
			EnvironmentVariableLogOutMaxArchive,
		)
		if err != nil {
			panic(err)
		}
		return NewMultiWriter(primary, secondary)
	}
	return NewSyncWriter(primary)
}

// NewStdErrMultiWriterFromEnvironment creates a new multiplexed stderr writer.
func NewStdErrMultiWriterFromEnvironment() io.WriteCloser {
	primary := os.Stderr
	filePath := os.Getenv(EnvironmentVariableLogErrFile)
	if len(filePath) > 0 {
		secondary, err := NewFileWriterFromEnvironmentVars(
			EnvironmentVariableLogErrFile,
			EnvironmentVariableLogErrArchiveCompress,
			EnvironmentVariableLogErrMaxSizeBytes,
			EnvironmentVariableLogErrMaxArchive,
		)
		if err != nil {
			panic(err)
		}
		return NewMultiWriter(primary, secondary)
	}
	return NewSyncWriter(primary)
}

// NewMultiWriter creates a new MultiWriter that wraps an array of writers.
func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{
		writers: writers,
	}
}

// MultiWriter writes to many writers at once.
type MultiWriter struct {
	writers []io.Writer
}

func (mw MultiWriter) Write(buffer []byte) (int, error) {
	var written int
	var err error

	for x := 0; x < len(mw.writers); x++ {
		if mw.writers[x] != nil {
			written, err = mw.writers[x].Write(buffer)
		}
	}
	return written, err
}

// Close closes all of the inner writers (if they are io.WriteClosers).
func (mw MultiWriter) Close() error {
	var err error
	var closeErr error
	for x := 0; x < len(mw.writers); x++ {
		if typed, isTyped := mw.writers[x].(io.Closer); isTyped {
			closeErr = typed.Close()
			if closeErr != nil {
				err = closeErr
			}
		}
	}
	return err
}

// NewSyncWriteCloser returns a new sync write closer.
func NewSyncWriteCloser(innerWriter io.WriteCloser) *SyncWriteCloser {
	return &SyncWriteCloser{
		innerWriter: innerWriter,
		syncRoot:    &sync.Mutex{},
	}
}

// SyncWriteCloser wraps a write closer.
type SyncWriteCloser struct {
	innerWriter io.WriteCloser
	syncRoot    *sync.Mutex
}

// Write writes the given bytes to the inner writer.
func (sw *SyncWriteCloser) Write(buffer []byte) (int, error) {
	sw.syncRoot.Lock()
	defer sw.syncRoot.Unlock()

	return sw.innerWriter.Write(buffer)
}

// Close closes the file handle.
func (sw *SyncWriteCloser) Close() error {
	return sw.innerWriter.Close()
}

// NewSyncWriter returns a new interlocked writer.
func NewSyncWriter(innerWriter io.Writer) io.WriteCloser {
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

// Close is a no-op.
func (sw SyncWriter) Close() error { return nil }
