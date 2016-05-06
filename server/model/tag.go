package model

import (
	"database/sql"
	"strings"
	"time"
	"unicode"

	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
)

// NewTag returns a new tag.
func NewTag(createdBy int64, tagValue string) *Tag {
	return &Tag{
		UUID:       core.UUIDv4().ToShortString(),
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
	all := []Tag{}
	err := spiffy.DefaultDb().GetAllInTransaction(&all, tx)
	return all, err
}

// GetRandomTags gets a random selection of tags.
func GetRandomTags(count int, tx *sql.Tx) ([]Tag, error) {
	tags := []Tag{}
	err := spiffy.DefaultDb().QueryInTransaction(`select * from tag order by gen_random_uuid() limit $1;`, tx, count).OutMany(&tags)
	return tags, err
}

// GetTagByID returns a tag for a id.
func GetTagByID(id int64, tx *sql.Tx) (*Tag, error) {
	tag := Tag{}
	err := spiffy.DefaultDb().
		QueryInTransaction(`select * from tag where id = $1`, tx, id).Out(&tag)
	return &tag, err
}

// GetTagByUUID returns a tag for a uuid.
func GetTagByUUID(uuid string, tx *sql.Tx) (*Tag, error) {
	tag := Tag{}
	err := spiffy.DefaultDb().
		QueryInTransaction(`select * from tag where uuid = $1`, tx, uuid).Out(&tag)
	return &tag, err
}

// GetTagByValue returns a tag for a uuid.
func GetTagByValue(tagValue string, tx *sql.Tx) (*Tag, error) {
	tag := Tag{}
	err := spiffy.DefaultDb().
		QueryInTransaction(`select * from tag where tag_value ilike $1`, tx, tagValue).Out(&tag)
	return &tag, err
}

// SearchTags searches tags.
func SearchTags(query string, tx *sql.Tx) ([]Tag, error) {
	tags := []Tag{}
	err := spiffy.DefaultDb().QueryInTransaction(`select * from tag where tag_value % $1 order by similarity(tag_value, $1) desc;`, tx, query).OutMany(&tags)
	return tags, err
}

// SearchTagsRandom searches tags taking a randomly selected count.
func SearchTagsRandom(query string, count int, tx *sql.Tx) ([]Tag, error) {
	tags := []Tag{}
	err := spiffy.DefaultDb().QueryInTransaction(`select * from tag where tag_value % $1 order by gen_random_uuid() limit $2;`, tx, query, count).OutMany(&tags)
	return tags, err
}

// SetTagValue sets a tag value
func SetTagValue(tagID int64, tagValue string, tx *sql.Tx) error {
	return spiffy.DefaultDb().ExecInTransaction(`update tag set tag_value = $1 where id = $2`, tx, tagValue, tagID)
}

// DeleteTagByID deletes a tag.
func DeleteTagByID(tagID int64, tx *sql.Tx) error {
	return spiffy.DefaultDb().ExecInTransaction(`delete from tag where id = $1`, tx, tagID)
}

// DeleteTagAndVotesByID deletes an tag fully.
func DeleteTagAndVotesByID(tagID int64, tx *sql.Tx) error {
	err := spiffy.DefaultDb().ExecInTransaction(`delete from vote_summary where tag_id = $1`, tx, tagID)
	if err != nil {
		return err
	}
	err = spiffy.DefaultDb().ExecInTransaction(`delete from vote where tag_id = $1`, tx, tagID)
	if err != nil {
		return err
	}
	return DeleteTagByID(tagID, tx)
}

// MergeTags merges the fromTagID into the toTagID, deleting the fromTagID.
func MergeTags(fromTagID, toTagID int64, tx *sql.Tx) error {
	votes, err := GetVotesForTag(fromTagID, tx)
	if err != nil {
		return err
	}

	for _, vote := range votes {
		existingVote, err := GetVote(vote.UserID, vote.ImageID, toTagID, tx)
		if err != nil {
			return err
		}

		if existingVote.IsZero() {
			err = SetVoteTagID(vote.UserID, vote.ImageID, fromTagID, toTagID, tx)
			if err != nil {
				return err
			}
		} else {
			err = DeleteVote(vote.UserID, vote.ImageID, fromTagID, tx)
			if err != nil {
				return err
			}
		}
	}

	links, err := GetVoteSummariesForTag(fromTagID, tx)
	if err != nil {
		return err
	}

	for _, link := range links {
		existingLink, err := GetVoteSummary(link.ImageID, toTagID, tx)
		if err != nil {
			return err
		}

		if existingLink.IsZero() {
			err = SetVoteSummaryTagID(link.ImageID, fromTagID, toTagID, tx)
			if err != nil {
				return err
			}
		} else {
			err = ReconcileVoteSummaryTotals(link.ImageID, toTagID, tx)
			if err != nil {
				return err
			}

			err = DeleteVoteSummary(link.ImageID, fromTagID, tx)
			if err != nil {
				return err
			}
		}
	}

	return DeleteTagByID(fromTagID, tx)
}

// DeleteOrphanedTags deletes tags that have no vote_summary link to an image.
func DeleteOrphanedTags(tx *sql.Tx) error {
	err := spiffy.DefaultDb().ExecInTransaction(`delete from vote where not exists (select 1 from vote_summary vs where vs.tag_id = vote.tag_id);`, tx)
	if err != nil {
		return err
	}
	return spiffy.DefaultDb().ExecInTransaction(`delete from tag where not exists (select 1 from vote_summary vs where vs.tag_id = tag.id);`, tx)
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
