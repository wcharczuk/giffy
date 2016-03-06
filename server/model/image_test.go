package model

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
)

func createTestImageTag(userID int64, imageID int, tagValue string, tx *sql.Tx) (*ImageTag, error) {
	it := NewImageTag()
	it.ImageID = imageID
	it.CreatedBy = userID
	it.TagValue = tagValue
	err := spiffy.DefaultDb().CreateInTransaction(&it, tx)
	return it, err
}

func createTestImage(userID int64, tx *sql.Tx) (*Image, error) {
	i := NewImage()
	i.CreatedBy = userID
	i.Extension = "gif"
	i.Width = 720
	i.Height = 480
	i.S3Bucket = core.UUIDv4().ToShortString()
	i.S3Key = core.UUIDv4().ToShortString()
	i.S3ReadURL = fmt.Sprintf("https://s3.amazonaws.com/%s/%s", i.S3Bucket, i.S3Key)
	i.MD5 = core.UUIDv4()
	i.DisplayName = "Test Image"
	err := spiffy.DefaultDb().CreateInTransaction(&i, tx)
	return i, err
}

func createTestUser(tx *sql.Tx) (*User, error) {
	u := NewUser()
	u.Username = "__test_user__"
	u.FirstName = "Test"
	u.LastName = "User"
	err := spiffy.DefaultDb().CreateInTransaction(&u, tx)
	return u, err
}

func TestQueryImages(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := spiffy.DefaultDb().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := createTestUser(tx)
	assert.Nil(err)

	i, err := createTestImage(u.ID, tx)
	assert.Nil(err)

	_, err = createTestImageTag(u.ID, image.ID, "test", tx)
	assert.Nil(err)

	_, err = createTestImageTag(u.ID, image.ID, "foo", tx)
	assert.Nil(err)

	images, err := QueryImages("test", tx)
	assert.Nil(err)
	assert.NotEmpty(images)
}
