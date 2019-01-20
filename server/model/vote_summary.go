package model

import (
	"time"
)

// VoteSummary is the link between an image and a tag.
type VoteSummary struct {
	ImageID        int64     `json:"image_id" db:"image_id,pk"`
	ImageUUID      string    `json:"image_uuid" db:"image_uuid,readonly"`
	TagID          int64     `json:"tag_id" db:"tag_id,pk"`
	TagUUID        string    `json:"tag_uuid" db:"tag_uuid,readonly"`
	CreatedUTC     time.Time `json:"created_utc" db:"created_utc"`
	LastVoteUTC    time.Time `json:"last_vote_utc" db:"last_vote_utc"`
	LastVoteBy     int64     `json:"last_vote_by" db:"last_vote_by"`
	LastVoteByUUID string    `json:"last_vote_by_uuid" db:"last_vote_by_uuid,readonly"`
	VotesFor       int       `json:"votes_for" db:"votes_for"`
	VotesAgainst   int       `json:"votes_against" db:"votes_against"`
	VotesTotal     int       `json:"votes_total" db:"votes_total"`
}

// IsZero returns if an image has been set.
func (itv VoteSummary) IsZero() bool {
	return itv.ImageID == 0 || itv.TagID == 0
}

// TableName returns the tablename for an object.
func (itv VoteSummary) TableName() string {
	return "vote_summary"
}

// NewVoteSummary returns a new instance for an ImageTagVotes.
func NewVoteSummary(imageID, tagID, lastVoteBy int64, lastVoteUTC time.Time) *VoteSummary {
	return &VoteSummary{
		CreatedUTC:  time.Now().UTC(),
		ImageID:     imageID,
		TagID:       tagID,
		LastVoteUTC: lastVoteUTC,
		LastVoteBy:  lastVoteBy,
	}
}
