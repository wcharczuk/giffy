package jobs

import (
	"context"
	"database/sql"

	"github.com/blend/go-sdk/cron"
	"github.com/blend/go-sdk/exception"
	"github.com/wcharczuk/giffy/server/model"
)

// CleanTagValues is a job that cleans tags of punctuation etc.
type CleanTagValues struct{}

// Name returns the job name
func (ot CleanTagValues) Name() string {
	return "clean_tag_values"
}

// Schedule returns the job schedule.
func (ot CleanTagValues) Schedule() cron.Schedule {
	return cron.EveryHour()
}

// Execute runs the job
func (ot CleanTagValues) Execute(ctx context.Context) error {
	tx, err := model.DB().Begin()
	if err != nil {
		return err
	}
	err = ot.ExecuteInTx(ctx, tx)
	if err != nil {
		return exception.Wrap(tx.Rollback())
	}
	return exception.Wrap(tx.Commit())
}

// ExecuteInTx runs the job in a transaction
func (ot CleanTagValues) ExecuteInTx(ctx context.Context, tx *sql.Tx) error {
	allTags, err := model.GetAllTags(tx)
	if err != nil {
		return err
	}

	for _, tag := range allTags {
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
