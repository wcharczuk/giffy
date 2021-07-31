package jobs

import (
	"context"
	"time"

	"github.com/blend/go-sdk/cron"
	"github.com/blend/go-sdk/db"
	"github.com/wcharczuk/giffy/server/model"
)

// FixContentRating fixes images that have been uploaded with a 0 content rating.
type FixContentRating struct {
	Model *model.Manager
}

// Name returns the job name.
func (fis FixContentRating) Name() string {
	return "fix_content_rating"
}

// Schedule returns the schedule.
func (fis FixContentRating) Schedule() cron.Schedule {
	return cron.Every(5 * time.Minute)
}

// Execute runs the job
func (fis FixContentRating) Execute(ctx context.Context) error {
	imageIDs := []int64{}

	err := fis.Model.Invoke(ctx).Query(`select id from image where content_rating = 0;`).Each(func(r db.Rows) error {
		var id int64
		err := r.Scan(&id)
		if err != nil {
			return err
		}

		imageIDs = append(imageIDs, id)
		return nil
	})
	if err != nil {
		return err
	}

	var image model.Image
	for _, id := range imageIDs {

		_, err = fis.Model.Invoke(ctx).Get(&image, id)
		if err != nil {
			return err
		}

		image.ContentRating = model.ContentRatingG
		_, err = fis.Model.Invoke(ctx).Update(&image)
		if err != nil {
			return err
		}
	}
	return nil
}
