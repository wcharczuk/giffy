package chronometer

import "time"

// TaskListener is an event hook for jobs running
type TaskListener func(taskName string, elapsed time.Duration, err error)
