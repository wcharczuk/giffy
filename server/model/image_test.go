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

	if existing == nil || existing.IsZero() {
		v := NewImageTagVote(imageID, tag.ID, userID, time.Now().UTC(), 1, 0)
		err = spiffy.DefaultDb().CreateInTransaction(v, tx)
	}
	return tag, err
}

func createTestImage(userID int64, tx *sql.Tx) (*Image, error) {
	i := NewImage()
	i.CreatedBy = userID
	i.UpdatedBy = userID
	i.UpdatedUTC = time.Now().UTC()
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
	spiffy.DefaultDb().IsolateToTransaction(tx)
	defer spiffy.DefaultDb().ReleaseIsolation()

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
}
