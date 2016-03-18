package model

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/blendlabs/spiffy"
)

// Vote is a vote by a user for an image and tag.
type Vote struct {
	UserID       int64     `json:"-" db:"user_id"`
	UserUUID     string    `json:"user_uuid" db:"user_uuid,readonly"`
	ImageID      int64     `json:"-" db:"image_id"`
	ImageUUID    string    `json:"image_uuid" db:"image_uuid,readonly"`
	TagID        int64     `json:"-" db:"tag_id"`
	TagUUID      string    `json:"tag_uuid" db:"tag_uuid,readonly"`
	TimestampUTC time.Time `json:"timestamp_utc" db:"timestamp_utc"`
	IsUpvote     bool      `json:"is_upvote" db:"is_upvote"`
}

// TableName returns the name of the table.
func (v Vote) TableName() string {
	return "vote"
}

// IsZero returns if the object is set or not.
func (v Vote) IsZero() bool {
	return v.UserID == 0
}

// Populate skips spiffy struct parsing.
func (v *Vote) Populate(r *sql.Rows) error {
	return r.Scan(&v.UserID, &v.ImageID, &v.TagID, &v.TimestampUTC, &v.IsUpvote, &v.UserUUID, &v.ImageUUID, &v.TagUUID)
}

// NewVote returns a new vote log entry.
func NewVote(userID, imageID, tagID int64, isUpvote bool) *Vote {
	return &Vote{
		UserID:       userID,
		ImageID:      imageID,
		TagID:        tagID,
		TimestampUTC: time.Now().UTC(),
		IsUpvote:     isUpvote,
	}
}

func getVotesQuery(whereClause string) string {
	return fmt.Sprintf(`
select 
	v.* 
	, u.uuid as user_uuid
	, i.uuid as image_uuid
	, t.uuid as tag_uuid
from 
	vote v
	join users u on v.user_id = u.id
	join image i on v.image_id = i.id
	join tag t on v.tag_id = t.id
%s
order by 
	v.timestamp_utc desc;
`, whereClause)
}

// GetVotesForUser gets all the vote log entries for a user.
func GetVotesForUser(userID int64, tx *sql.Tx) ([]Vote, error) {
	votes := []Vote{}
	err := spiffy.DefaultDb().QueryInTransaction(getVotesQuery("where v.user_id = $1"), tx, userID).OutMany(&votes)
	return votes, err
}

// GetVotesForUserForImage gets the votes for an image by a user.
func GetVotesForUserForImage(userID, imageID int64, tx *sql.Tx) ([]Vote, error) {
	votes := []Vote{}
	err := spiffy.DefaultDb().QueryInTransaction(getVotesQuery("where v.user_id = $1 and v.image_id = $2"), tx, userID, imageID).OutMany(&votes)
	return votes, err
}

// GetVotesForUserForTag gets the votes for an image by a user.
func GetVotesForUserForTag(userID, tagID int64, tx *sql.Tx) ([]Vote, error) {
	votes := []Vote{}
	err := spiffy.DefaultDb().QueryInTransaction(getVotesQuery("where v.user_id = $1 and v.tag_id = $2"), tx, userID, tagID).OutMany(&votes)
	return votes, err
}

// GetVote gets a user's vote for an image and a tag.
func GetVote(userID, imageID, tagID int64, tx *sql.Tx) (*Vote, error) {
	voteLog := Vote{}
	err := spiffy.DefaultDb().QueryInTransaction(getVotesQuery("where v.user_id = $1 and v.image_id = $2 and v.tag_id = $3"), tx, userID, imageID, tagID).Out(&voteLog)
	return &voteLog, err
}
