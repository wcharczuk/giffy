package jobs

import (
	"database/sql"
	"io/ioutil"
	"net/http"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/model"
)

// FixImageSizes fetches an image and updates the file size.
type FixImageSizes struct{}

// Name returns the job name.
func (fis FixImageSizes) Name() string {
	return "fix_image_sizes"
}

// Schedule returns the schedule.
func (fis FixImageSizes) Schedule() chronometer.Schedule {
	return chronometer.EveryHour()
}

// Execute runs the job
func (fis FixImageSizes) Execute(ct *chronometer.CancellationToken) error {
	imageIDs := []int64{}

	err := spiffy.DefaultDb().Query(`select id from image where file_size = 0;`).Each(func(r *sql.Rows) error {
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

		if ct.ShouldCancel() {
			return ct.Cancel()
		}

		err = spiffy.DefaultDb().GetByID(&image, id)
		if err != nil {
			return err
		}

		size, err := fis.getImageSize(image.S3ReadURL)
		if err != nil {
			return err
		}

		if ct.ShouldCancel() {
			return ct.Cancel()
		}

		image.FileSize = size
		err = spiffy.DefaultDb().Update(&image)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fis FixImageSizes) getImageSize(s3ReadURL string) (int, error) {
	res, err := http.Get(s3ReadURL)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	return len(bytes), nil
}
