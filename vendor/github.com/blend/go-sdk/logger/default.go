package logger

import "sync"

var (
	_default     *Logger
	_defaultLock sync.Mutex
)

// Default returnes a default Agent singleton.
func Default() *Logger {
	return _default
}

// SetDefault sets the diagnostics singleton.
func SetDefault(log *Logger) {
	_defaultLock.Lock()
	defer _defaultLock.Unlock()
	_default = log
}
