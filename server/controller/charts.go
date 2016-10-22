package controller

import (
	"database/sql"
	"time"

	"github.com/wcharczuk/giffy/server/model"
	chart "github.com/wcharczuk/go-chart"
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

// Marshal returns the dayCount as a usable value.
func (dc *dayCount) Marshal() (time.Time, float64) {
	return time.Date(dc.Year, time.Month(dc.Month), dc.Day, 0, 0, 0, 0, time.UTC), float64(dc.Count)
}

func (c Chart) getSearchChartAction(rc *web.RequestContext) web.ControllerResult {
	data := []dayCount{}
	err := model.DB().QueryInTx(`
select
	date_part('year', timestamp_utc) as year
	, date_part('month', timestamp_utc) as month
	, date_part('day', timestamp_utc) as day
	, count(*) as count
	from search_history
where
	timestamp_utc > $1
group by
	date_part('year', timestamp_utc)
	, date_part('month', timestamp_utc)
	, date_part('day', timestamp_utc)
`, rc.Tx(), time.Now().UTC().AddDate(0, -1, 0)).OutMany(&data)

	if err != nil {
		return rc.API().InternalError(err)
	}

	var xvalues []time.Time
	var yvalues []float64
	for _, dc := range data {
		xv, yv := dc.Marshal()
		xvalues = append(xvalues, xv)
		yvalues = append(yvalues, yv)
	}

	mainSeries := chart.TimeSeries{
		Name: "Search Count By Day",
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.ColorBlue,
			FillColor:   chart.ColorBlue.WithAlpha(100),
		},
		XValues: xvalues,
		YValues: yvalues,
	}

	graph := chart.Chart{
		Width:  960,
		Height: 128,
		Series: []chart.Series{
			mainSeries,
		},
	}

	rc.Response.Header().Set("Content-Type", "image/png")
	graph.Render(chart.PNG, rc.Response)
	return nil
}

// Register registers the controller.
func (c Chart) Register(app *web.App) {
	app.GET("/chart/searches", c.getSearchChartAction)
}
