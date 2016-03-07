package model

import (
	"database/sql"
	"time"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/spiffy"
)

type ImageTagVotes struct {
	ImageID      int64     `json:"image_id" db:"image_id,pk"`
	TagID        int64     `json:"tag_id" db:"tag_id,pk"`
	LastVoteUTC  time.Time `json:"last_vote_utc" db:"last_vote_utc"`
	LastVoteBy   int64     `json:"last_vote_by" db:"last_vote_by"`
	VotesFor     int       `json:"votes_for" db:"votes_for"`
	VotesAgainst int       `json:"votes_against" db:"votes_against"`
	VotesTotal   int       `json:"votes_total" db:"votes_total"`
}

func (itv ImageTagVotes) IsZero() bool {
	return itv.ImageID == 0 || itv.TagID == 0
}

func (itv ImageTagVotes) TableName() string {
	return "image_tag_votes"
}

func NewImageTagVote(imageID, tagID, lastVoteBy int64, lastVoteUTC time.Time, votesFor, votesAgainst int) *ImageTagVotes {
	return &ImageTagVotes{
		ImageID:      imageID,
		TagID:        tagID,
		LastVoteUTC:  lastVoteUTC,
		LastVoteBy:   lastVoteBy,
		VotesFor:     votesFor,
		VotesAgainst: votesAgainst,
		VotesTotal:   votesFor - votesAgainst,
	}
}

func Vote(userID, imageID, tagID int64, isUpvote bool, tx *sql.Tx) error {
	existing, existingErr := GetImageTagVote(imageID, tagID, tx)
	if existingErr != nil {
		return existingErr
	}

	if existing != nil && !existing.IsZero() {
		if isUpvote {
			existing.VotesFor = existing.VotesFor + 1
		} else {
			existing.VotesAgainst = existing.VotesAgainst + 1
		}
		existing.LastVoteBy = userID
		existing.LastVoteUTC = time.Now().UTC()
		existing.VotesTotal = existing.VotesFor - existing.VotesAgainst

		updateErr := spiffy.DefaultDb().UpdateInTransaction(existing, tx)
		if updateErr != nil {
			return updateErr
		}

		logEntry := NewVoteLog(userID, imageID, tagID, isUpvote)
		return spiffy.DefaultDb().CreateInTransaction(logEntry, tx)
	}

	return exception.New("Invalid schema state; no `image_tag_votes` for image.")
}

func GetImageTagVote(imageID, tagID int64, tx *sql.Tx) (*ImageTagVotes, error) {
	var imv ImageTagVotes
	query := `select * from image_tag_votes where image_id = $1 and tag_id = $2`
	err := spiffy.DefaultDb().QueryInTransaction(query, tx, imageID, tagID).Out(&imv)
	return &imv, err
}
