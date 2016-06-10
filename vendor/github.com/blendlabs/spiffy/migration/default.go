package migration

import "sync"

var (
	defaultRunner     = &Runner{IsDefault: true}
	defaultRunnerLock sync.RWMutex
)

// Register adds a process to the default suite.
func Register(m Migration) {
	defaultRunnerLock.Lock()
	defer defaultRunnerLock.Unlock()
	defaultRunner.Migrations = append(defaultRunner.Migrations, m)
}

// Default passes the default suite to the action method. It acquires a read lock wrapping the action.
func Default(action func(Migration) error) error {
	defaultRunnerLock.RLock()
	defer defaultRunnerLock.RUnlock()
	return action(defaultRunner)
}
