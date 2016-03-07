package model

import "time"

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
