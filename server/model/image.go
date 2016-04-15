package model

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"fmt"
	"image"
	"path/filepath"
	"sort"
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

	DisplayName string `json:"display_name" db:"display_name"`
	IsCensored  bool   `json:"is_censored" db:"is_censored"`

	MD5       []byte `json:"md5" db:"md5"`
	S3ReadURL string `json:"s3_read_url" db:"s3_read_url"`
	S3Bucket  string `json:"-" db:"s3_bucket"`
	S3Key     string `json:"-" db:"s3_key"`

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

//Populate popultes the object from rows.
func (i *Image) Populate(r *sql.Rows) error {
	return r.Scan(
		&i.ID,
		&i.UUID,
		&i.CreatedUTC,
		&i.CreatedBy,
		&i.DisplayName,
		&i.IsCensored,
		&i.MD5,
		&i.S3ReadURL,
		&i.S3Bucket,
		&i.S3Key,
		&i.Width,
		&i.Height,
		&i.FileSize,
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
func NewImageFromPostedFile(userID int64, shouldValidate bool, fileContents []byte, fileName string) (*Image, error) {
	newImage := NewImage()
	newImage.MD5 = ConvertMD5(md5.Sum(fileContents))
	newImage.CreatedBy = userID

	imageBuf := bytes.NewBuffer(fileContents)

	// read the image metadata
	// this relies on the `image/*` imports.
	imageMeta, _, err := image.DecodeConfig(imageBuf)
	if err != nil {
		return nil, exception.Wrap(err)
	}

	newImage.DisplayName = fileName
	newImage.Extension = strings.ToLower(filepath.Ext(fileName))
	newImage.Height = imageMeta.Height
	newImage.Width = imageMeta.Width
	newImage.FileSize = len(fileContents)

	if shouldValidate {
		if newImage.Width < MinImageWidth {
			return nil, exception.Newf("Image width needs to be > %dpx.", MinImageWidth)
		}

		if newImage.Height < MinImageHeight {
			return nil, exception.Newf("Image height needs to be > %dpx.", MinImageHeight)
		}

		if newImage.Width < MinImageHeightOrWidth && newImage.Height < MinImageHeightOrWidth {
			return nil, exception.Newf("Image width or height need to be > %dpx.", MinImageHeightOrWidth)
		}

		if newImage.FileSize > MaxImageSize {
			return nil, exception.New("Image file size should be < 32 mb.")
		}
	}

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
		QueryInTransaction(`select id from (select id, row_number() over (order by gen_random_uuid()) as rank from image where is_censored = false) data where rank <= $1`, tx, count).OutMany(&imageIDs)

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
	if len(images) == 0 {
		return &Image{}, nil
	}

	return &images[0], err
}

// GetImageByMD5 returns an image by uuid.
func GetImageByMD5(md5sum []byte, tx *sql.Tx) (*Image, error) {
	image := Image{}
	imageColumns := spiffy.CachedColumnCollectionFromInstance(Image{}).ColumnNames()
	err := spiffy.DefaultDb().
		QueryInTransaction(fmt.Sprintf(`select %s from image where md5 = $1`, strings.Join(imageColumns, ",")), tx, md5sum).Out(&image)
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

// SearchImages searches for an image.
func SearchImages(query string, tx *sql.Tx) ([]Image, error) {
	var imageIDs []imageSignature

	imageQuery := `
select
	id
from
(
	select
		vs.image_id as id
		, similarity(t.tag_value, $1) as relevance
		, vs.votes_total as votes_total
	from
		tag t
		join vote_summary vs on t.id = vs.tag_id
		join image i on vs.image_id = i.id
	where
		vs.votes_total > 0
		and t.tag_value % $1
		and i.is_censored = false
) as results
order by
	relevance desc,
	votes_total desc
`
	err := spiffy.DefaultDb().QueryInTransaction(imageQuery, tx, query).OutMany(&imageIDs)

	if err != nil {
		return nil, err
	}

	if len(imageIDs) == 0 {
		return []Image{}, nil
	}

	ids := imageSignatures(imageIDs).AsInt64s()
	return GetImagesByID(ids, tx)
}

// SearchImagesSlack is the query we use for slack.
func SearchImagesSlack(query string, tx *sql.Tx) (*Image, error) {
	imageQuery := `
select
	id
from
	(
		select
			vs.image_id as id
		from
			(
				select
					t.id as tag_id
				from
					tag t
				where
					t.tag_value % $1
				order by
					similarity(t.tag_value, $1) desc
				limit 1
			) best_tag
			join vote_summary vs on vs.tag_id = best_tag.tag_id
			join image i on vs.image_id = i.id
		where
			vs.votes_total > 0
			and i.is_censored = false
		order by
			gen_random_uuid()
		limit 1
	) results
limit 1
`
	imageID := imageSignature{}
	err := spiffy.DefaultDb().QueryInTransaction(imageQuery, tx, query).Out(&imageID)

	if err != nil {
		return nil, err
	}

	if imageID.ID == 0 {
		return &Image{}, nil
	}

	images, err := GetImagesByID([]int64{imageID.ID}, tx)
	if err != nil {
		return &Image{}, err
	}

	return &images[0], nil
}

// GetImagesForUserID returns images for a user.
func GetImagesForUserID(userID int64, tx *sql.Tx) ([]Image, error) {
	var imageIDs []imageSignature
	imageQuery := `select i.id from image i where created_by = $1`
	err := spiffy.DefaultDb().QueryInTransaction(imageQuery, tx, userID).OutMany(&imageIDs)
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

	imageColumns := spiffy.CachedColumnCollectionFromInstance(Image{}).ColumnNames()

	imageQueryAll := fmt.Sprintf(`select %s from image`, strings.Join(imageColumns, ","))
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

	intermediateImages := []*Image{}
	imageLookup := map[int64]*Image{}
	userLookup := map[int64]*User{}

	imageConsumer := func(r *sql.Rows) error {
		i := &Image{}
		populateErr = i.Populate(r)
		if populateErr != nil {
			return populateErr
		}
		intermediateImages = append(intermediateImages, i)
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

	finalImages := make([]Image, len(intermediateImages))
	for x := 0; x < len(intermediateImages); x++ {
		img := intermediateImages[x]
		if u, ok := userLookup[img.CreatedBy]; ok {
			img.CreatedByUser = u
		}

		finalImages[x] = *img
	}

	if len(ids) > 1 {
		sort.Sort(newImagesByIndex(&finalImages, ids))
	}

	return finalImages, nil
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
