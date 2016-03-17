package model

import (
	"database/sql"
	"time"

	"github.com/blendlabs/spiffy"
)

// VoteLog is a log of votes by a user.
type VoteLog struct {
	UserID       int64     `json:"user_id" db:"user_id"`
	ImageID      int64     `json:"image_id" db:"image_id"`
	TagID        int64     `json:"tag_id" db:"tag_id"`
	TimestampUTC time.Time `json:"timestamp_utc" db:"timestamp_utc"`
	IsUpvote     bool      `json:"is_upvote" db:"is_upvote"`
}

// TableName returns the name of the table.
func (vl VoteLog) TableName() string {
	return "vote_log"
}

// IsZero returns if the object is set or not.
func (vl VoteLog) IsZero() bool {
	return vl.UserID == 0
}

// NewVoteLog returns a new vote log entry.
func NewVoteLog(userID, imageID, tagID int64, isUpvote bool) *VoteLog {
	return &VoteLog{
		UserID:       userID,
		ImageID:      imageID,
		TagID:        tagID,
		TimestampUTC: time.Now().UTC(),
		IsUpvote:     isUpvote,
	}
}

// GetVoteLogsForUserID gets all the vote log entries for a user.
func GetVoteLogsForUserID(userID int64, tx *sql.Tx) ([]VoteLog, error) {
	logEntries := []VoteLog{}
	err := spiffy.DefaultDb().QueryInTransaction("select * from vote_log where user_id = $1 order by timestamp_utc desc", tx, userID).OutMany(&logEntries)
	return logEntries, err
}

// GetUserVoteForImage gets a user's vote for an image.
func GetUserVoteForImage(userID, imageID int64, tx *sql.Tx) (*VoteLog, error) {
	voteLog := VoteLog{}
	err := spiffy.DefaultDb().QueryInTransaction("select * from vote_log where user_id = $1 and image_id = $2", tx, userID, imageID).Out(&voteLog)
	return &voteLog, err
}

// GetUserVoteForTag gets a user's vote for an tag.
func GetUserVoteForTag(userID, tagID int64, tx *sql.Tx) (*VoteLog, error) {
	voteLog := VoteLog{}
	err := spiffy.DefaultDb().QueryInTransaction("select * from vote_log where user_id = $1 and tag_id = $2", tx, userID, tagID).Out(&voteLog)
	return &voteLog, err
}

// GetUserVoteForImageOrTag gets a user's vote for an image or a tag.
func GetUserVoteForImageOrTag(userID, imageID, tagID int64, tx *sql.Tx) (*VoteLog, error) {
	voteLog := VoteLog{}
	err := spiffy.DefaultDb().QueryInTransaction("select * from vote_log where user_id = $1 and (tag_id = $2 or image_id = $3)", tx, userID, tagID, imageID).Out(&voteLog)
	return &voteLog, err
}

// GetUserVoteForImageAndTag gets a user's vote for an image or a tag.
func GetUserVoteForImageAndTag(userID, imageID, tagID int64, tx *sql.Tx) (*VoteLog, error) {
	voteLog := VoteLog{}
	err := spiffy.DefaultDb().QueryInTransaction("select * from vote_log where user_id = $1 and tag_id = $2 and image_id = $3", tx, userID, tagID, imageID).Out(&voteLog)
	return &voteLog, err
}
