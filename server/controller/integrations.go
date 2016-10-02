package controller

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/external"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/go-web"
)

type slackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type slackMessageAttachment struct {
	Text   string       `json:"text"`
	Fields []slackField `json:"field"`
}

type slackImageAttachment struct {
	Title    string `json:"title"`
	ImageURL string `json:"image_url"`
	ThumbURL string `json:"thumb_url,omitempty"`
}

type slackResponse struct {
	ImageUUID    string        `json:"image_uuid"`
	ResponseType string        `json:"response_type"`
	Text         string        `json:"text,omitempty"`
	Attachments  []interface{} `json:"attachments"`
}

// Integrations controller is responsible for integration responses.
type Integrations struct{}

func (i Integrations) slackAction(rc *web.RequestContext) web.ControllerResult {
	teamID := rc.Param("team_id")
	channelID := rc.Param("channel_id")
	userID := rc.Param("user_id")

	teamName := rc.Param("team_domain")
	channelName := rc.Param("channel_name")
	userName := rc.Param("user_name")

	query := rc.Param("text")

	if len(query) < 3 {
		return rc.RawWithContentType("text/plain; charset=utf-8", []byte("Please type at least (3) characters."))
	}

	var result *model.Image
	var err error
	if strings.HasPrefix(query, "img:") {
		uuid := strings.Replace(query, "img:", "", -1)
		result, err = model.GetImageByUUID(uuid, rc.Tx())
	} else {
		result, err = model.SearchImagesBestResult(query, model.ContentRatingNR, rc.Tx())
	}
	if err != nil {
		model.QueueSearchHistoryEntry("slack", teamID, teamName, channelID, channelName, userID, userName, query, false, nil, nil)
		rc.Diagnostics().Error(err)
		return rc.RawWithContentType("text/plain; charset=utf-8", []byte("There was an error processing your request. Sadness."))
	}

	if result == nil || result.IsZero() {
		model.QueueSearchHistoryEntry("slack", teamID, teamName, channelID, channelName, userID, userName, query, false, nil, nil)
		return rc.RawWithContentType("text/plain; charset=utf-8", []byte(fmt.Sprintf("Giffy couldn't find what you were looking for; maybe add it here? %s/#/add_image", core.ConfigURL())))
	}

	var tagID *int64

	res := slackResponse{}
	res.ImageUUID = result.UUID
	res.ResponseType = "in_channel"

	if !strings.HasPrefix(query, "img:") {
		if len(result.Tags) > 0 {
			tagID = &result.Tags[0].ID
		}

		res.Attachments = []interface{}{
			slackImageAttachment{Title: query, ImageURL: result.S3ReadURL},
		}
	} else {
		res.Attachments = []interface{}{
			slackImageAttachment{Title: result.Tags[0].TagValue, ImageURL: result.S3ReadURL},
		}
	}

	responseBytes, err := json.Marshal(res)
	if err != nil {
		rc.Diagnostics().Error(err)
		return rc.RawWithContentType("text/plain; charset=utf-8", []byte("There was an error processing your request. Sadness."))
	}

	model.QueueSearchHistoryEntry("slack", teamID, teamName, channelID, channelName, userID, userName, query, true, util.OptionalInt64(result.ID), tagID)
	external.StatHatSearch()
	return rc.RawWithContentType("application/json; charset=utf-8", responseBytes)
}

func (i Integrations) slackSearchAction(rc *web.RequestContext) web.ControllerResult {
	teamID := rc.Param("team_id")
	channelID := rc.Param("channel_id")
	userID := rc.Param("user_id")

	teamName := rc.Param("team_domain")
	channelName := rc.Param("channel_name")
	userName := rc.Param("user_name")

	query := rc.Param("text")
	if len(query) < 3 {
		return rc.RawWithContentType("text/plain; charset=utf-8", []byte("Please type at least (3) characters."))
	}

	results, err := model.SearchImagesWeightedRandom(query, model.ContentRatingFilterDefault, 3, rc.Tx())
	if err != nil {
		rc.Diagnostics().Error(err)
		return rc.RawWithContentType("text/plain; charset=utf-8", []byte("There was an error processing your request. Sadness."))
	}

	if len(results) == 0 {
		model.QueueSearchHistoryEntry("slack", teamID, teamName, channelID, channelName, userID, userName, query, false, nil, nil)
		return rc.RawWithContentType("text/plain; charset=utf-8", []byte(fmt.Sprintf("Giffy couldn't find what you were looking for; maybe add it here? %s/#/add_image", core.ConfigURL())))
	}

	res := slackResponse{}
	res.ResponseType = "ephemeral"
	for _, i := range results {
		res.Attachments = append(res.Attachments,
			slackImageAttachment{Title: i.GetTagsSummary(), ImageURL: i.S3ReadURL},
		)
	}

	responseBytes, err := json.Marshal(res)
	if err != nil {
		rc.Diagnostics().Error(err)
		return rc.RawWithContentType("text/plain; charset=utf-8", []byte("There was an error processing your request. Sadness."))
	}
	return rc.RawWithContentType("application/json; charset=utf-8", responseBytes)
}

// Register registers the controller's actions with the app.
func (i Integrations) Register(app *web.App) {
	app.GET("/integrations/slack", i.slackAction)
	app.POST("/integrations/slack", i.slackAction)

	app.GET("/integrations/slack.search", i.slackSearchAction)
	app.POST("/integrations/slack.search", i.slackSearchAction)
}
