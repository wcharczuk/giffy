package jobs

import (
	"context"
	"database/sql"
	"time"

	"github.com/blendlabs/go-chronometer"
	"github.com/wcharczuk/giffy/server/model"
)

// FixContentRating fixes images that have been uploaded with a 0 content rating.
type FixContentRating struct{}

// Name returns the job name.
func (fis FixContentRating) Name() string {
	return "fix_content_rating"
}

// Schedule returns the schedule.
func (fis FixContentRating) Schedule() chronometer.Schedule {
	return chronometer.Every(5 * time.Minute)
}

// Execute runs the job
func (fis FixContentRating) Execute(ctx context.Context) error {
	imageIDs := []int64{}

	err := model.DB().Query(`select id from image where content_rating = 0;`).Each(func(r *sql.Rows) error {
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

		err = model.DB().Get(&image, id)
		if err != nil {
			return err
		}

		image.ContentRating = model.ContentRatingG
		err = model.DB().Update(&image)
		if err != nil {
			return err
		}
	}
	return nil
}
