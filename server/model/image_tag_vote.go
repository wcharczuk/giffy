package model

import (
	"database/sql"
	"time"

	"github.com/blendlabs/spiffy"
)

// ImageTagVotes is the link between an image and a tag.
type ImageTagVotes struct {
	ImageID      int64     `json:"image_id" db:"image_id,pk"`
	TagID        int64     `json:"tag_id" db:"tag_id,pk"`
	LastVoteUTC  time.Time `json:"last_vote_utc" db:"last_vote_utc"`
	LastVoteBy   int64     `json:"last_vote_by" db:"last_vote_by"`
	VotesFor     int       `json:"votes_for" db:"votes_for"`
	VotesAgainst int       `json:"votes_against" db:"votes_against"`
	VotesTotal   int       `json:"votes_total" db:"votes_total"`
}

// IsZero returns if an image has been set.
func (itv ImageTagVotes) IsZero() bool {
	return itv.ImageID == 0 || itv.TagID == 0
}

// TableName returns the tablename for an object.
func (itv ImageTagVotes) TableName() string {
	return "image_tag_votes"
}

// NewImageTagVote returns a new instance for an ImageTagVotes.
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

// SetTagVotes updates the votes for a tag to an image.
func SetTagVotes(imageID, tagID int64, votesFor, votesAgainst int, tx *sql.Tx) error {
	votesTotal := votesFor - votesAgainst
	return spiffy.DefaultDb().ExecInTransaction(`update image_tag_votes set votes_for = $1, votes_against = $2, votes_total = $3 where image_id = $4 and tag_id = $5`, tx, votesFor, votesAgainst, votesTotal, imageID, tagID)
}

// Vote votes for a tag for an image in the db.
func Vote(userID, imageID, tagID int64, isUpvote bool, tx *sql.Tx) error {
	existing, existingErr := GetImageTagVote(imageID, tagID, tx)
	if existingErr != nil {
		return existingErr
	}

	if existing.IsZero() {
		var itv *ImageTagVotes
		if isUpvote {
			itv = NewImageTagVote(imageID, tagID, userID, time.Now().UTC(), 1, 0)
		} else {
			itv = NewImageTagVote(imageID, tagID, userID, time.Now().UTC(), 0, 1)
		}
		err := spiffy.DefaultDb().CreateInTransaction(itv, tx)
		if err != nil {
			return err
		}
	} else {
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
	}
	logEntry := NewVoteLog(userID, imageID, tagID, isUpvote)
	return spiffy.DefaultDb().CreateInTransaction(logEntry, tx)
}

// GetImageTagVote fetches an ImageTagVotes by constituent pks.
func GetImageTagVote(imageID, tagID int64, tx *sql.Tx) (*ImageTagVotes, error) {
	var imv ImageTagVotes
	query := `select * from image_tag_votes where image_id = $1 and tag_id = $2`
	err := spiffy.DefaultDb().QueryInTransaction(query, tx, imageID, tagID).Out(&imv)
	return &imv, err
}

// GetImagesForTagID gets all the images attached to a tag.
func GetImagesForTagID(tagID int64, tx *sql.Tx) ([]Image, error) {
	var imageIDs []imageSignature
	err := spiffy.DefaultDb().
		QueryInTransaction(`select image_id as id from image_tag_votes itv where tag_id = $1`, tx, tagID).OutMany(&imageIDs)
	if err != nil {
		return nil, err
	}

	return GetImagesByID(imageSignatures(imageIDs).AsInt64s(), tx)
}

// GetTagsForImageID returns all the tags for an image.
func GetTagsForImageID(imageID int64, tx *sql.Tx) ([]Tag, error) {
	var tags []Tag
	query := `
select 
	t.*
	, itv.image_id
	, itv.votes_for
	, itv.votes_against
	, itv.votes_total
    , row_number() over (partition by itv.image_id order by itv.votes_total desc) as vote_rank
from 
	tag t 
	join image_tag_votes itv on t.id = itv.tag_id 
where 
	itv.image_id = $1
order by
	itv.votes_total desc;
`
	err := spiffy.DefaultDb().QueryInTransaction(query, tx, imageID).Each(func(r *sql.Rows) error {
		t := &Tag{}
		popErr := t.PopulateExtra(r)
		if popErr != nil {
			return popErr
		}
		tags = append(tags, *t)
		return nil
	})
	return tags, err
}
