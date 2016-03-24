package jobs

import (
	"github.com/blendlabs/go-chronometer"
	"github.com/wcharczuk/giffy/server/model"
)

// OrphanedTags is a job that deletes orphaned tags
type DeleteOrphanedTags struct{}

// Name returns the job name
func (ot DeleteOrphanedTags) Name() string {
	return "delete_orphaned_tags"
}

// Schedule returns the job schedule.
func (ot DeleteOrphanedTags) Schedule() chronometer.Schedule {
	return chronometer.EveryHour()
}

// Execute runs the job
func (ot DeleteOrphanedTags) Execute(ct *chronometer.CancellationToken) error {
	return model.DeleteOrphanedTags(nil)
}
