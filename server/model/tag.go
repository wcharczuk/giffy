package model

import (
	"database/sql"
	"time"

	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
)

// NewTag returns a new tag.
func NewTag() *Tag {
	return &Tag{
		UUID:       core.UUIDv4().ToShortString(),
		CreatedUTC: time.Now().UTC(),
	}
}

// Tag is a label for an image or set of images.
type Tag struct {
	ID         int64     `json:"-" db:"id,pk,serial"`
	UUID       string    `json:"uuid" db:"uuid"`
	CreatedUTC time.Time `json:"created_utc" db:"created_utc"`
	CreatedBy  int64     `json:"-" db:"created_by"`
	TagValue   string    `json:"tag_value" db:"tag_value"`

	CreatedByUUID string `json:"created_by" db:"created_by_uuid,readonly"`
	ImageID       int64  `json:"-" db:"image_uuid,readonly"`
	VotesFor      int    `json:"votes_for,omitempty" db:"votes_for,readonly"`
	VotesAgainst  int    `json:"votes_against,omitempty" db:"votes_against,readonly"`
	VotesTotal    int    `json:"votes_total,omitempty" db:"votes_total,readonly"`
	VoteRank      int    `json:"vote_rank,omitempty" db:"vote_rank,readonly"`
}

// TableName returns the name of a table.
func (t Tag) TableName() string {
	return "tag"
}

// Populate pulls data off a reader and sets fields on the struct.
func (t *Tag) Populate(r *sql.Rows) error {
	return r.Scan(&t.ID, &t.UUID, &t.CreatedUTC, &t.CreatedBy, &t.TagValue)
}

// PopulateExtra pulls data off a reader and sets fields on the struct.
func (t *Tag) PopulateExtra(r *sql.Rows) error {
	return r.Scan(&t.ID, &t.UUID, &t.CreatedUTC, &t.CreatedBy, &t.TagValue, &t.CreatedByUUID, &t.ImageID, &t.VotesFor, &t.VotesAgainst, &t.VotesTotal, &t.VoteRank)
}

// IsZero denotes if an object has been set or not.
func (t *Tag) IsZero() bool {
	return t.ID == 0
}

// GetAllTags returns all the tags in the db.
func GetAllTags(tx *sql.Tx) ([]Tag, error) {
	var all []Tag
	err := spiffy.DefaultDb().GetAllInTransaction(&all, tx)
	return all, err
}

// GetTagByID returns a tag for a id.
func GetTagByID(id int64, tx *sql.Tx) (*Tag, error) {
	var tag Tag
	err := spiffy.DefaultDb().
		QueryInTransaction(`select * from tag where id = $1`, tx, id).Out(&tag)
	return &tag, err
}

// GetTagByUUID returns a tag for a uuid.
func GetTagByUUID(uuid string, tx *sql.Tx) (*Tag, error) {
	var tag Tag
	err := spiffy.DefaultDb().
		QueryInTransaction(`select * from tag where uuid = $1`, tx, uuid).Out(&tag)
	return &tag, err
}

// GetTagByValue returns a tag for a uuid.
func GetTagByValue(tagValue string, tx *sql.Tx) (*Tag, error) {
	var tag Tag
	err := spiffy.DefaultDb().
		QueryInTransaction(`select * from tag where tag_value ilike $1`, tx, tagValue).Out(&tag)
	return &tag, err
}

// DeleteTagByID deletes an tag fully.
func DeleteTagByID(tagID int64, tx *sql.Tx) error {
	err := spiffy.DefaultDb().ExecInTransaction(`delete from image_tag_votes where tag_id = $1`, tx, tagID)
	if err != nil {
		return err
	}
	err = spiffy.DefaultDb().ExecInTransaction(`delete from vote_log where tag_id = $1`, tx, tagID)
	if err != nil {
		return err
	}
	return spiffy.DefaultDb().ExecInTransaction(`delete from tag where id = $1`, tx, tagID)
}
