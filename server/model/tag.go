package model

import (
	"strings"
	"time"
	"unicode"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/uuid"
)

// NewTag returns a new tag.
func NewTag(createdBy int64, tagValue string) *Tag {
	return &Tag{
		UUID:       uuid.V4().String(),
		CreatedUTC: time.Now().UTC(),
		CreatedBy:  createdBy,
		TagValue:   tagValue,
	}
}

// Tag is a label for an image or set of images.
type Tag struct {
	ID         int64     `json:"-" db:"id,pk,serial"`
	UUID       string    `json:"uuid" db:"uuid"`
	CreatedUTC time.Time `json:"created_utc" db:"created_utc"`
	CreatedBy  int64     `json:"-" db:"created_by"`
	TagValue   string    `json:"tag_value" db:"tag_value"`

	CreatedByUUID string `json:"created_by,omitempty" db:"created_by_uuid,readonly"`
	ImageID       int64  `json:"-" db:"image_uuid,readonly"`
	VotesFor      int    `json:"votes_for,omitempty" db:"votes_for,readonly"`
	VotesAgainst  int    `json:"votes_against,omitempty" db:"votes_against,readonly"`
	VotesTotal    int    `json:"votes_total" db:"votes_total,readonly"`
	VoteRank      int    `json:"vote_rank,omitempty" db:"vote_rank,readonly"`
}

// TableName returns the name of a table.
func (t Tag) TableName() string {
	return "tag"
}

// Populate pulls data off a reader and sets fields on the struct.
func (t *Tag) Populate(r db.Rows) error {
	return r.Scan(&t.ID, &t.UUID, &t.CreatedUTC, &t.CreatedBy, &t.TagValue)
}

// PopulateExtra pulls data off a reader and sets fields on the struct.
func (t *Tag) PopulateExtra(r db.Rows) error {
	return r.Scan(&t.ID, &t.UUID, &t.CreatedUTC, &t.CreatedBy, &t.TagValue, &t.CreatedByUUID, &t.ImageID, &t.VotesFor, &t.VotesAgainst, &t.VotesTotal, &t.VoteRank)
}

// IsZero denotes if an object has been set or not.
func (t *Tag) IsZero() bool {
	return t.ID == 0
}

// CleanTagValue cleans a prospective tag value.
func CleanTagValue(tagValue string) string {
	tagValue = strings.ToLower(tagValue)
	tagValue = strings.Trim(tagValue, " \t\n\r")

	output := ""
	for _, r := range tagValue {
		if unicode.IsLetter(r) || unicode.IsSpace(r) || unicode.IsDigit(r) {
			output = output + string(r)
		}
	}

	return output
}
