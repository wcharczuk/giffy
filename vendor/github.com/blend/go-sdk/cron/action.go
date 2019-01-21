package cron

import "context"

// Action is an function that can be run as a task
type Action func(ctx context.Context) error
