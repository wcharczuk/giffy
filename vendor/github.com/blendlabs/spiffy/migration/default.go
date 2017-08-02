package migration

var (
	defaultSuite = &Suite{}
)

// Register adds a process to the default suite.
func Register(m ...Migration) error {
	defaultSuite.Add(m...)
	return nil
}

// SetDefault sets the default.
func SetDefault(suite *Suite) {
	defaultSuite = suite
}

// Default returns the default migration suite.
func Default() Migration {
	return defaultSuite
}
