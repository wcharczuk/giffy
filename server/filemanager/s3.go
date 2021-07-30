package filemanager

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	exception "github.com/blend/go-sdk/ex"
	"github.com/blend/go-sdk/uuid"
	"github.com/wcharczuk/giffy/server/config"
)

// Location is an s3 location.
type Location struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

// FileType is metadata for a file.
type FileType struct {
	Extension string
	MimeType  string
}

// New returns a new filemanager.
func New(s3Bucket string, cfg *config.Aws) *FileManager {
	awsConfig := &aws.Config{
		Region:      aws.String(cfg.GetRegion()),
		Credentials: credentials.NewStaticCredentials(cfg.GetAccessKeyID(), cfg.GetSecretAccessKey(), cfg.GetSecurityToken()),
	}
	awsSession := session.New(awsConfig)
	return &FileManager{
		s3Bucket: s3Bucket,
		config:   cfg,
		session:  awsSession,
		uploader: s3manager.NewUploader(awsSession),
		s3Client: s3.New(awsSession),
	}
}

// FileManager is a helper for s3 related operations.
type FileManager struct {
	s3Bucket       string
	config         *config.Aws
	session        *session.Session
	s3Client       *s3.S3
	uploader       *s3manager.Uploader
	mockFiles      map[string][]byte
	mockingEnabled bool
}

// NewLocationFromKey makes a new location from a key.
func (fm *FileManager) NewLocationFromKey(key string) *Location {
	return &Location{
		Bucket: fm.s3Bucket,
		Key:    key,
	}
}

//UploadFile uploads a file.
func (fm *FileManager) UploadFile(uploadFile io.Reader, fileType FileType) (*Location, error) {
	return fm.UploadFileToBucket(fm.s3Bucket, uploadFile, fileType)
}

//UploadFileToBucket uploads a file to a given location.
func (fm *FileManager) UploadFileToBucket(bucket string, uploadFile io.Reader, fileType FileType) (*Location, error) {
	fileKey := uuid.V4().String() + fileType.Extension

	var err error
	if fm.mockingEnabled {
		err = fm.mockUploadFile(bucket, fileKey, uploadFile)
	} else {
		_, err = fm.uploader.Upload(&s3manager.UploadInput{
			Bucket:      &bucket,
			Key:         &fileKey,
			Body:        uploadFile,
			ContentType: &fileType.MimeType,
			ACL:         aws.String("public-read"),
		})
	}

	if err != nil {
		return nil, err
	}

	return &Location{
		Bucket: bucket,
		Key:    fileKey,
	}, nil
}

func (fm *FileManager) mockUploadFile(bucket, key string, body io.Reader) error {
	contents, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	if fm.mockFiles == nil {
		fm.mockFiles = map[string][]byte{}
	}
	fm.mockFiles[bucket+key] = contents
	return nil
}

//DeleteFile deletes a file from s3.
func (fm *FileManager) DeleteFile(fileLocation *Location) error {
	if fm.mockingEnabled {
		if _, hasFile := fm.mockFiles[fileLocation.Bucket+fileLocation.Key]; hasFile {
			delete(fm.mockFiles, fileLocation.Bucket+fileLocation.Key)
			return nil
		}
		return exception.New("File not found")
	}

	deletionParams := &s3.DeleteObjectInput{
		Bucket: &fileLocation.Bucket,
		Key:    &fileLocation.Key,
	}
	_, err := fm.s3Client.DeleteObject(deletionParams)
	return err
}

// GetFile gets a file.
func (fm *FileManager) GetFile(fileLocation *Location) (io.Reader, error) {
	if fm.mockingEnabled {
		if file, hasFile := fm.mockFiles[fileLocation.Bucket+fileLocation.Key]; hasFile {
			return bytes.NewBuffer(file), nil
		}
		return nil, exception.New("File not found.")
	}

	downloadParams := &s3.GetObjectInput{
		Bucket: &fileLocation.Bucket,
		Key:    &fileLocation.Key,
	}
	res, err := fm.s3Client.GetObject(downloadParams)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

// Mock mocks file uploads.
func (fm *FileManager) Mock() {
	fm.mockingEnabled = true
}

// ReleaseMock releases file upload mocking.
func (fm *FileManager) ReleaseMock() {
	fm.mockingEnabled = false
	fm.mockFiles = map[string][]byte{}
}
