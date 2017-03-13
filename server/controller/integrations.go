package controller

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blendlabs/go-util"
	"github.com/blendlabs/go-web"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/model"
)

const (
	slackContenttypeJSON      = "application/json; charset=utf-8"
	slackContentTypeTextPlain = "text/plain; charset=utf-8"
	slackErrorInvalidQuery    = "Please type at least (3) characters."
	slackErrorInternal        = "There was an error processing your request. Sadness."
	slackErrorTeamDisabled    = "Your team has been disabled, contact the integration owner to re-enable."
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

func (i Integrations) slackAction(rc *web.Ctx) web.Result {
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
		rc.Logger().OnEvent(core.EventFlagSearch, model.NewSearchHistoryDetailed("slack", teamID, teamName, channelID, channelName, userID, userName, query, foundResult, resultID, tagID))
	}()

	contentRatingFilter := model.ContentRatingNR
	teamSettings, err := model.GetSlackTeamByTeamID(teamID, rc.Tx())
	if err != nil {
		rc.Logger().Error(err)
	} else if !teamSettings.IsZero() {
		if !teamSettings.IsEnabled {
			return rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorTeamDisabled))
		}

		contentRatingFilter = teamSettings.ContentRatingFilter
	}

	if strings.HasPrefix(query, "img:") {
		uuid := strings.TrimPrefix(query, "img:")
		result, err = model.GetImageByUUID(uuid, rc.Tx())
	} else {
		result, err = model.SearchImagesBestResult(query, contentRatingFilter, rc.Tx())
	}
	if err != nil {
		rc.Logger().Error(err)
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
		rc.Logger().Error(err)
		return rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInternal))
	}

	return rc.RawWithContentType(slackContenttypeJSON, responseBytes)
}

func (i Integrations) slackSearchAction(rc *web.Ctx) web.Result {
	query := rc.Param("text")
	if len(query) < 3 {
		return rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInvalidQuery))
	}

	teamID := rc.Param("team_id")
	contentRatingFilter := model.ContentRatingFilterDefault
	teamSettings, err := model.GetSlackTeamByTeamID(teamID, rc.Tx())
	if err != nil {
		rc.Logger().Error(err)
	} else if !teamSettings.IsZero() {
		if !teamSettings.IsEnabled {
			return rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorTeamDisabled))
		}

		contentRatingFilter = teamSettings.ContentRatingFilter
	}

	results, err := model.SearchImagesWeightedRandom(query, contentRatingFilter, 3, rc.Tx())
	if err != nil {
		rc.Logger().Error(err)
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
		rc.Logger().Error(err)
		return rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInternal))
	}
	return rc.RawWithContentType(slackContenttypeJSON, responseBytes)
}

type slackEvent struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
	Token     string `json:"token"`
}

type slackEventChallegeRepsonse struct {
	Challenge string `json:"challenge"`
}

func (i Integrations) slackEventAction(rc *web.Ctx) web.Result {
	var e slackEvent
	err := rc.PostBodyAsJSON(&e)
	if err != nil {
		return rc.API().BadRequest(err.Error())
	}

	switch e.Type {
	case "url_verification":
		return rc.RawWithContentType("application/json", []byte(util.JSON.Serialize(slackEventChallegeRepsonse{Challenge: e.Challenge})))
	}

	return rc.NoContent()
}

// Register registers the controller's actions with the app.
func (i Integrations) Register(app *web.App) {
	app.GET("/integrations/slack", i.slackAction)
	app.POST("/integrations/slack", i.slackAction)

	app.GET("/integrations/slack.search", i.slackSearchAction)
	app.POST("/integrations/slack.search", i.slackSearchAction)

	app.GET("/integrations/slack.event", i.slackEventAction)
	app.POST("/integrations/slack.event", i.slackEventAction)
}
