package controller

import (
	"database/sql"
	"time"

	"github.com/wcharczuk/giffy/server/model"
	web "github.com/wcharczuk/go-web"
)

// Chart is a controller for common chart endpoints.
type Chart struct{}

type dayCount struct {
	Year  int `db:"year,readonly"`
	Month int `db:"month,readonly"`
	Day   int `db:"day,readonly"`
	Count int `db:"count,readonly"`
}

// Populate manually populates the object.
func (dc *dayCount) Populate(row *sql.Rows) error {
	return row.Scan(&dc.Year, &dc.Month, &dc.Day, &dc.Count)
}

func (c Chart) getSearchChartAction(rc *web.RequestContext) web.ControllerResult {
	var data []dayCount
	err := model.DB().QueryInTx(`
select
	datepart('year', timestamp_utc) as year
	, datepart('month', timestamp_utc) as month
	, datepart('day', timestamp_utc) as day
	, count(*) as count
from search_history
where
	timestamp_utc > $1
group by
	datepart('year', timestamp_utc)
	, datepart('month', timestamp_utc)
	, datepart('day', timestamp_utc)
`, rc.Tx(), time.Now().UTC().AddDate(0, -1, 0)).OutMany(&data)

}
