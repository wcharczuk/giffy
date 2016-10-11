package viewmodel

import (
	"database/sql"

	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/model"
)

// ImageStats contains some basic stats for a given image.
type ImageStats struct {
	ImageID    int64 `json:"image_id" db:"image_id,readonly"`
	VotesTotal int   `json:"votes_total" db:"votes_total,readonly"`
	Searches   int   `json:"searches" db:"searches,readonly"`
}

// TableName returns the table name.
func (is ImageStats) TableName() string {
	return "image_stats"
}

// GetImageStats gets image stats.
func GetImageStats(imageID int64, txs ...*sql.Tx) (*ImageStats, error) {
	tx := spiffy.OptionalTx(txs...)

	var results ImageStats
	query := `
	select
		i.id as image_id
		, (select sum(votes_total) from vote_summary where image_id = $1) as votes_total
		, (select count(image_id) from search_history where image_id = $1) as searches
	from
		image i
	where
		i.id = $1
	`

	err := model.DB().QueryInTransaction(query, tx, imageID).Out(&results)
	if err != nil {
		return nil, err
	}

	return &results, nil
}
