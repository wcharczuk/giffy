package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/blendlabs/spiffy"
)

type Image struct {
	ID          int64     `json:"-" db:"id"`
	UUID        string    `json:"uuid" db:"uuid"`
	CreatedUTC  time.Time `json:"created_utc" db:"created_utc"`
	CreatedBy   int       `json:"created_by" db:"created_by"`
	UpdatedUTC  time.Time `json:"updated_utc,omitempty" db:"updated_utc"`
	UpdatedBy   int       `json:"updated_by,omitempty" db:"updated_by"`
	DisplayName string    `json:"display_name" db:"display_name"`

	MD5       []byte `json:"-" db:"md5"`
	S3ReadURL string `json:"s3_read_url" db:"s3_read_url"`
	S3Bucket  string `json:"-" db:"s3_bucket"`
	S3Key     string `json:"-" db:"s3_key"`
	Extension string `json:"extension" db:"extension"`

	Width  int `json:"width" db:"width"`
	Height int `json:"height" db:"height"`

	Tags []ImageTag `json:"tags" db:"-"`
}

func (i Image) TableName() string {
	return "image"
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

func NewImage() *Image {
	return &Image{
		UUID:       slack.UUIDv4().ToShortString(),
		CreatedUTC: time.Now().UTC(),
	}
}

func GetAllImages(tx *sql.Tx) ([]Image, error) {
	var all []Image
	err := spiffy.DefaultDb().GetAllInTransaction(&all, tx)
	return all, err
}

func GetImageByID(id int64, tx *sql.Tx) (*Image, error) {
	var image Image
	err := spiffy.DefaultDb().GetByIdInTransaction(&image, tx, id)
	return &image, err
}

func GetImageByUUID(uuid string, tx *sql.Tx) (*Image, error) {
	var image Image
	err := spiffy.DefaultDb().
		QueryInTransaction(`select * from image where uuid = $1`, tx, uuid).Out(&image)
	return &image, err
}

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

	var imageTags []ImageTag
	query = `select * from image_tag where image_id = ANY($1::bigint[])`
	err = spiffy.DefaultDb().QueryInTransaction(query, tx, idsCSV).OutMany(&imageTags)

	if err != nil {
		return nil, err
	}

	for y := 0; y < len(imageTags); y++ {
		it := imageTags[y]
		if i, hasImage := imageLookup[it.ImageID]; hasImage {
			i.Tags = append(i.Tags, it)
		}
	}

	return images, nil
}

func QueryImages(query string, tx *sql.Tx) ([]Image, error) {
	var imageIDs []imageSignature

	imageQuery := `
select 
	i.id
from 
	images i
	join image_tag_votes itv on i.id = itv.image_id
	join tag t on  = t.id = itv.tag_id
where
	it.tag_value ilike "%$1%";
`
	err := spiffy.DefaultDb().QueryInTransaction(imageQuery, tx, query).OutMany(imageIDs)

	if err != nil {
		return nil, err
	}

	return GetImagesByID(imageSignatures(imageIDs).AsInt64s(), tx)
}

func csvOfInt(input []int64) string {
	outputStrings := []string{}
	for _, v := range input {
		outputStrings = append(outputStrings, IntToString(v))
	}
	return strings.Join(outputStrings, ",")
}
