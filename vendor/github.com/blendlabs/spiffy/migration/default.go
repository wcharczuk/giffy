package migration

var (
	defaultGroup = &Group{}
)

// RegisterDefault adds migrations to the default group.
func RegisterDefault(m ...Migration) error {
	defaultGroup.Add(m...)
	return nil
}

// SetDefault sets the default group.
func SetDefault(group *Group) {
	defaultGroup = group
}

// Default returns the default migration suite.
func Default() *Group {
	return defaultGroup
}
