package model

import (
	"database/sql"
	"time"
)

// DayCounts is an array of DayCount.
type DayCounts []DayCount

// ChartData returns the day count list as a chart data item.
func (dcs DayCounts) ChartData() ([]time.Time, []float64) {
	var xvalues []time.Time
	var yvalues []float64
	for _, dc := range dcs {
		xv, yv := dc.ChartData()
		xvalues = append(xvalues, xv)
		yvalues = append(yvalues, yv)
	}
	return xvalues, yvalues
}

// DayCount represents a count per day.
type DayCount struct {
	Year  int     `db:"year,readonly"`
	Month int     `db:"month,readonly"`
	Day   int     `db:"day,readonly"`
	Count float64 `db:"count,readonly"`
}

// Populate manually populates the object.
func (dc *DayCount) Populate(row *sql.Rows) error {
	return row.Scan(&dc.Year, &dc.Month, &dc.Day, &dc.Count)
}

// ChartData returns the dayCount as a usable value.
func (dc *DayCount) ChartData() (time.Time, float64) {
	return time.Date(dc.Year, time.Month(dc.Month), dc.Day, 0, 0, 0, 0, time.UTC), float64(dc.Count)
}
