package jobs

import (
	"context"

	"github.com/blend/go-sdk/cron"
	"github.com/wcharczuk/giffy/server/model"
)

// CleanTagValues is a job that cleans tags of punctuation etc.
type CleanTagValues struct {
	Model *model.Manager
}

// Name returns the job name
func (ot CleanTagValues) Name() string {
	return "clean_tag_values"
}

// Schedule returns the job schedule.
func (ot CleanTagValues) Schedule() cron.Schedule {
	return cron.EveryHour()
}

// ExecuteInTx runs the job in a transaction
func (ot CleanTagValues) ExecuteInTx(ctx context.Context) error {
	allTags, err := ot.Model.GetAllTags(ctx)
	if err != nil {
		return err
	}

	for _, tag := range allTags {
		newTagValue := model.CleanTagValue(tag.TagValue)
		if newTagValue != tag.TagValue {
			existingTag, err := ot.Model.GetTagByValue(ctx, newTagValue)
			if err != nil {
				return err
			}
			if existingTag.IsZero() {
				err = ot.Model.SetTagValue(ctx, tag.ID, newTagValue)
				if err != nil {
					return err
				}
			} else {
				err = ot.Model.MergeTags(ctx, tag.ID, existingTag.ID)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
