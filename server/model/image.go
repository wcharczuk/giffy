package model

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"fmt"
	"image"
	"path/filepath"
	"strings"
	"time"

	// for image processing
	_ "image/gif"
	// for image processing
	_ "image/jpeg"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util/linq"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/core/web"
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

	DisplayName string `json:"display_name" db:"display_name"`

	MD5       []byte `json:"md5" db:"md5"`
	S3ReadURL string `json:"s3_read_url" db:"s3_read_url"`
	S3Bucket  string `json:"-" db:"s3_bucket"`
	S3Key     string `json:"-" db:"s3_key"`

	Width  int `json:"width" db:"width"`
	Height int `json:"height" db:"height"`

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

//Populate popultes the object from rows.
func (i *Image) Populate(r *sql.Rows) error {
	return r.Scan(
		&i.ID,
		&i.UUID,
		&i.CreatedUTC,
		&i.CreatedBy,
		&i.DisplayName,
		&i.MD5,
		&i.S3ReadURL,
		&i.S3Bucket,
		&i.S3Key,
		&i.Width,
		&i.Height,
		&i.Extension,
	)
}

// ImagePredicate is used in linq queries
type ImagePredicate func(i Image) bool

// NewImagePredicate creates a new ImagePredicate that resolves to a linq.Predicate
func NewImagePredicate(predicate ImagePredicate) linq.Predicate {
	return func(item interface{}) bool {
		if typed, isTyped := item.(Image); isTyped {
			return predicate(typed)
		}
		return false
	}
}

// NewImage returns a new instance of an image.
func NewImage() *Image {
	return &Image{
		UUID:       core.UUIDv4().ToShortString(),
		CreatedUTC: time.Now().UTC(),
	}
}

// NewImageFromPostedFile creates an image and parses the meta data for an image from a posted file.
func NewImageFromPostedFile(userID int64, postedFile web.PostedFile) (*Image, error) {
	md5sum := ConvertMD5(md5.Sum(postedFile.Contents))
	existing, err := GetImageByMD5(md5sum, nil)
	if err != nil {
		return nil, err
	}

	if !existing.IsZero() {
		return existing, nil
	}

	newImage := NewImage()
	newImage.MD5 = md5sum
	newImage.CreatedBy = userID

	imageBuf := bytes.NewBuffer(postedFile.Contents)

	// read the image metadata
	// this relies on the `image/*` imports.
	imageMeta, _, err := image.DecodeConfig(imageBuf)
	if err != nil {
		return nil, exception.Wrap(err)
	}

	newImage.DisplayName = postedFile.Filename
	newImage.Extension = filepath.Ext(postedFile.Filename)
	newImage.Height = imageMeta.Height
	newImage.Width = imageMeta.Width
	return newImage, nil
}

// GetAllImages returns all the images in the database.
func GetAllImages(tx *sql.Tx) ([]Image, error) {
	return GetImagesByID(nil, tx)
}

// GetRandomImages returns an image by uuid.
func GetRandomImages(count int, tx *sql.Tx) ([]Image, error) {
	var imageIDs []imageSignature
	err := spiffy.DefaultDb().
		QueryInTransaction(`select id from (select id, row_number() over (order by gen_random_uuid()) as rank from image) data where rank <= $1`, tx, count).OutMany(&imageIDs)

	if err != nil {
		return nil, err
	}

	images, err := GetImagesByID(imageSignatures(imageIDs).AsInt64s(), tx)
	if err != nil {
		return nil, err
	}
	return images, err
}

// GetImageByID returns an image for an id.
func GetImageByID(id int64, tx *sql.Tx) (*Image, error) {
	images, err := GetImagesByID([]int64{id}, tx)
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return &Image{}, nil
	}
	return &images[0], err
}

// GetImageByUUID returns an image by uuid.
func GetImageByUUID(uuid string, tx *sql.Tx) (*Image, error) {
	var image imageSignature
	err := spiffy.DefaultDb().
		QueryInTransaction(`select id from image where uuid = $1`, tx, uuid).Out(&image)

	images, err := GetImagesByID([]int64{image.ID}, tx)
	if err != nil {
		return nil, err
	}
	return &images[0], err
}

// GetImageByMD5 returns an image by uuid.
func GetImageByMD5(md5sum []byte, tx *sql.Tx) (*Image, error) {
	image := Image{}
	err := spiffy.DefaultDb().
		QueryInTransaction(`select * from image where md5 = $1`, tx, md5sum).Out(&image)
	return &image, err
}

// UpdateImageDisplayName sets just the display name for an image.
func UpdateImageDisplayName(imageID int64, displayName string, tx *sql.Tx) error {
	return spiffy.DefaultDb().ExecInTransaction("update image set display_name = $2 where id = $1", tx, imageID, displayName)
}

// DeleteImageByID deletes an image fully.
func DeleteImageByID(imageID int64, tx *sql.Tx) error {
	err := spiffy.DefaultDb().ExecInTransaction(`delete from vote_summary where image_id = $1`, tx, imageID)
	if err != nil {
		return err
	}
	err = spiffy.DefaultDb().ExecInTransaction(`delete from vote where image_id = $1`, tx, imageID)
	if err != nil {
		return err
	}
	return spiffy.DefaultDb().ExecInTransaction(`delete from image where id = $1`, tx, imageID)
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
	join vote_summary vs on i.id = vs.image_id
	join tag t on t.id = vs.tag_id
where
	t.tag_value ilike $1;
`
	err := spiffy.DefaultDb().QueryInTransaction(imageQuery, tx, queryFormat).OutMany(&imageIDs)

	if err != nil {
		return nil, err
	}

	if len(imageIDs) == 0 {
		return []Image{}, nil
	}

	ids := imageSignatures(imageIDs).AsInt64s()
	return GetImagesByID(ids, tx)
}

// GetImagesByID returns images with tags for a list of ids.
func GetImagesByID(ids []int64, tx *sql.Tx) ([]Image, error) {
	var err error
	var populateErr error

	imageQueryAll := `select * from image`
	imageQuerySingle := fmt.Sprintf(`%s where id = $1`, imageQueryAll)
	imageQueryMany := fmt.Sprintf(`%s where id = ANY($1::bigint[])`, imageQueryAll)

	tagQueryAll := `
select 
	t.*
	, u.uuid as created_by_uuid
	, vs.image_id
	, vs.votes_for
	, vs.votes_against
	, vs.votes_total
	, row_number() over (partition by image_id order by vs.votes_total desc) as vote_rank 
from 
			tag t 
	join 	vote_summary 	vs 	on vs.tag_id = t.id 
	join 	users 			u 	on u.id = t.created_by
`
	tagQuerySingle := fmt.Sprintf(`%s where vs.image_id = $1`, tagQueryAll)
	tagQueryMany := fmt.Sprintf(`%s where vs.image_id = ANY($1::bigint[])`, tagQueryAll)

	tagQueryOuter := `
select * from (
%s
) as intermediate
where vote_rank <= 5
`

	userQueryAll := `select u.* from image i join users u on i.created_by = u.id`
	userQuerySingle := fmt.Sprintf(`%s where i.id = $1`, userQueryAll)
	userQueryMany := fmt.Sprintf(`%s where i.id = ANY($1::bigint[])`, userQueryAll)

	images := []*Image{}
	imageLookup := map[int64]*Image{}
	userLookup := map[int64]*User{}

	imageConsumer := func(r *sql.Rows) error {
		i := &Image{}
		populateErr = i.Populate(r)
		if populateErr != nil {
			return populateErr
		}
		images = append(images, i)
		imageLookup[i.ID] = i
		return nil
	}

	tagConsumer := func(r *sql.Rows) error {
		t := &Tag{}
		populateErr = t.PopulateExtra(r)
		if populateErr != nil {
			return populateErr
		}

		i := imageLookup[t.ImageID]
		if i != nil {
			i.Tags = append(i.Tags, *t)
		}

		return nil
	}

	userConsumer := func(r *sql.Rows) error {
		u := &User{}
		populateErr = u.Populate(r)
		if populateErr != nil {
			return populateErr
		}
		userLookup[u.ID] = u
		return nil
	}

	if len(ids) > 1 {
		idsCSV := fmt.Sprintf("{%s}", csvOfInt(ids))
		err = spiffy.DefaultDb().QueryInTransaction(imageQueryMany, tx, idsCSV).Each(imageConsumer)
		if err != nil {
			return nil, err
		}
		err = spiffy.DefaultDb().QueryInTransaction(fmt.Sprintf(tagQueryOuter, tagQueryMany), tx, idsCSV).Each(tagConsumer)
		if err != nil {
			return nil, err
		}
		err = spiffy.DefaultDb().QueryInTransaction(userQueryMany, tx, idsCSV).Each(userConsumer)
		if err != nil {
			return nil, err
		}
	} else if len(ids) == 1 {
		err = spiffy.DefaultDb().QueryInTransaction(imageQuerySingle, tx, ids[0]).Each(imageConsumer)
		if err != nil {
			return nil, err
		}
		err = spiffy.DefaultDb().QueryInTransaction(fmt.Sprintf(tagQueryOuter, tagQuerySingle), tx, ids[0]).Each(tagConsumer)
		if err != nil {
			return nil, err
		}
		err = spiffy.DefaultDb().QueryInTransaction(userQuerySingle, tx, ids[0]).Each(userConsumer)
		if err != nil {
			return nil, err
		}
	} else {
		err = spiffy.DefaultDb().QueryInTransaction(imageQueryAll, tx).Each(imageConsumer)
		if err != nil {
			return nil, err
		}
		err = spiffy.DefaultDb().QueryInTransaction(fmt.Sprintf(tagQueryOuter, tagQueryAll), tx).Each(tagConsumer)
		if err != nil {
			return nil, err
		}
		err = spiffy.DefaultDb().QueryInTransaction(userQueryAll, tx).Each(userConsumer)
		if err != nil {
			return nil, err
		}
	}

	finalImages := make([]Image, len(images))
	for x := 0; x < len(images); x++ {
		img := images[x]
		if u, ok := userLookup[img.CreatedBy]; ok {
			img.CreatedByUser = u
		}

		finalImages[x] = *img
	}

	return finalImages, nil
}

// --------------------------------------------------------------------------------
// Helper Functions / Types
// --------------------------------------------------------------------------------

func csvOfInt(input []int64) string {
	outputStrings := []string{}
	for _, v := range input {
		outputStrings = append(outputStrings, fmt.Sprintf("%d", v))
	}
	return strings.Join(outputStrings, ",")
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
