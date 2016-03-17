package model

import (
	"database/sql"
	"time"

	"github.com/blendlabs/spiffy"
)

// Vote is a vote by a user for an image and tag.
type Vote struct {
	UserID       int64     `json:"user_id" db:"user_id"`
	ImageID      int64     `json:"image_id" db:"image_id"`
	TagID        int64     `json:"tag_id" db:"tag_id"`
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

// GetVotesForUser gets all the vote log entries for a user.
func GetVotesForUser(userID int64, tx *sql.Tx) ([]Vote, error) {
	votes := []Vote{}
	err := spiffy.DefaultDb().QueryInTransaction("select * from vote where user_id = $1 order by timestamp_utc desc", tx, userID).OutMany(&votes)
	return votes, err
}

// GetVotesForUserForImage gets the votes for an image by a user.
func GetVotesForUserForImage(userID, imageID int64, tx *sql.Tx) ([]Vote, error) {
	votes := []Vote{}
	err := spiffy.DefaultDb().QueryInTransaction("select * from vote where user_id = $1 and image_id = $2 order by timestamp_utc desc", tx, userID, imageID).OutMany(&votes)
	return votes, err
}

// GetVotesForUserForTag gets the votes for an image by a user.
func GetVotesForUserForTag(userID, tagID int64, tx *sql.Tx) ([]Vote, error) {
	votes := []Vote{}
	err := spiffy.DefaultDb().QueryInTransaction("select * from vote where user_id = $1 and tag_id = $2 order by timestamp_utc desc", tx, userID, tagID).OutMany(&votes)
	return votes, err
}

// GetVote gets a user's vote for an image and a tag.
func GetVote(userID, imageID, tagID int64, tx *sql.Tx) (*Vote, error) {
	voteLog := Vote{}
	err := spiffy.DefaultDb().QueryInTransaction("select * from vote where user_id = $1 and tag_id = $2 and image_id = $3", tx, userID, tagID, imageID).Out(&voteLog)
	return &voteLog, err
}
