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
	slackErrorBadPayload      = "There was an error processing a payload from slack. Saddness."
	slackErrorInvalidAction   = "An invalid action was passed to the button handler."
	slackErrorInternal        = "There was an error processing your request. Sadness."
	slackErrorTeamDisabled    = "Your team has been disabled; contact the integration owner to re-enable."

	slackActionShuffle = "shuffle"
	slackActionPost    = "post"
)

var (
	slackErrorNoResults = fmt.Sprintf("Giffy couldn't find what you were looking for; maybe add it here? %s/#/add_image", core.ConfigURL())
)

// Integrations controller is responsible for integration responses.
type Integrations struct{}

// Register registers the controller's actions with the app.
func (i Integrations) Register(app *web.App) {
	app.POST("/integrations/slack", i.slack)
	app.POST("/integrations/slack.action", i.slack)
	app.POST("/integrations/slack.event", i.slackEvent)
}

func (i Integrations) slack(rc *web.Ctx) web.Result {
	args := i.arguments(rc)

	if len(args.Query) < 3 {
		return rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInvalidQuery))
	}

	result, errRes := i.getResult(args, rc)
	if errRes != nil {
		return errRes
	}

	res := slackMessage{}
	if !strings.HasPrefix(args.Query, "img:") {
		res.ResponseType = "in_channel"
		res.Attachments = []interface{}{
			slackImageAttachment{Title: args.Query, ImageURL: result.S3ReadURL},
		}
	} else {
		res.ResponseType = "ephemeral"
		var title string
		if len(result.Tags) > 0 {
			title = result.Tags[0].TagValue
		} else {
			title = fmt.Sprintf("search: `%s`", args.Query)
		}

		res.Attachments = []interface{}{
			slackImageAttachment{Title: title, ImageURL: result.S3ReadURL},
			i.buttonActions(result.UUID),
		}
	}

	return i.renderResult(res, rc)
}

func (i Integrations) slackAction(rc *web.Ctx) web.Result {
	var payload slackActionPayload
	err := rc.PostBodyAsJSON(&payload)
	if err != nil {
		return rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorBadPayload))
	}

	switch payload.Action() {
	case slackActionShuffle:
		return i.slackShuffle(payload, rc)
	case slackActionPost:
		return i.slackPost(payload, rc)
	}
	return rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInvalidAction))
}

func (i Integrations) slackShuffle(payload slackActionPayload, rc *web.Ctx) web.Result {
	query := payload.OriginalMessage.Text
	contentRatingFilter, errRes := i.contentRatingFilter(payload.Team.ID, rc)
	if errRes != nil {
		return errRes
	}

	result, err := model.SearchImagesBestResult(query, contentRatingFilter, rc.Tx())
	if err != nil {
		return rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInternal))
	}

	res := slackMessage{}
	res.ReplaceOriginal = true
	res.ResponseType = "ephemeral"
	var title string
	if len(result.Tags) > 0 {
		title = result.Tags[0].TagValue
	} else {
		title = result.DisplayName
	}

	res.Attachments = []interface{}{
		slackImageAttachment{Title: title, ImageURL: result.S3ReadURL},
		i.buttonActions(result.UUID),
	}

	return i.renderResult(res, rc)
}

func (i Integrations) slackPost(payload slackActionPayload, rc *web.Ctx) web.Result {
	var result *model.Image
	var resultID *int64
	var tagID *int64

	defer func() {
		rc.Logger().OnEvent(core.EventFlagSearch, model.NewSearchHistoryDetailed("slack", payload.Team.ID, payload.Team.Name, payload.Channel.ID, payload.Channel.Name, payload.User.ID, payload.User.Name, payload.OriginalMessage.Text, true, resultID, tagID))
	}()
	uuid := payload.CallbackID

	result, err := model.GetImageByUUID(uuid, rc.Tx())
	if err != nil {
		return rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInternal))
	}

	res := slackMessage{}
	res.DeleteOriginal = true
	res.ResponseType = "in_channel"
	var title string
	if len(result.Tags) > 0 {
		title = result.Tags[0].TagValue
	} else {
		title = result.DisplayName
	}
	res.Attachments = []interface{}{
		slackImageAttachment{Title: title, ImageURL: result.S3ReadURL},
	}

	return i.renderResult(res, rc)
}

func (i Integrations) slackEvent(rc *web.Ctx) web.Result {
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

// --------------------------------------------------------------------------------
// Slack Helpers
// --------------------------------------------------------------------------------

func (i Integrations) contentRatingFilter(teamID string, rc *web.Ctx) (int, web.Result) {
	contentRatingFilter := model.ContentRatingNR
	teamSettings, err := model.GetSlackTeamByTeamID(teamID, rc.Tx())
	if err != nil {
		rc.Logger().FatalWithReq(err, rc.Request)
		return contentRatingFilter, rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInternal))
	} else if !teamSettings.IsZero() {
		if !teamSettings.IsEnabled {
			return contentRatingFilter, rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorTeamDisabled))
		}

		contentRatingFilter = teamSettings.ContentRatingFilter
	}
	return contentRatingFilter, nil
}

func (i Integrations) arguments(rc *web.Ctx) slackArguments {
	return slackArguments{
		TeamID:      rc.Param("team_id"),
		ChannelID:   rc.Param("channel_id"),
		UserID:      rc.Param("user_id"),
		TeamName:    rc.Param("team_domain"),
		ChannelName: rc.Param("channel_name"),
		UserName:    rc.Param("user_name"),
		Query:       rc.Param("text"),
	}
}

func (i Integrations) getResult(args slackArguments, rc *web.Ctx) (*model.Image, web.Result) {
	var result *model.Image
	var resultID *int64
	var tagID *int64
	var foundResult bool
	var err error

	defer func() {
		rc.Logger().OnEvent(core.EventFlagSearch, model.NewSearchHistoryDetailed("slack", args.TeamID, args.TeamName, args.ChannelID, args.ChannelName, args.UserID, args.UserName, args.Query, foundResult, resultID, tagID))
	}()

	contentRatingFilter, errRes := i.contentRatingFilter(args.TeamID, rc)
	if errRes != nil {
		return nil, errRes
	}

	if strings.HasPrefix(args.Query, "img:") {
		uuid := strings.TrimPrefix(args.Query, "img:")
		result, err = model.GetImageByUUID(uuid, rc.Tx())
	} else {
		result, err = model.SearchImagesBestResult(args.Query, contentRatingFilter, rc.Tx())
	}

	if err != nil {
		rc.Logger().FatalWithReq(err, rc.Request)
		return nil, rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInternal))
	}

	if result == nil || result.IsZero() {
		return nil, rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorNoResults))
	}

	foundResult = true
	resultID = util.OptionalInt64(result.ID)
	return result, nil
}

func (i Integrations) renderResult(res slackMessage, rc *web.Ctx) web.Result {
	responseBytes, err := json.Marshal(res)
	if err != nil {
		rc.Logger().FatalWithReq(err, rc.Request)
		return rc.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInternal))
	}

	return rc.RawWithContentType(slackContenttypeJSON, responseBytes)
}

func (i Integrations) buttonActions(imageUUID string) slackActionAttachment {
	return slackActionAttachment{
		Fallback:       "Unable to do image things.",
		CallbackID:     imageUUID,
		AttachmentType: "default",
		Actions: []slackAction{
			{
				Name:  "action",
				Text:  "Shuffle",
				Type:  "button",
				Value: slackActionShuffle,
			},
			{
				Name:  "action",
				Text:  "Post",
				Type:  "button",
				Value: slackActionPost,
			},
		},
	}
}

// --------------------------------------------------------------------------------
// Slack Types
// --------------------------------------------------------------------------------

type slackArguments struct {
	TeamID      string
	ChannelID   string
	UserID      string
	TeamName    string
	ChannelName string
	UserName    string
	Query       string
}

type slackActionPayload struct {
	Actions         []slackAction   `json:"actions"`
	CallbackID      string          `json:"callback_id"`
	Team            slackIdentifier `json:"team"`
	Channel         slackIdentifier `json:"channel"`
	User            slackIdentifier `json:"user"`
	ActionTS        string          `json:"action_ts"`
	MessageTS       string          `json:"message_ts"`
	Token           string          `json:"token"`
	OriginalMessage slackMessage    `json:"original_message"`
	ResponseURL     string          `json:"response_url"`
}

func (sap slackActionPayload) Action() (action string) {
	if len(sap.Actions) == 0 {
		return
	}
	action = sap.Actions[0].Value
	return
}

type slackIdentifier struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type slackMessage struct {
	ResponseType    string        `json:"response_type"`
	ReplaceOriginal bool          `json:"replace_original"`
	DeleteOriginal  bool          `json:"delete_original"`
	Text            string        `json:"text,omitempty"`
	Attachments     []interface{} `json:"attachments"`
}

type slackActionAttachment struct {
	Text           string        `json:"text"`
	Fallback       string        `json:"fallback"`
	CallbackID     string        `json:"callback_id"`
	Color          string        `json:"color"`
	AttachmentType string        `json:"attachment_type"`
	Actions        []slackAction `json:"actions"`
}

type slackAction struct {
	Name  string `json:"name"`
	Text  string `json:"text"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type slackImageAttachment struct {
	Title    string `json:"title"`
	ImageURL string `json:"image_url"`
	ThumbURL string `json:"thumb_url,omitempty"`
}

type slackMessageAttachment struct {
	Text   string       `json:"text"`
	Fields []slackField `json:"field"`
}

type slackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type slackEvent struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
	Token     string `json:"token"`
}

type slackEventChallegeRepsonse struct {
	Challenge string `json:"challenge"`
}
