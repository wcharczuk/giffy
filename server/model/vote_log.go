package model

import "time"

type VoteLog struct {
	UserID       int       `json:"user_id" db:"user_id"`
	ImageID      int       `json:"image_id" db:"image_id"`
	TagID        int       `json:"tag_id" db:"tag_id"`
	TimestampUTC time.Time `json:"timestamp_utc" db:"timestamp_utc"`
	IsUpvote     bool      `json:"is_upvote" db:"is_upvote"`
}

func (vl VoteLog) TableName() string {
	return "vote_log"
}

func NewVoteLog(userID, imageID, tagID int, isUpvote bool) *VoteLog {
	return &VoteLog{
		UserID:       userID,
		ImageID:      imageID,
		TagID:        tagID,
		TimestampUTC: time.Now().UTC(),
		IsUpvote:     isUpvote,
	}
}
