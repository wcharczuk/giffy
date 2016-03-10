package filecache

import (
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/wcharczuk/giffy/server/core"
)

var (
	//S3Bucket is the bucket we save images to.
	S3Bucket = os.Getenv("S3_BUCKET")

	//AWSKey is the aws key.
	AWSKey = os.Getenv("AWS_ACCESS_KEY_ID")

	//AWSSecret is the aws secret.
	AWSSecret = os.Getenv("AWS_SECRET_ACCESS_KEY")

	s3Client = s3.New(session.New())
	uploader = s3manager.NewUploader(session.New())
)

// Location is an s3 location.
type Location struct {
	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
	AwsKey    string `json:"aws_key"`
	AwsSecret string `json:"aws_secret"`
}

// NewLocationFromKey makes a new location from a key.
func NewLocationFromKey(key string) *Location {
	return &Location{
		Bucket:    S3Bucket,
		Key:       key,
		AwsKey:    AWSKey,
		AwsSecret: AWSSecret,
	}
}

// FileType is metadata for a file.
type FileType struct {
	Extension string
	MimeType  string
}

//UploadFile uploads a file.
func UploadFile(uploadFile io.Reader, fileType FileType) (*Location, error) {
	return UploadFileToBucket(S3Bucket, uploadFile, fileType)
}

//UploadFileToBucket uploads a file to a given location.
func UploadFileToBucket(bucket string, uploadFile io.Reader, fileType FileType) (*Location, error) {
	fileKey := core.UUIDv4().ToShortString() + "." + fileType.Extension

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      &bucket,
		Key:         &fileKey,
		Body:        uploadFile,
		ContentType: &fileType.MimeType,
	})

	if err != nil {
		return nil, err
	}

	return &Location{
		Bucket:    bucket,
		Key:       fileKey,
		AwsKey:    AWSKey,
		AwsSecret: AWSSecret,
	}, nil
}

//DeleteFile deletes a file from s3.
func DeleteFile(fileLocation *Location) error {
	deletionParams := &s3.DeleteObjectInput{
		Bucket: &fileLocation.Bucket,
		Key:    &fileLocation.Key,
	}
	_, err := s3Client.DeleteObject(deletionParams)
	return err
}

//GetFile gets a file.
func GetFile(fileLocation *Location) (io.Reader, error) {
	downloadParams := &s3.GetObjectInput{
		Bucket: &fileLocation.Bucket,
		Key:    &fileLocation.Key,
	}
	res, err := s3Client.GetObject(downloadParams)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}
