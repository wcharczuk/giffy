package viewmodel

import (
	"database/sql"
	"time"

	"github.com/blendlabs/spiffy"
)

// StatAtTime is a stat value at a time.
type StatAtTime struct {
	TimestampUTC time.Time `json:"timestamp_utc"`
	Value        float64   `json:"value"`
	Label        string    `json:"label"`
}

// SiteStats is the current counts for various things on the site.
type SiteStats struct {
	UserCount        int `json:"user_count"`
	ImageCount       int `json:"image_count"`
	TagCount         int `json:"tag_count"`
	KarmaTotal       int `json:"karma_total"`
	OrphanedTagCount int `json:"orphaned_tag_count"`
}

// GetSiteStats returns the stats for the site.
func GetSiteStats(tx *sql.Tx) (*SiteStats, error) {
	imageCountQuery := `select count(*) as value from image;`
	tagCountQuery := `select count(*) as value from tag;`
	userCountQuery := `select count(*) as value from users;`
	karmaTotalQuery := `select sum(votes_total) as value from vote_summary;`
	orphanedTagCountQuery := `select count(*) from tag t where not exists (select 1 from vote_summary vs where vs.tag_id = t.id);`

	var userCount int
	var imageCount int
	var tagCount int
	var karmaTotal int
	var orphanedTagCount int

	var userDaily []StatAtTime
	var imageDaily []StatAtTime
	var tagDaily []StatAtTime
	var karmaTotalDaily []StatAtTime

	err := spiffy.DefaultDb().QueryInTransaction(userCountQuery, tx).Scan(&userCount)
	if err != nil {
		return nil, err
	}
	err = spiffy.DefaultDb().QueryInTransaction(imageCountQuery, tx).Scan(&imageCount)
	if err != nil {
		return nil, err
	}
	err = spiffy.DefaultDb().QueryInTransaction(tagCountQuery, tx).Scan(&tagCount)
	if err != nil {
		return nil, err
	}
	err = spiffy.DefaultDb().QueryInTransaction(karmaTotalQuery, tx).Scan(&karmaTotal)
	if err != nil {
		return nil, err
	}
	err = spiffy.DefaultDb().QueryInTransaction(orphanedTagCountQuery, tx).Scan(&orphanedTagCount)
	if err != nil {
		return nil, err
	}

	return &SiteStats{
		UserCount:        userCount,
		ImageCount:       imageCount,
		TagCount:         tagCount,
		KarmaTotal:       karmaTotal,
		OrphanedTagCount: orphanedTagCount,
	}, nil

}
