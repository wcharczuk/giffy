package viewmodel

import (
	"fmt"

	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/model"
)

// NewImage creates a new viewmodel image.
func NewImage(img model.Image, cfg *config.Giffy) Image {
	if cfg.Meta.IsProdlike() && len(cfg.CloudFrontDNS) > 0 {
		return Image{
			Image:     img,
			S3ReadURL: fmt.Sprintf("https://%s/%s", cfg.CloudFrontDNS, img.S3Key),
		}
	}
	return Image{
		Image:     img,
		S3ReadURL: fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", img.S3Bucket, cfg.Aws.RegionOrDefault(), img.S3Key),
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
