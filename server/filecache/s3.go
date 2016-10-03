package filecache

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	exception "github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/giffy/server/core"
)

var (
	//S3Bucket is the bucket we save images to.
	S3Bucket = os.Getenv("AWS_S3_BUCKET")

	//AWSKey is the aws key.
	AWSKey = os.Getenv("AWS_ACCESS_KEY_ID")

	//AWSSecret is the aws secret.
	AWSSecret = os.Getenv("AWS_SECRET_ACCESS_KEY")

	// AWSRegion is the default region
	AWSRegion = util.OptionalString(os.Getenv("AWS_REGION"))

	s3Client = s3.New(session.New(&aws.Config{Region: AWSRegion}))
	uploader = s3manager.NewUploader(session.New(&aws.Config{Region: AWSRegion}))
)

var (
	mockFiles      = map[string][]byte{}
	mockingEnabled = false
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
	fileKey := core.UUIDv4().ToShortString() + fileType.Extension

	var err error
	if mockingEnabled {
		err = mockUploadFile(bucket, fileKey, uploadFile)
	} else {
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket:      &bucket,
			Key:         &fileKey,
			Body:        uploadFile,
			ContentType: &fileType.MimeType,
			ACL:         util.OptionalString("public-read"),
		})
	}

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

func mockUploadFile(bucket, key string, body io.Reader) error {
	contents, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	mockFiles[bucket+key] = contents
	return nil
}

//DeleteFile deletes a file from s3.
func DeleteFile(fileLocation *Location) error {
	if mockingEnabled {
		if _, hasFile := mockFiles[fileLocation.Bucket+fileLocation.Key]; hasFile {
			delete(mockFiles, fileLocation.Bucket+fileLocation.Key)
			return nil
		}
		return exception.New("File not found")
	}

	deletionParams := &s3.DeleteObjectInput{
		Bucket: &fileLocation.Bucket,
		Key:    &fileLocation.Key,
	}
	_, err := s3Client.DeleteObject(deletionParams)
	return err
}

// GetFile gets a file.
func GetFile(fileLocation *Location) (io.Reader, error) {
	if mockingEnabled {
		if file, hasFile := mockFiles[fileLocation.Bucket+fileLocation.Key]; hasFile {
			return bytes.NewBuffer(file), nil
		}
		return nil, exception.New("File not found.")
	}

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

// Mock mocks file uploads.
func Mock() {
	mockingEnabled = true
}

// ReleaseMock releases file upload mocking.
func ReleaseMock() {
	mockingEnabled = false
	mockFiles = map[string][]byte{}
}
