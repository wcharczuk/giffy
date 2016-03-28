package jobs

import (
	"database/sql"

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
	return ot.ExecuteInTransaction(ct, nil)
}

// ExecuteInTransaction runs the job in a transaction
func (ot CleanTagValues) ExecuteInTransaction(ct *chronometer.CancellationToken, tx *sql.Tx) error {
	allTags, err := model.GetAllTags(tx)
	if err != nil {
		return err
	}

	for _, tag := range allTags {
		if ct.ShouldCancel() {
			return ct.Cancel()
		}

		tag.TagValue = model.CleanTagValue(tag.TagValue)

		existingTag, err := model.GetTagByValue(tag.TagValue, tx)
		if err != nil {
			return err
		}
		if existingTag.IsZero() {
			err = spiffy.DefaultDb().UpdateInTransaction(&tag, tx)
			if err != nil {
				return err
			}
		} else {
			err = model.MergeTags(tag.ID, existingTag.ID, tx)
			if err != nil {
				return err
			}
		}

	}
	return nil
}
