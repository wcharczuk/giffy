package controller

import (
	"time"

	"github.com/wcharczuk/giffy/server/viewmodel"
	chart "github.com/wcharczuk/go-chart"
	web "github.com/wcharczuk/go-web"
)

// Chart is a controller for common chart endpoints.
type Chart struct{}

func (c Chart) getSearchChartAction(rc *web.RequestContext) web.ControllerResult {
	data, err := viewmodel.GetSearchesPerDay(time.Now().UTC().AddDate(0, -6, 0), rc.Tx())
	if err != nil {
		return rc.API().InternalError(err)
	}

	var width, height int
	if widthParam, err := rc.ParamInt("width"); err == nil {
		width = widthParam
	} else {
		width = 1280
	}

	if heightParam, err := rc.ParamInt("height"); err == nil {
		height = heightParam
	} else {
		height = 256
	}

	xvalues, yvalues := viewmodel.DayCounts(data).ChartData()

	mainSeries := chart.TimeSeries{
		Name: "Search Count By Day",
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.ColorBlue,
			//FillColor:   chart.ColorBlue.WithAlpha(100),
			FontSize: 8,
		},
		XValues: xvalues,
		YValues: yvalues,
	}

	linreg := &chart.LinearRegressionSeries{
		Style: chart.Style{
			Show:            true,
			StrokeColor:     chart.ColorAlternateBlue,
			StrokeDashArray: []float64{5.0, 5.0},
			FontSize:        8,
		},
		InnerSeries: mainSeries,
	}

	graph := chart.Chart{
		Width:  width,
		Height: height,
		YAxis: chart.YAxis{
			Style: chart.Style{
				Show:     false,
				FontSize: 8,
			},
		},
		XAxis: chart.XAxis{
			Style: chart.Style{
				Show:     false,
				FontSize: 8,
			},
			ValueFormatter: chart.TimeValueFormatter,
		},
		Series: []chart.Series{
			mainSeries,
			linreg,
		},
	}

	rc.Response.Header().Set("Content-Type", "image/svg+xml")
	graph.Render(chart.SVG, rc.Response)
	return nil
}

// Register registers the controller.
func (c Chart) Register(app *web.App) {
	app.GET("/chart/searches", c.getSearchChartAction)
}
