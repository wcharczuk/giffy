package model

import "time"

type ImageTagVotes struct {
	ImageID      int64     `json:"image_id" db:"image_id"`
	TagID        int64     `json:"tag_id" db:"tag_id"`
	LastVoteUTC  time.Time `json:"last_vote_utc" db:"last_vote_utc"`
	LastVoteBy   int64     `json:"last_vote_by" db:"last_vote_by"`
	VotesFor     int       `json:"votes_for" db:"votes_for"`
	VotesAgainst int       `json:"votes_for" db:"votes_for"`
	VotesTotal   int       `json:"votes_total" db:"votes_total"`
}

func (itv ImageTagVote) TableName() string {
	return "image_tag_votes"
}
