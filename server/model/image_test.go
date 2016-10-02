package model

import (
	"fmt"
	"sort"
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-util"
	"github.com/blendlabs/go-util/linq"
	"github.com/wcharczuk/giffy/server/core"
)

func TestGetAllImages(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	_, err = CreateTestTagForImageWithVote(u.ID, i.ID, core.UUIDv4().ToShortString(), tx)
	assert.Nil(err)

	_, err = CreateTestTagForImageWithVote(u.ID, i.ID, core.UUIDv4().ToShortString(), tx)
	assert.Nil(err)

	allImages, err := GetAllImages(tx)
	assert.Nil(err)
	assert.NotEmpty(allImages)

	firstResult := linq.First(allImages, NewImagePredicate(func(image Image) bool {
		return image.ID == i.ID
	}))

	assert.NotNil(firstResult)

	firstImage, resultIsImage := firstResult.(Image)
	assert.True(resultIsImage)
	assert.NotNil(firstImage)
	assert.False(firstImage.IsZero())
	assert.NotNil(firstImage.CreatedByUser)
	assert.NotEmpty(firstImage.Tags)
}

func TestGetRandomImages(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	for x := 0; x < 10; x++ {
		_, err = CreateTestImage(u.ID, tx)
		assert.Nil(err)
	}

	images, err := GetRandomImages(5, tx)
	assert.Nil(err)
	assert.Len(images, 5)
}

func TestGetImageByID(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	_, err = CreateTestTagForImageWithVote(u.ID, i.ID, core.UUIDv4().ToShortString(), tx)
	assert.Nil(err)

	verify, err := GetImageByID(i.ID, tx)
	assert.Nil(err)
	assert.NotNil(verify)
	assert.False(verify.IsZero())
	assert.Equal(i.ID, verify.ID)
	assert.Equal(i.UUID, verify.UUID)

	assert.NotNil(verify.CreatedByUser)
	assert.NotEmpty(verify.Tags)
}

func TestGetImageByIDNotFound(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	verify, err := GetImageByID(-1, tx)
	assert.Nil(err)
	assert.NotNil(verify)
	assert.True(verify.IsZero())
}

func TestUpdateImageDisplayName(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	err = UpdateImageDisplayName(i.ID, fmt.Sprintf("not %s", i.DisplayName), tx)
	assert.Nil(err)

	verify, err := GetImageByID(i.ID, tx)
	assert.Nil(err)
	assert.Equal(fmt.Sprintf("not %s", i.DisplayName), verify.DisplayName)
}

func TestDeleteImageByID(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	_, err = CreateTestTagForImageWithVote(u.ID, i.ID, core.UUIDv4().ToShortString(), tx)
	assert.Nil(err)

	err = DeleteImageByID(i.ID, tx)
	assert.Nil(err)

	verify, err := GetImageByID(i.ID, tx)
	assert.Nil(err)
	assert.True(verify.IsZero())
}

func TestImageMD5Check(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	verify, err := GetImageByMD5(i.MD5, tx)
	assert.Nil(err)
	assert.False(verify.IsZero())
}

func TestSearchImages(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	i4, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	i3, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	i2, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	_, err = CreateTestTagForImageWithVote(u.ID, i4.ID, "not_foo_bar", tx)
	assert.Nil(err)

	_, err = CreateTestTagForImageWithVote(u.ID, i3.ID, "__test_foo_bar", tx)
	assert.Nil(err)

	_, err = CreateTestTagForImageWithVote(u.ID, i2.ID, "__test_bar", tx)
	assert.Nil(err)

	_, err = CreateTestTagForImageWithVote(u.ID, i.ID, "__test", tx)
	assert.Nil(err)

	images, err := SearchImages("__test", ContentRatingFilterDefault, tx)
	assert.Nil(err)
	assert.NotEmpty(images)
	assert.Len(images, 3)

	firstImage := images[0]
	assert.NotNil(firstImage)
	assert.False(firstImage.IsZero())

	assert.Equal(i.ID, firstImage.ID)

	assert.NotNil(firstImage.CreatedByUser)
	assert.NotEmpty(firstImage.Tags)
}

func TestSearchImagesRandom(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	i4, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	i3, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	i2, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	_, err = CreateTestTagForImageWithVote(u.ID, i4.ID, "not__test_foo_bar", tx)
	assert.Nil(err)

	_, err = CreateTestTagForImageWithVote(u.ID, i3.ID, "__test_foo_bar", tx)
	assert.Nil(err)

	_, err = CreateTestTagForImageWithVote(u.ID, i2.ID, "__test_bar", tx)
	assert.Nil(err)

	_, err = CreateTestTagForImageWithVote(u.ID, i.ID, "__test", tx)
	assert.Nil(err)

	images, err := SearchImagesWeightedRandom("__test", ContentRatingFilterDefault, 2, tx)
	assert.Nil(err)
	assert.NotNil(images)
	assert.NotEmpty(images)
}

func TestSearchImagesBestResult(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	i4, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	i3, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	i2, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	tag4value := fmt.Sprintf("__test_foo_bar_%s", util.RandomString(4))
	tag3value := fmt.Sprintf("__test_bar_%s", util.RandomString(4))
	tag2value := fmt.Sprintf("__test_foo_%s", util.RandomString(4))
	tag1value := fmt.Sprintf("__test_%s", util.RandomString(4))

	_, err = CreateTestTagForImageWithVote(u.ID, i4.ID, tag4value, tx)
	assert.Nil(err)

	_, err = CreateTestTagForImageWithVote(u.ID, i3.ID, tag3value, tx)
	assert.Nil(err)

	_, err = CreateTestTagForImageWithVote(u.ID, i2.ID, tag2value, tx)
	assert.Nil(err)

	_, err = CreateTestTagForImageWithVote(u.ID, i.ID, tag1value, tx)
	assert.Nil(err)

	image, err := SearchImagesBestResult(tag1value, ContentRatingFilterDefault, tx)
	assert.Nil(err)
	assert.NotNil(image)
	assert.Equal(i.ID, image.ID)
}

func TestGetImagesByID(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	_, err = CreateTestImage(u.ID, tx)
	assert.Nil(err)

	_, err = CreateTestImage(u.ID, tx)
	assert.Nil(err)

	_, err = CreateTestTagForImageWithVote(u.ID, i.ID, core.UUIDv4().ToShortString(), tx)
	assert.Nil(err)

	_, err = CreateTestTagForImageWithVote(u.ID, i.ID, core.UUIDv4().ToShortString(), tx)
	assert.Nil(err)

	_, err = CreateTestTagForImageWithVote(u.ID, i.ID, core.UUIDv4().ToShortString(), tx)
	assert.Nil(err)

	baz, err := CreateTestTagForImageWithVote(u.ID, i.ID, core.UUIDv4().ToShortString(), tx)
	assert.Nil(err)

	biz, err := CreateTestTagForImageWithVote(u.ID, i.ID, core.UUIDv4().ToShortString(), tx)
	assert.Nil(err)

	err = SetVoteSummaryVoteCounts(i.ID, baz.ID, 100, 3, tx)
	assert.Nil(err)

	err = SetVoteSummaryVoteCounts(i.ID, biz.ID, 1000, 30, tx)
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

func TestSortImagesByIndex(t *testing.T) {
	assert := assert.New(t)

	ids := []int64{4, 3, 2, 1}

	images := []Image{
		Image{ID: 1},
		Image{ID: 2},
		Image{ID: 3},
		Image{ID: 4},
	}
	assert.Equal(1, images[0].ID)
	sort.Sort(newImagesByIndex(&images, ids))
	assert.Equal(4, images[0].ID)
}

func TestGetAllImagesCensored(t *testing.T) {
	assert := assert.New(t)
	tx, txErr := DB().Begin()
	assert.Nil(txErr)
	defer tx.Rollback()

	u, err := CreateTestUser(tx)
	assert.Nil(err)

	i, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	i2, err := CreateTestImage(u.ID, tx)
	assert.Nil(err)

	i.ContentRating = ContentRatingNR
	err = DB().UpdateInTransaction(i, tx)
	assert.Nil(err)

	censored, err := GetAllImagesWithContentRating(ContentRatingNR, tx)
	assert.NotEmpty(censored)
	assert.None(censored, NewImagePredicate(func(img Image) bool {
		return img.ID == i2.ID
	}))
}

func TestImageWeightedRandom(t *testing.T) {
	assert := assert.New(t)

	images := []imageSignature{
		imageSignature{1, 1.0},
		imageSignature{2, 0.675},
		imageSignature{3, 0.325},
		imageSignature{4, 0.125},
		imageSignature{5, 0.075},
	}

	random1 := imageSignatures(images).WeightedRandom(1)
	assert.Len(random1, 1)
	random2 := imageSignatures(images).WeightedRandom(2)
	assert.Len(random2, 2)
	random3 := imageSignatures(images).WeightedRandom(3)
	assert.Len(random3, 3)
	random4 := imageSignatures(images).WeightedRandom(4)
	assert.Len(random4, 4)
	random5 := imageSignatures(images).WeightedRandom(5)
	assert.Len(random5, 5)
	randomN := imageSignatures(images).WeightedRandom(10)
	assert.Len(randomN, 5)
}
