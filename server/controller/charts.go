package controller

import (
	"bytes"
	"net/http"
	"time"

	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/web"
	"github.com/blend/go-sdk/webutil"
	"github.com/wcharczuk/go-chart"

	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/model"
)

// Chart is a controller for common chart endpoints.
type Chart struct {
	Log    logger.Log
	Config *config.Giffy
	Model  *model.Manager
}

func (c Chart) getSearchChartAction(rc *web.Ctx) web.Result {
	data, err := c.Model.GetSearchesPerDay(rc.Context(), time.Now().UTC().AddDate(0, -6, 0))
	if err != nil {
		return API(rc).InternalError(err)
	}

	var width, height int
	if widthParam, err := web.IntValue(rc.Param("width")); err == nil {
		width = widthParam
	} else {
		width = 1280
	}

	if heightParam, err := web.IntValue(rc.Param("height")); err == nil {
		height = heightParam
	} else {
		height = 256
	}

	xvalues, yvalues := model.DayCounts(data).ChartData()

	mainSeries := chart.TimeSeries{
		Name: "Search Count By Day",
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.ColorBlue,
			FontSize:    8,
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

	buf := new(bytes.Buffer)
	err = graph.Render(chart.SVG, buf)
	if err != nil {
		logger.MaybeFatal(c.Log, err)
		return nil
	}
	rc.Response.Header().Set(webutil.HeaderContentType, chart.ContentTypeSVG)
	rc.Response.WriteHeader(http.StatusOK)
	_, _ = buf.WriteTo(rc.Response)
	return nil
}

// Register registers the controller.
func (c Chart) Register(app *web.App) {
	app.GET("/chart/searches", c.getSearchChartAction)
}
