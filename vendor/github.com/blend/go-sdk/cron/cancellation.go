package cron

import "context"

// IsContextCancelled check if a job is cancelled
func IsContextCancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
