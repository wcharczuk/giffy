package model

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/util"
	"github.com/blend/go-sdk/uuid"
)

func TestGetAllImages(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	i, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, uuid.V4().String())
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, uuid.V4().String())
	assert.Nil(err)

	allImages, err := m.GetAllImages(todo)
	assert.Nil(err)
	assert.NotEmpty(allImages)

	var firstImage *Image
	for _, image := range allImages {
		if image.ID == i.ID {
			firstImage = &image
			break
		}
	}

	assert.NotNil(firstImage)
	assert.False(firstImage.IsZero())
	assert.NotNil(firstImage.CreatedByUser)
	assert.NotEmpty(firstImage.Tags)
}

func TestGetRandomImages(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	for x := 0; x < 10; x++ {
		_, err = m.CreateTestImage(todo, u.ID)
		assert.Nil(err)
	}

	images, err := m.GetRandomImages(todo, 5)
	assert.Nil(err)
	assert.Len(images, 5)
}

func TestGetImageByID(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	i, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, uuid.V4().String())
	assert.Nil(err)

	verify, err := m.GetImageByID(todo, i.ID)
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
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	verify, err := m.GetImageByID(todo, -1)
	assert.Nil(err)
	assert.NotNil(verify)
	assert.True(verify.IsZero())
}

func TestUpdateImageDisplayName(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	i, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	err = m.UpdateImageDisplayName(todo, i.ID, fmt.Sprintf("not %s", i.DisplayName))
	assert.Nil(err)

	verify, err := m.GetImageByID(todo, i.ID)
	assert.Nil(err)
	assert.Equal(fmt.Sprintf("not %s", i.DisplayName), verify.DisplayName)
}

func TestDeleteImageByID(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	i, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, uuid.V4().String())
	assert.Nil(err)

	err = m.DeleteImageByID(todo, i.ID)
	assert.Nil(err)

	verify, err := m.GetImageByID(todo, i.ID)
	assert.Nil(err)
	assert.True(verify.IsZero())
}

func TestImageMD5Check(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	i, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	verify, err := m.GetImageByMD5(todo, i.MD5, tx)
	assert.Nil(err)
	assert.False(verify.IsZero())
}

func TestSearchImages(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	i4, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	i3, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	i2, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	i, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i4.ID, "not_foo_bar")
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i3.ID, "__test_foo_bar")
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i2.ID, "__test_bar")
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, "__test")
	assert.Nil(err)

	images, err := m.SearchImages(todo, "__test", ContentRatingFilterDefault)
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
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	i4, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	i3, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	i2, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	i, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i4.ID, "not__test_foo_bar")
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i3.ID, "__test_foo_bar")
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i2.ID, "__test_bar")
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, "__test")
	assert.Nil(err)

	images, err := m.SearchImagesWeightedRandom(todo, "__test", ContentRatingFilterDefault, 2)
	assert.Nil(err)
	assert.NotNil(images)
	assert.NotEmpty(images)
}

func TestSearchImagesBestResult(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	i4, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	i3, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	i2, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	i, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	tag4value := fmt.Sprintf("__test_foo_bar_%s", util.String.Random(4))
	tag3value := fmt.Sprintf("__test_bar_%s", util.String.Random(4))
	tag2value := fmt.Sprintf("__test_foo_%s", util.String.Random(4))
	tag1value := fmt.Sprintf("__test_%s", util.String.Random(4))

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i4.ID, tag4value)
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i3.ID, tag3value)
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i2.ID, tag2value)
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, tag1value)
	assert.Nil(err)

	image, err := m.SearchImagesBestResult(todo, tag1value, nil, ContentRatingFilterDefault)
	assert.Nil(err)
	assert.NotNil(image)
	assert.Equal(i.ID, image.ID)

	image, err = m.SearchImagesBestResult(todo, "__test", []string{i.UUID}, ContentRatingFilterDefault)
	assert.Nil(err)
	assert.NotNil(image)
	assert.NotEqual(i.ID, image.ID)
}

func TestImageSignaturesWeightedRandom(t *testing.T) {
	assert := assert.New(t)

	r := rand.New(rand.NewSource(1))

	sigs := imageSignatures([]imageSignature{
		{1, 3.0},
		{2, 3.0},
		{3, 3.0},
		{4, 2.0},
		{5, 1.0},
	})
	best := sigs.WeightedRandom(1, r)
	assert.Len(best, 1)
	assert.Equal(2, best[0].ID, best.String())

	best = sigs.WeightedRandom(1, r)
	assert.Len(best, 1)
	assert.Equal(2, best[0].ID, best.String())

	best = sigs.WeightedRandom(1, r)
	assert.Len(best, 1)
	assert.Equal(1, best[0].ID, best.String())
}

func TestImageSignaturesSortRandom(t *testing.T) {
	assert := assert.New(t)

	r := rand.New(rand.NewSource(1))

	sigs := imageSignatures([]imageSignature{
		{1, 1.0},
		{2, 2.0},
		{3, 3.0},
		{4, 4.0},
		{5, 5.0},
		{6, 6.0},
		{7, 7.0},
		{8, 8.0},
	})

	sort.Sort(imageSignaturesRandom(sigs, r))
	assert.Equal(8, sigs[0].ID)
	assert.Equal(7, sigs[1].ID)
	assert.Equal(3, sigs[2].ID)
	assert.Equal(5, sigs[3].ID)
	assert.Equal(4, sigs[4].ID)
}

func TestImageSignaturesSortScoreDescending(t *testing.T) {
	assert := assert.New(t)

	sigs := imageSignatures([]imageSignature{
		{1, 1.0},
		{2, 1.5},
		{3, 2.0},
		{4, 3.0},
		{5, 5.0},
	})

	sort.Sort(imageSignaturesScoreDescending(sigs))
	assert.Equal(5, sigs[0].ID)
	assert.Equal(4, sigs[1].ID)
	assert.Equal(3, sigs[2].ID)
	assert.Equal(2, sigs[3].ID)
	assert.Equal(1, sigs[4].ID)
}

func TestGetImagesByID(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	i, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	_, err = m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	_, err = m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, uuid.V4().String())
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, uuid.V4().String())
	assert.Nil(err)

	_, err = m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, uuid.V4().String())
	assert.Nil(err)

	baz, err := m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, uuid.V4().String())
	assert.Nil(err)

	biz, err := m.CreateTestTagForImageWithVote(todo, u.ID, i.ID, uuid.V4().String())
	assert.Nil(err)

	err = m.SetVoteSummaryVoteCounts(todo, i.ID, baz.ID, 100, 3)
	assert.Nil(err)

	err = m.SetVoteSummaryVoteCounts(todo, i.ID, biz.ID, 1000, 30)
	assert.Nil(err)

	images, err := m.GetAllImages(todo)
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
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	u, err := m.CreateTestUser(todo)
	assert.Nil(err)

	i, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	i2, err := m.CreateTestImage(todo, u.ID)
	assert.Nil(err)

	i.ContentRating = ContentRatingNR
	err = m.Invoke(todo).Update(i)
	assert.Nil(err)

	censored, err := m.GetAllImagesWithContentRating(todo, ContentRatingNR)
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
