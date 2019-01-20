package jobs

import (
	"context"
	"time"

	"github.com/blend/go-sdk/cron"
	"github.com/wcharczuk/giffy/server/model"
)

// DeleteOrphanedTags is a job that deletes orphaned tags
type DeleteOrphanedTags struct {
	Model *model.Manager
}

// Name returns the job name
func (ot DeleteOrphanedTags) Name() string {
	return "delete_orphaned_tags"
}

// Schedule returns the job schedule.
func (ot DeleteOrphanedTags) Schedule() cron.Schedule {
	return cron.Every(1 * time.Minute)
}

// Execute runs the job
func (ot DeleteOrphanedTags) Execute(ctx context.Context) error {
	return ot.Model.DeleteOrphanedTags(ctx)
}
