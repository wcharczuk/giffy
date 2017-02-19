package chronometer

// Job is an interface structs can satisfy to be loaded into the JobManager.
type Job interface {
	Name() string
	Schedule() Schedule
	Execute(ct *CancellationToken) error
}

// ShowMessagesProvider is a type that enables or disables messages.
type ShowMessagesProvider interface {
	ShowMessages() bool
}
