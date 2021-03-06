package viewmodel

import (
	"fmt"

	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/model"
)

// NewImage creates a new viewmodel image.
func NewImage(img model.Image, cfg *config.Giffy) Image {
	if cfg.IsProduction() && len(cfg.GetCloudFrontDNS()) > 0 {
		return Image{
			Image:     img,
			S3ReadURL: fmt.Sprintf("https://%s/%s", cfg.GetCloudFrontDNS(), img.S3Key),
		}
	}
	return Image{
		Image:     img,
		S3ReadURL: fmt.Sprintf("https://s3-%s.amazonaws.com/%s/%s", cfg.Aws.GetRegion(), img.S3Bucket, img.S3Key),
	}
}

// WrapImages wraps the image list as a viewmodel image.
func WrapImages(images []model.Image, cfg *config.Giffy) []Image {
	output := make([]Image, len(images))
	for x := 0; x < len(images); x++ {
		output[x] = NewImage(images[x], cfg)
	}
	return output
}

// Image is a wrapper viewmodel for an image that injects the s3 read url.
type Image struct {
	model.Image `json:",inline"`
	S3ReadURL   string `json:"s3_read_url"`
}
