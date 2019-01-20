package model

import (
	"database/sql"
	"time"
)

// Vote is a vote by a user for an image and tag.
type Vote struct {
	UserID     int64     `json:"-" db:"user_id,pk"`
	UserUUID   string    `json:"user_uuid" db:"user_uuid,readonly"`
	ImageID    int64     `json:"-" db:"image_id,pk"`
	ImageUUID  string    `json:"image_uuid" db:"image_uuid,readonly"`
	TagID      int64     `json:"-" db:"tag_id,pk"`
	TagUUID    string    `json:"tag_uuid" db:"tag_uuid,readonly"`
	CreatedUTC time.Time `json:"created_utc" db:"created_utc"`
	IsUpvote   bool      `json:"is_upvote" db:"is_upvote"`
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
	return r.Scan(&v.UserID, &v.ImageID, &v.TagID, &v.CreatedUTC, &v.IsUpvote, &v.UserUUID, &v.ImageUUID, &v.TagUUID)
}

// NewVote returns a new vote log entry.
func NewVote(userID, imageID, tagID int64, isUpvote bool) *Vote {
	return &Vote{
		UserID:     userID,
		ImageID:    imageID,
		TagID:      tagID,
		CreatedUTC: time.Now().UTC(),
		IsUpvote:   isUpvote,
	}
}
