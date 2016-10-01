package jobs

import (
	"database/sql"
	"time"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/spiffy"
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
func (fis FixContentRating) Execute(ct *chronometer.CancellationToken) error {
	imageIDs := []int64{}

	err := spiffy.DefaultDb().Query(`select id from image where content_rating = 0;`).Each(func(r *sql.Rows) error {
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

		ct.CheckCancellation()

		err = spiffy.DefaultDb().GetByID(&image, id)
		if err != nil {
			return err
		}

		ct.CheckCancellation()

		image.ContentRating = model.ContentRatingG
		err = spiffy.DefaultDb().Update(&image)
		if err != nil {
			return err
		}
	}
	return nil
}
