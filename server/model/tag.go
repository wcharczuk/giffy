package model

import (
	"database/sql"
	"time"

	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
)

type Tag struct {
	ID         int64     `json:"-" db:"id"`
	UUID       string    `json:"uuid" db:"uuid"`
	CreatedUTC time.Time `json:"created_utc" db:"created_utc"`
	CreatedBy  int64     `json:"created_by" db:"created_by"`
	TagValue   string    `json:"tag_value" db:"tag_value"`

	ImageID      int64 `json:"image_id,omitempty" db:"image_id,readonly"`
	VotesFor     int   `json:"votes_for,omitempty" db:"votes_for,readonly"`
	VotesAgainst int   `json:"votes_against,omitempty" db:"votes_against,readonly"`
	VotesTotal   int   `json:"votes_total,omitempty" db:"votes_total,readonly"`
}

func (it Tag) TableName() string {
	return "tag"
}

func NewTag() *Tag {
	return &Tag{
		UUID:       core.UUIDv4().ToShortString(),
		CreatedUTC: time.Now().UTC(),
	}
}

func GetAllTags(tx *sql.Tx) ([]Tag, error) {
	var all []Tag
	err := spiffy.DefaultDb().GetAllInTransaction(&all, tx)
	return all, err
}

func GetTagByUUID(uuid string, tx *sql.Tx) (*Tag, error) {
	var imageTag Tag
	err := spiffy.DefaultDb().
		QueryInTransaction(`select * from tag where uuid = $1`, tx, uuid).Out(&imageTag)
	return &imageTag, err
}

func GetTagsForImageID(imageID int64, tx *sql.Tx) ([]Tag, error) {
	var tags []Tag
	query := `
select 
	t.*
	, itv.image_id
	, itv.votes_for
	, itv.votes_against
	, itv.votes_total
from 
	tag t 
	join image_tag_votes itv on t.id = itv.tag_id 
where 
	itv.image_id = $1
order by
	itv.votes_total desc;
`
	err := spiffy.DefaultDb().QueryInTransaction(query, tx, imageID).OutMany(&tags)
	return tags, err
}
