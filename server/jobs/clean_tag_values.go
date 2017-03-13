package jobs

import (
	"database/sql"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/model"
)

// CleanTagValues is a job that cleans tags of punctuation etc.
type CleanTagValues struct{}

// Name returns the job name
func (ot CleanTagValues) Name() string {
	return "clean_tag_values"
}

// Schedule returns the job schedule.
func (ot CleanTagValues) Schedule() chronometer.Schedule {
	return chronometer.EveryHour()
}

// Execute runs the job
func (ot CleanTagValues) Execute(ct *chronometer.CancellationToken) error {
	tx, err := spiffy.DB().Begin()
	if err != nil {
		return err
	}
	err = ot.ExecuteInTx(ct, tx)
	if err != nil {
		return exception.Wrap(tx.Rollback())
	}
	return exception.Wrap(tx.Commit())
}

// ExecuteInTx runs the job in a transaction
func (ot CleanTagValues) ExecuteInTx(ct *chronometer.CancellationToken, tx *sql.Tx) error {
	allTags, err := model.GetAllTags(tx)
	if err != nil {
		return err
	}

	for _, tag := range allTags {
		ct.CheckCancellation()

		newTagValue := model.CleanTagValue(tag.TagValue)
		if newTagValue != tag.TagValue {
			existingTag, err := model.GetTagByValue(newTagValue, tx)
			if err != nil {
				return err
			}
			if existingTag.IsZero() {
				err = model.SetTagValue(tag.ID, newTagValue, tx)
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
	}
	return nil
}
