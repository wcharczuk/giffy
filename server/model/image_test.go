package model

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
)

func createTestTag(userID, imageID int64, tagValue string, tx *sql.Tx) (*Tag, error) {
	tag := NewTag()
	tag.CreatedBy = userID
	tag.TagValue = tagValue
	err := spiffy.DefaultDb().CreateInTransaction(tag, tx)

	if err != nil {
		return nil, err
	}

	existing, existingErr := GetImageTagVote(imageID, tag.ID, tx)
	if existingErr != nil {
		return nil, existingErr
	}

	if existing.IsZero() {
		v := NewImageTagVote(imageID, tag.ID, userID, time.Now().UTC(), 1, 0)
		err = spiffy.DefaultDb().CreateInTransaction(v, tx)
	}
	return tag, err
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
	err := spiffy.DefaultDb().CreateInTransaction(i, tx)
	return i, err
}

func createTestUser(tx *sql.Tx) (*User, error) {
	u := NewUser(fmt.Sprintf("__test_user_%s__", core.UUIDv4().ToShortString()))
	u.FirstName = "Test"
	u.LastName = "User"
	err := spiffy.DefaultDb().CreateInTransaction(u, tx)
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

	_, err = createTestTag(u.ID, i.ID, "test", tx)
	assert.Nil(err)

	_, err = createTestTag(u.ID, i.ID, "foo", tx)
	assert.Nil(err)

	images, err := QueryImages("test", tx)
	assert.Nil(err)
	assert.NotEmpty(images)

	firstImage := images[0]
	assert.False(firstImage.IsZero())
	assert.NotNil(firstImage.CreatedByUser)
	assert.NotEmpty(firstImage.Tags)
}

func TestGetImagesByID(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := spiffy.DefaultDb().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := createTestUser(tx)
	assert.Nil(err)

	i, err := createTestImage(u.ID, tx)
	assert.Nil(err)

	_, err = createTestImage(u.ID, tx)
	assert.Nil(err)

	_, err = createTestImage(u.ID, tx)
	assert.Nil(err)

	_, err = createTestTag(u.ID, i.ID, "test", tx)
	assert.Nil(err)

	_, err = createTestTag(u.ID, i.ID, "foo", tx)
	assert.Nil(err)

	_, err = createTestTag(u.ID, i.ID, "bar", tx)
	assert.Nil(err)

	baz, err := createTestTag(u.ID, i.ID, "baz", tx)
	assert.Nil(err)

	biz, err := createTestTag(u.ID, i.ID, "biz", tx)
	assert.Nil(err)

	err = SetTagVotes(i.ID, baz.ID, 100, 3, tx)
	assert.Nil(err)

	err = SetTagVotes(i.ID, biz.ID, 1000, 30, tx)
	assert.Nil(err)

	images, err := GetAllImages(tx)
	assert.Nil(err)
	assert.NotNil(images)
	assert.NotEmpty(images)

	for _, returnedImage := range images {
		if i.ID == returnedImage.ID {
			assert.NotEmpty(returnedImage.Tags)
			assert.Len(returnedImage.Tags, 5)
		}
	}
}
