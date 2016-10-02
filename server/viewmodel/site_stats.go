package viewmodel

import (
	"database/sql"
	"time"

	"github.com/wcharczuk/giffy/server/model"
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
	imageCountQuery := `select coalesce(count(*), 0) as value from image;`
	tagCountQuery := `select coalesce(count(*), 0) as value from tag;`
	userCountQuery := `select coalesce(count(*), 0) as value from users;`
	karmaTotalQuery := `select coalesce(sum(votes_total), 0) as value from vote_summary;`
	orphanedTagCountQuery := `select coalesce(count(*), 0) from tag t where not exists (select 1 from vote_summary vs where vs.tag_id = t.id);`

	var userCount int
	var imageCount int
	var tagCount int
	var karmaTotal int
	var orphanedTagCount int

	err := model.DB().QueryInTransaction(userCountQuery, tx).Scan(&userCount)
	if err != nil {
		return nil, err
	}
	err = model.DB().QueryInTransaction(imageCountQuery, tx).Scan(&imageCount)
	if err != nil {
		return nil, err
	}
	err = model.DB().QueryInTransaction(tagCountQuery, tx).Scan(&tagCount)
	if err != nil {
		return nil, err
	}
	err = model.DB().QueryInTransaction(karmaTotalQuery, tx).Scan(&karmaTotal)
	if err != nil {
		return nil, err
	}
	err = model.DB().QueryInTransaction(orphanedTagCountQuery, tx).Scan(&orphanedTagCount)
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
