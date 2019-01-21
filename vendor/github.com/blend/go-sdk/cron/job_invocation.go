package cron

import (
	"context"
	"time"
)

// JobInvocation is metadata for a job invocation (or instance of a job running).
type JobInvocation struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	StartTime time.Time          `json:"startTime"`
	Timeout   time.Time          `json:"timeout"`
	Err       error              `json:"err"`
	Elapsed   time.Duration      `json:"elapsed"`
	Context   context.Context    `json:"-"`
	Cancel    context.CancelFunc `json:"-"`
}
