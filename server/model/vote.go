package model

import (
	"database/sql"
	"fmt"
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
	v.created_utc desc;
`, whereClause)
}

// GetVotesForUser gets all the vote log entries for a user.
func GetVotesForUser(userID int64, tx *sql.Tx) ([]Vote, error) {
	votes := []Vote{}
	err := DB().QueryInTransaction(getVotesQuery("where v.user_id = $1"), tx, userID).OutMany(&votes)
	return votes, err
}

// GetVotesForImage gets all the votes log entries for an image.
func GetVotesForImage(imageID int64, tx *sql.Tx) ([]Vote, error) {
	votes := []Vote{}
	err := DB().QueryInTransaction(getVotesQuery("where v.image_id = $1"), tx, imageID).OutMany(&votes)
	return votes, err
}

// GetVotesForTag gets all the votes log entries for an image.
func GetVotesForTag(tagID int64, tx *sql.Tx) ([]Vote, error) {
	votes := []Vote{}
	err := DB().QueryInTransaction(getVotesQuery("where v.tag_id = $1"), tx, tagID).OutMany(&votes)
	return votes, err
}

// GetVotesForUserForImage gets the votes for an image by a user.
func GetVotesForUserForImage(userID, imageID int64, tx *sql.Tx) ([]Vote, error) {
	votes := []Vote{}
	err := DB().QueryInTransaction(getVotesQuery("where v.user_id = $1 and v.image_id = $2"), tx, userID, imageID).OutMany(&votes)
	return votes, err
}

// GetVotesForUserForTag gets the votes for an image by a user.
func GetVotesForUserForTag(userID, tagID int64, tx *sql.Tx) ([]Vote, error) {
	votes := []Vote{}
	err := DB().QueryInTransaction(getVotesQuery("where v.user_id = $1 and v.tag_id = $2"), tx, userID, tagID).OutMany(&votes)
	return votes, err
}

// GetVote gets a user's vote for an image and a tag.
func GetVote(userID, imageID, tagID int64, tx *sql.Tx) (*Vote, error) {
	voteLog := Vote{}
	err := DB().QueryInTransaction(getVotesQuery("where v.user_id = $1 and v.image_id = $2 and v.tag_id = $3"), tx, userID, imageID, tagID).Out(&voteLog)
	return &voteLog, err
}

// SetVoteTagID sets the tag_id for a vote object.
func SetVoteTagID(userID, imageID, oldTagID, newTagID int64, tx *sql.Tx) error {
	return DB().ExecInTransaction(`update vote set tag_id = $1 where user_id = $2 and image_id = $3 and tag_id = $4`, tx, newTagID, userID, imageID, oldTagID)
}

// DeleteVote deletes a vote.
func DeleteVote(userID, imageID, tagID int64, tx *sql.Tx) error {
	return DB().ExecInTransaction(`DELETE from vote where user_id = $1 and image_id = $2 and tag_id = $3`, tx, userID, imageID, tagID)
}
