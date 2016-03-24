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

	UserCountDaily  []StatAtTime `json:"user_count_daily"`
	ImageCountDaily []StatAtTime `json:"image_count_daily"`
	TagCountDaily   []StatAtTime `json:"tag_count_daily"`
	KarmaTotalDaily []StatAtTime `json:"karma_total_daily"`
}

func GetSiteStats() (*SiteStats, error) {
	imageCountQuery := `select count(*) as value from image;`
	tagCountQuery := `select count(*) as value from tag;`
	userCountQuery := `select count(*) as value from users;`
	karmaTotalQuery := `select sum(votes_total) as value from vote_summary;`
	orphanedTagCountQuery := `select count(*) from tag t where not exists (select 1 from vote_summary vs where vs.tag_id = t.id);`

	userCountDailyQuery := `
select 
    date_part('year', created_utc) as year,
    date_part('month', created_utc) as month,
    date_part('day', created_utc) as day,
    count(*) as value,
    'users' as label
from users u
group by 
    date_part('year', created_utc), 
    date_part('month', created_utc), 
    date_part('day', created_utc);`

	imageCountDailyQuery := `
select 
    date_part('year', created_utc) as year, 
    date_part('month', created_utc) as month, 
    date_part('day', created_utc) as day,
    count(*) as value,
    'image' as label
from image
group by 
    date_part('year', created_utc), 
    date_part('month', created_utc), 
    date_part('day', created_utc);`

	tagCountDailyQuery := `
select 
    date_part('year', created_utc) as year, 
    date_part('month', created_utc) as month, 
    date_part('day', created_utc) as day,
    count(*) as value,
    'tag' as label
from tag
group by 
    date_part('year', created_utc), 
    date_part('month', created_utc), 
    date_part('day', created_utc);`

	karmaTotalDailyQuery := `
select 
    year,
    month,
    day,
    sum(votes_for) - sum(votes_against) as value,
    'karma_total' as label
from 
    (
        select
            date_part('year', created_utc) as year, 
            date_part('month', created_utc) as month, 
            date_part('day', created_utc) as day,
            v.image_id,
            v.tag_id,
            case when v.is_upvote = true then 1 else 0 end as votes_for,
            case when v.is_upvote = true then 0 else 1 end as votes_against
        from 
            vote v
    ) votes
group by 
    year,
    month,
    day,
    image_id,
    tag_id;`

	var userCount int
	var imageCount int
	var tagCount int
	var karmaTotal int
	var orphanedTagCount int

	var userDaily []StatAtTime
	var imageDaily []StatAtTime
	var tagDaily []StatAtTime
	var karmaTotalDaily []StatAtTime

	err := spiffy.DefaultDb().Query(userCountQuery).Scan(&userCount)
	if err != nil {
		return nil, err
	}
	err = spiffy.DefaultDb().Query(imageCountQuery).Scan(&imageCount)
	if err != nil {
		return nil, err
	}
	err = spiffy.DefaultDb().Query(tagCountQuery).Scan(&tagCount)
	if err != nil {
		return nil, err
	}
	err = spiffy.DefaultDb().Query(karmaTotalQuery).Scan(&karmaTotal)
	if err != nil {
		return nil, err
	}
	err = spiffy.DefaultDb().Query(orphanedTagCountQuery).Scan(&orphanedTagCount)
	if err != nil {
		return nil, err
	}

	makeDailyCollector := func(values *[]StatAtTime) spiffy.RowsConsumer {
		return func(r *sql.Rows) error {
			var year int
			var month int
			var day int

			stat := StatAtTime{}
			scanErr := r.Scan(&year, &month, &day, &stat.Value, &stat.Label)
			if scanErr != nil {
				return scanErr
			}
			stat.TimestampUTC = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

			*values = append(*values, stat)
			return nil
		}
	}

	err = spiffy.DefaultDb().Query(userCountDailyQuery).Each(makeDailyCollector(&userDaily))
	if err != nil {
		return nil, err
	}

	err = spiffy.DefaultDb().Query(imageCountDailyQuery).Each(makeDailyCollector(&imageDaily))
	if err != nil {
		return nil, err
	}

	err = spiffy.DefaultDb().Query(tagCountDailyQuery).Each(makeDailyCollector(&tagDaily))
	if err != nil {
		return nil, err
	}

	err = spiffy.DefaultDb().Query(karmaTotalDailyQuery).Each(makeDailyCollector(&karmaTotalDaily))
	if err != nil {
		return nil, err
	}

	return &SiteStats{
		UserCount:        userCount,
		ImageCount:       imageCount,
		TagCount:         tagCount,
		KarmaTotal:       karmaTotal,
		OrphanedTagCount: orphanedTagCount,
		UserCountDaily:   userDaily,
		ImageCountDaily:  imageDaily,
		TagCountDaily:    tagDaily,
		KarmaTotalDaily:  karmaTotalDaily,
	}, nil

}
