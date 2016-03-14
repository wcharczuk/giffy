package model

import (
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/spiffy"
)

func TestQueryImages(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := spiffy.DefaultDb().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := createTestUser(tx)
	assert.Nil(err)

	i, err := createTestImage(u.ID, tx)
	assert.Nil(err)

	_, err = createTestTagForImage(u.ID, i.ID, "test", tx)
	assert.Nil(err)

	_, err = createTestTagForImage(u.ID, i.ID, "foo", tx)
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

	_, err = createTestTagForImage(u.ID, i.ID, "test", tx)
	assert.Nil(err)

	_, err = createTestTagForImage(u.ID, i.ID, "foo", tx)
	assert.Nil(err)

	_, err = createTestTagForImage(u.ID, i.ID, "bar", tx)
	assert.Nil(err)

	baz, err := createTestTagForImage(u.ID, i.ID, "baz", tx)
	assert.Nil(err)

	biz, err := createTestTagForImage(u.ID, i.ID, "biz", tx)
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
