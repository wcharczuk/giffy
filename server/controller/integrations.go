package controller

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/go-web"
)

const (
	slackContenttypeJSON      = "application/json; charset=utf-8"
	slackContentTypeTextPlain = "text/plain; charset=utf-8"
	slackErrorInvalidQuery    = "Please type at least (3) characters."
	slackErrorInternal        = "There was an error processing your request. Sadness."
)

var (
	slackErrorNoResults = fmt.Sprintf("Giffy couldn't find what you were looking for; maybe add it here? %s/#/add_image", core.ConfigURL())
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
		return rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInvalidQuery))
	}

	var result *model.Image
	var resultID *int64
	var tagID *int64
	var foundResult bool
	var err error

	defer func() {
		rc.Diagnostics().OnEvent(core.EventFlagSearch, model.NewSearchHistoryDetailed("slack", teamID, teamName, channelID, channelName, userID, userName, query, foundResult, resultID, tagID))
	}()

	if strings.HasPrefix(query, "img:") {
		uuid := strings.TrimPrefix(query, "img:")
		result, err = model.GetImageByUUID(uuid, rc.Tx())
	} else {
		result, err = model.SearchImagesBestResult(query, model.ContentRatingNR, rc.Tx())
	}
	if err != nil {
		rc.Diagnostics().Error(err)
		return rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInternal))
	}

	if result == nil || result.IsZero() {
		return rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorNoResults))
	}

	foundResult = true
	resultID = util.OptionalInt64(result.ID)

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
		if len(result.Tags) > 0 {
			res.Attachments = []interface{}{
				slackImageAttachment{Title: result.Tags[0].TagValue, ImageURL: result.S3ReadURL},
			}
		} else {
			res.Attachments = []interface{}{
				slackImageAttachment{Title: query, ImageURL: result.S3ReadURL},
			}
		}
	}

	responseBytes, err := json.Marshal(res)
	if err != nil {
		rc.Diagnostics().Error(err)
		return rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInternal))
	}

	return rc.RawWithContentType(slackContenttypeJSON, responseBytes)
}

func (i Integrations) slackSearchAction(rc *web.RequestContext) web.ControllerResult {
	query := rc.Param("text")
	if len(query) < 3 {
		return rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInvalidQuery))
	}

	results, err := model.SearchImagesWeightedRandom(query, model.ContentRatingFilterDefault, 3, rc.Tx())
	if err != nil {
		rc.Diagnostics().Error(err)
		return rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInternal))
	}

	if len(results) == 0 {
		return rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorNoResults))
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
		return rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInternal))
	}
	return rc.RawWithContentType(slackContenttypeJSON, responseBytes)
}

// Register registers the controller's actions with the app.
func (i Integrations) Register(app *web.App) {
	app.GET("/integrations/slack", i.slackAction)
	app.POST("/integrations/slack", i.slackAction)

	app.GET("/integrations/slack.search", i.slackSearchAction)
	app.POST("/integrations/slack.search", i.slackSearchAction)
}
