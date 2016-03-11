package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
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
	ID          int64     `json:"-" db:"id,pk,serial"`
	UUID        string    `json:"uuid" db:"uuid"`
	CreatedUTC  time.Time `json:"created_utc" db:"created_utc"`
	CreatedBy   int64     `json:"created_by" db:"created_by"`
	UpdatedUTC  time.Time `json:"updated_utc,omitempty" db:"updated_utc"`
	UpdatedBy   int64     `json:"updated_by,omitempty" db:"updated_by"`
	DisplayName string    `json:"display_name" db:"display_name"`

	MD5       []byte `json:"md5" db:"md5"`
	S3ReadURL string `json:"s3_read_url" db:"s3_read_url"`
	S3Bucket  string `json:"-" db:"s3_bucket"`
	S3Key     string `json:"-" db:"s3_key"`
	Extension string `json:"extension" db:"extension"`

	Width  int `json:"width" db:"width"`
	Height int `json:"height" db:"height"`

	Tags []Tag `json:"tags" db:"-"`
}

// TableName returns the tablename for the object.
func (i Image) TableName() string {
	return "image"
}

// IsZero returns if the object has been set.
func (i Image) IsZero() bool {
	return i.ID == 0
}

type imageSignature struct {
	ID int64 `json:"-" db:"id"`
}

func (i imageSignature) TableName() string {
	return "image"
}

type imageSignatures []imageSignature

func (is imageSignatures) AsInt64s() []int64 {
	var all []int64
	for x := 0; x < len(is); x++ {
		all = append(all, is[x].ID)
	}
	return all
}

// NewImage returns a new instance of an image.
func NewImage() *Image {
	return &Image{
		UUID:       core.UUIDv4().ToShortString(),
		CreatedUTC: time.Now().UTC(),
	}
}

// GetAllImages returns all the images in the database.
func GetAllImages(tx *sql.Tx) ([]Image, error) {
	var all []Image
	err := spiffy.DefaultDb().GetAllInTransaction(&all, tx)
	return all, err
}

// GetImageByID returns an image for an id.
func GetImageByID(id int64, tx *sql.Tx) (*Image, error) {
	var image Image
	err := spiffy.DefaultDb().GetByIDInTransaction(&image, tx, id)
	return &image, err
}

// GetImageByUUID returns an image by uuid.
func GetImageByUUID(uuid string, tx *sql.Tx) (*Image, error) {
	var image Image
	err := spiffy.DefaultDb().
		QueryInTransaction(`select * from image where uuid = $1`, tx, uuid).Out(&image)
	return &image, err
}

// GetImageByMD5 returns an image by uuid.
func GetImageByMD5(md5sum []byte, tx *sql.Tx) (*Image, error) {
	var image Image
	err := spiffy.DefaultDb().
		QueryInTransaction(`select * from image where md5 = $1`, tx, md5sum).Out(&image)
	return &image, err
}

// GetImagesByID returns images with tags for a list of ids.
func GetImagesByID(ids []int64, tx *sql.Tx) ([]Image, error) {
	idsCSV := fmt.Sprintf("{%s}", csvOfInt(ids))
	query := `select * from image where id = ANY($1::bigint[])`
	var images []Image
	err := spiffy.DefaultDb().QueryInTransaction(query, tx, idsCSV).OutMany(&images)

	if err != nil {
		return nil, err
	}

	imageLookup := map[int64]*Image{}
	for x := 0; x < len(images); x++ {
		i := images[x]
		imageLookup[i.ID] = &i
	}

	var tags []Tag
	query = `
select 
	t.* 
	, itv.image_id
	, itv.votes_for
	, itv.votes_against
	, itv.votes_total
from 
	tag t
	join image_tag_votes itv on itv.tag_id = t.id
where 
	itv.image_id = ANY($1::bigint[])`
	err = spiffy.DefaultDb().QueryInTransaction(query, tx, idsCSV).OutMany(&tags)

	if err != nil {
		return nil, err
	}

	for y := 0; y < len(tags); y++ {
		it := tags[y]
		if i, hasImage := imageLookup[it.ImageID]; hasImage {
			i.Tags = append(i.Tags, it)
		}
	}

	return images, nil
}

// QueryImages searches for an image.
func QueryImages(query string, tx *sql.Tx) ([]Image, error) {
	var imageIDs []imageSignature

	queryFormat := fmt.Sprintf("%%%s%%", query)

	imageQuery := `
select 
	i.id
from 
	image i
	join image_tag_votes itv on i.id = itv.image_id
	join tag t on t.id = itv.tag_id
where
	t.tag_value ilike $1;
`
	err := spiffy.DefaultDb().QueryInTransaction(imageQuery, tx, queryFormat).OutMany(&imageIDs)

	if err != nil {
		return nil, err
	}

	return GetImagesByID(imageSignatures(imageIDs).AsInt64s(), tx)
}

func csvOfInt(input []int64) string {
	outputStrings := []string{}
	for _, v := range input {
		outputStrings = append(outputStrings, fmt.Sprintf("%d", v))
	}
	return strings.Join(outputStrings, ",")
}
