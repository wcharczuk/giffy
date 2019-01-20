package model

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"image"
	"math/rand"
	"path/filepath"
	"sort"
	"strings"
	"time"

	// for image processing
	_ "image/gif"
	// for image processing
	_ "image/jpeg"
	// for image processing
	_ "image/png"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/uuid"
	"github.com/wcharczuk/giffy/server/core"
)

const (
	// MaxImageSize is 32 megabytes.
	MaxImageSize = 1 << 25 // 32 mb

	// MinImageHeight is the min height
	MinImageHeight = 200

	// MinImageWidth is the min image width
	MinImageWidth = 200

	// MinImageHeightOrWidth is the minimum height or width.
	MinImageHeightOrWidth = 300
)

// ConvertMD5 takes a fixed buffer and turns it into a byte slice.
func ConvertMD5(md5sum [16]byte) []byte {
	typedBuffer := make([]byte, 16)
	for i, b := range md5sum {
		typedBuffer[i] = b
	}
	return typedBuffer
}

// Image is an image stored in the db.
type Image struct {
	ID            int64     `json:"-" db:"id,pk,serial"`
	UUID          string    `json:"uuid" db:"uuid"`
	CreatedUTC    time.Time `json:"created_utc" db:"created_utc"`
	CreatedBy     int64     `json:"-" db:"created_by"`
	CreatedByUser *User     `json:"created_by,omitempty" db:"-"`

	DisplayName   string `json:"display_name" db:"display_name"`
	ContentRating int    `json:"content_rating" db:"content_rating"`

	MD5      []byte `json:"md5" db:"md5"`
	S3Bucket string `json:"s3_bucket" db:"s3_bucket"`
	S3Key    string `json:"s3_key" db:"s3_key"`

	Width  int `json:"width" db:"width"`
	Height int `json:"height" db:"height"`

	FileSize  int    `json:"file_size" db:"file_size"`
	Extension string `json:"extension" db:"extension"`

	Tags []Tag `json:"tags,omitempty" db:"-"`
}

// TableName returns the tablename for the object.
func (i Image) TableName() string {
	return "image"
}

// IsZero returns if the object has been set.
func (i Image) IsZero() bool {
	return i.ID == 0
}

// GetTagsSummary returns a csv of the tags for the image.
func (i Image) GetTagsSummary() string {
	if len(i.Tags) == 0 {
		return "N/A"
	}

	var values []string
	for _, t := range i.Tags {
		values = append(values, fmt.Sprintf("(%d) %s", t.VotesTotal, t.TagValue))
	}
	return strings.Join(values, ", ")
}

//Populate popultes the object from rows.
func (i *Image) Populate(r db.Rows) error {
	return exception.New(r.Scan(
		&i.ID,
		&i.UUID,
		&i.CreatedUTC,
		&i.CreatedBy,
		&i.DisplayName,
		&i.ContentRating,
		&i.MD5,
		&i.S3Bucket,
		&i.S3Key,
		&i.Width,
		&i.Height,
		&i.FileSize,
		&i.Extension,
	))
}

// NewImage returns a new instance of an image.
func NewImage() *Image {
	return &Image{
		UUID:          uuid.V4().String(),
		CreatedUTC:    time.Now().UTC(),
		ContentRating: ContentRatingPG13,
	}
}

// NewImageFromPostedFile creates an image and parses the meta data for an image from a posted file.
func NewImageFromPostedFile(userID int64, shouldValidate bool, fileContents []byte, fileName string) (*Image, error) {
	newImage := NewImage()
	newImage.MD5 = ConvertMD5(md5.Sum(fileContents))
	newImage.CreatedBy = userID

	imageBuf := bytes.NewBuffer(fileContents)

	// read the image metadata
	// this relies on the `image/*` imports.
	imageMeta, _, err := image.DecodeConfig(imageBuf)
	if err != nil {
		return nil, exception.New(err)
	}

	newImage.ContentRating = ContentRatingG
	newImage.DisplayName = fileName
	newImage.Extension = strings.ToLower(filepath.Ext(fileName))
	newImage.Height = imageMeta.Height
	newImage.Width = imageMeta.Width
	newImage.FileSize = len(fileContents)

	if shouldValidate {
		if newImage.Width < MinImageWidth {
			return nil, exception.New("invalid image").WithMessagef("Image width needs to be > %dpx.", MinImageWidth)
		}

		if newImage.Height < MinImageHeight {
			return nil, exception.New("invalid image").WithMessagef("Image height needs to be > %dpx.", MinImageHeight)
		}

		if newImage.Width < MinImageHeightOrWidth && newImage.Height < MinImageHeightOrWidth {
			return nil, exception.New("invalid image").WithMessagef("Image width or height need to be > %dpx.", MinImageHeightOrWidth)
		}

		if newImage.FileSize > MaxImageSize {
			return nil, exception.New("invalid image").WithMessagef("Image file size should be < 32 mb.")
		}
	}

	return newImage, nil
}

// --------------------------------------------------------------------------------
// Helper Functions / Types
// --------------------------------------------------------------------------------

func newImagesByIndex(images *[]Image, ids []int64) *imagesByIndex {
	is := imagesByIndex{
		images:  images,
		indexes: map[int64]int{},
	}
	for i, id := range ids {
		is.indexes[id] = i
	}
	return &is
}

type imagesByIndex struct {
	images  *[]Image
	indexes map[int64]int
}

func (is imagesByIndex) Len() int {
	return len(*is.images)
}

func (is imagesByIndex) Less(i, j int) bool {
	firstID := (*is.images)[i].ID
	secondID := (*is.images)[j].ID
	return is.indexes[firstID] < is.indexes[secondID]
}

func (is imagesByIndex) Swap(i, j int) {
	(*is.images)[i], (*is.images)[j] = (*is.images)[j], (*is.images)[i]
}

func csvOfInt(input []int64) string {
	outputStrings := []string{}
	for _, v := range input {
		outputStrings = append(outputStrings, fmt.Sprintf("%d", v))
	}
	return strings.Join(outputStrings, ",")
}

type imageSignature struct {
	ID    int64   `db:"id,readonly"`
	Score float64 `db:"score,readonly"`
}

func (i imageSignature) TableName() string {
	return "image_signature" //note this doesn't matter, its just for column metadata reasons.
}

type imageSignatures []imageSignature

func (is imageSignatures) AsInt64s() []int64 {
	var all []int64
	for x := 0; x < len(is); x++ {
		all = append(all, is[x].ID)
	}
	return all
}

func (is imageSignatures) TotalScore() float64 {
	var totalScore float64
	for x := 0; x < len(is); x++ {
		totalScore += is[x].Score
	}
	return totalScore
}

func (is imageSignatures) NormalizeScores(totalScore float64) imageSignatures {
	var normalizedScores []imageSignature
	for x := 0; x < len(is); x++ {
		i := is[x]
		normalizedScores = append(normalizedScores, imageSignature{
			ID:    i.ID,
			Score: i.Score / totalScore,
		})
	}
	return imageSignatures(normalizedScores)
}

func (is imageSignatures) String() string {
	var values []string
	for x := 0; x < len(is); x++ {
		v := is[x]
		values = append(values, fmt.Sprintf("%d:%0.5f", v.ID, v.Score))
	}
	return strings.Join(values, ", ")
}

func (is imageSignatures) WeightedRandom(count int, r ...*rand.Rand) imageSignatures {
	if count >= len(is) {
		return is
	}

	total := is.TotalScore()

	var randSource *rand.Rand
	if len(r) > 0 {
		randSource = r[0]
	} else {
		randSource = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	// randomly sort
	sort.Sort(imageSignaturesRandom(is, randSource))
	sort.Sort(imageSignaturesRandom(is, randSource))

	if count >= len(is) {
		return is
	}

	selections := core.SetOfInt64{}
	var selectedImages []imageSignature
	for len(selectedImages) < count {

		randomValue := randSource.Float64() * total

		for x := 0; x < len(is); x++ {
			i := is[x]

			if i.Score > randomValue {
				if !selections.Contains(i.ID) {
					selections.Add(i.ID)
					selectedImages = append(selectedImages, i)
					break
				}
			}
		}
	}

	return imageSignatures(selectedImages)
}

// sort random
func imageSignaturesRandom(values []imageSignature, r ...*rand.Rand) *imageSignaturesRandomSorter {
	if len(r) > 0 {
		return &imageSignaturesRandomSorter{
			values: values,
			r:      r[0],
		}
	}
	return &imageSignaturesRandomSorter{
		values: values,
		r:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

type imageSignaturesRandomSorter struct {
	values []imageSignature
	r      *rand.Rand
}

func (a imageSignaturesRandomSorter) Len() int {
	return len(a.values)
}

func (a imageSignaturesRandomSorter) Swap(i, j int) {
	a.values[i], a.values[j] = a.values[j], a.values[i]
}

func (a imageSignaturesRandomSorter) Less(i, j int) bool {
	rv := a.r.Float64()
	return rv > 0.5
}

// sort ascending

type imageSignaturesScoreAscending []imageSignature

func (issa imageSignaturesScoreAscending) Len() int {
	return len(issa)
}

func (issa imageSignaturesScoreAscending) Swap(i, j int) {
	issa[i], issa[j] = issa[j], issa[i]
}

func (issa imageSignaturesScoreAscending) Less(i, j int) bool {
	return issa[i].Score < issa[j].Score
}

// sort descending

type imageSignaturesScoreDescending []imageSignature

func (isda imageSignaturesScoreDescending) Len() int {
	return len(isda)
}

func (isda imageSignaturesScoreDescending) Swap(i, j int) {
	isda[i], isda[j] = isda[j], isda[i]
}

func (isda imageSignaturesScoreDescending) Less(i, j int) bool {
	return isda[i].Score > isda[j].Score
}
