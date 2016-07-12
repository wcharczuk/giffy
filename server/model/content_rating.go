package model

import (
	"database/sql"

	"github.com/blendlabs/spiffy"
)

const (
	// ContentRatingG = 1
	ContentRatingG = 1

	// ContentRatingPG = 2
	ContentRatingPG = 2

	// ContentRatingPG13 = 3
	ContentRatingPG13 = 3

	// ContentRatingR = 4
	ContentRatingR = 4

	// ContentRatingNR = 5
	ContentRatingNR = 5

	// ContentRatingAll will returns all images
	ContentRatingAll = 6
)

// NewContentRating returns a new ContentRating instance.
func NewContentRating() *ContentRating {
	return &ContentRating{}
}

// ContentRating is a rating for content similar to the MPAA ratings for movies.
type ContentRating struct {
	ID          int    `json:"id" db:"id,pk"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
}

// TableName returns the table name.
func (cr *ContentRating) TableName() string {
	return "content_rating"
}

// Populate skips spiffy reflection lookups for populting rows.
func (cr *ContentRating) Populate(rows *sql.Rows) error {
	return rows.Scan(&cr.ID, &cr.Name, &cr.Description)
}

// IsZero returns if the object is empty or not.
func (cr *ContentRating) IsZero() bool {
	return cr.ID == 0 && len(cr.Name) == 0
}

// GetContentRatingByName gets a content rating by name.
func GetContentRatingByName(name string, tx *sql.Tx) (*ContentRating, error) {
	var rating ContentRating
	err := spiffy.DefaultDb().QueryInTransaction(
		`SELECT * from content_rating where name = $1`, tx, name,
	).Out(&rating)
	return &rating, err
}
