package cron

// Status is a status object
type Status struct {
	Jobs    []*JobScheduler
	Running map[string]*JobInvocation
}
