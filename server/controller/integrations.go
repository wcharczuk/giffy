package controller

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/core/external"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/go-web"
)

// Integrations controller is responsible for integration responses.
type Integrations struct{}

func (i Integrations) slackAction(r *web.RequestContext) web.ControllerResult {

	teamID := r.Param("team_id")
	channelID := r.Param("channel_id")
	userID := r.Param("user_id")

	teamName := r.Param("team_domain")
	channelName := r.Param("channel_name")
	userName := r.Param("user_name")

	query := r.Param("text")

	var result *model.Image
	var err error
	if strings.HasPrefix(query, "img:") {
		uuid := strings.Replace(query, "img:", "", -1)
		result, err = model.GetImageByUUID(uuid, nil)
	} else {
		result, err = model.SearchImagesSlack(query, nil)
	}
	if err != nil {
		return r.API().InternalError(err)
	}
	if result.IsZero() {
		model.QueueSearchHistoryEntry("slack", teamID, teamName, channelID, channelName, userID, userName, query, false, nil, nil)
		return r.RawWithContentType("text/plaid; charset=utf-8", []byte(fmt.Sprintf("Giffy couldn't find what you were looking for; maybe add it here? %s/#/add_image", core.ConfigURL())))
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
		return r.API().InternalError(err)
	}

	model.QueueSearchHistoryEntry("slack", teamID, teamName, channelID, channelName, userID, userName, query, true, util.OptionalInt64(result.ID), tagID)
	external.StatHatSearch()
	return r.RawWithContentType("application/json; charset=utf-8", responseBytes)
}

// Register registers the controller's actions with the app.
func (i Integrations) Register(app *web.App) {
	app.GET("/integrations/slack", i.slackAction)
	app.POST("/integrations/slack", i.slackAction)
}
