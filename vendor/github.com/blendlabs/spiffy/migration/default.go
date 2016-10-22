package migration

import "sync"

var (
	defaultRunner     = &Runner{}
	defaultRunnerLock sync.RWMutex
)

// Register adds a process to the default suite.
func Register(m ...Migration) error {
	defaultRunnerLock.Lock()
	defer defaultRunnerLock.Unlock()
	defaultRunner.addMigrations(m...)
	return nil
}

// Run passes the default suite to the handler method. It acquires a read lock wrapping the action.
func Run(handler func(Migration) error) error {
	defaultRunnerLock.RLock()
	defer defaultRunnerLock.RUnlock()
	return handler(defaultRunner)
}
