package model

// ImageStats contains some basic stats for a given image.
type ImageStats struct {
	ImageID    int64 `json:"image_id" db:"image_id,readonly"`
	Image      *Image
	VotesTotal int `json:"votes_total" db:"votes_total,readonly"`
	Searches   int `json:"searches" db:"searches,readonly"`
}

// TableName returns the table name.
func (is ImageStats) TableName() string {
	return "image_stats"
}
