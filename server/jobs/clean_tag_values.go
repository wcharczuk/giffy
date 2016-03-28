package jobs

import (
	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/model"
)

// CleanTagValues is a job that deletes orphaned tags
type CleanTagValues struct{}

// Name returns the job name
func (ot CleanTagValues) Name() string {
	return "delete_orphaned_tags"
}

// Schedule returns the job schedule.
func (ot CleanTagValues) Schedule() chronometer.Schedule {
	return nil //can only be run on demand.
}

// Execute runs the job
func (ot CleanTagValues) Execute(ct *chronometer.CancellationToken) error {
	allTags, err := model.GetAllTags(nil)
	if err != nil {
		return err
	}

	for _, tag := range allTags {
		tag.TagValue = model.CleanTagValue(tag.TagValue)
		err = spiffy.DefaultDb().Update(&tag)
		if err != nil {
			return err
		}
	}
	return nil
}
