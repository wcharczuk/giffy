package model

import "time"

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
