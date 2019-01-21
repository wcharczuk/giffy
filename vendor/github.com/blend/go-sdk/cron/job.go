package cron

import "context"

// Job is an interface types can satisfy to be loaded into the JobManager.
type Job interface {
	Name() string
	Execute(ctx context.Context) error
}
