package controller

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	// heroku runtime metrics
	_ "github.com/heroku/x/hmetrics/onload"

	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/web"
	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/viewmodel"
)

const (
	slackContentTypeJSON           = "application/json; charset=utf-8"
	slackContentTypeTextPlain      = "text/plain; charset=utf-8"
	slackErrorInvalidQuery         = "Please type at least (3) characters."
	slackErrorBadPayload           = "There was an error processing a payload from slack. Saddness."
	slackErrorInvalidAction        = "An invalid action was passed to the button handler."
	slackErrorInvalidCallbackState = "An invalid callback state was passed to the button handler."
	slackErrorInternal             = "There was an error processing your request. Sadness."
	slackErrorTeamDisabled         = "Your team has been disabled; contact the integration owner to re-enable."

	slackActionShuffle = "shuffle"
	slackActionPost    = "post"
	slackActionCancel  = "cancel"
)

// Integrations controller is responsible for integration responses.
type Integrations struct {
	Log    logger.Log
	Config *config.Giffy
	Model  *model.Manager
}

// Register registers the controller's actions with the app.
func (i Integrations) Register(app *web.App) {
	app.POST("/integrations/slack", i.slack)
	app.POST("/integrations/slack.action", i.slackAction)
	app.POST("/integrations/slack.event", i.slackEvent)
}

func (i Integrations) slack(rc *web.Ctx) web.Result {
	args := i.arguments(rc)

	if len(args.Query) < 3 {
		return web.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInvalidQuery))
	}

	result, errRes := i.getResult(args, rc)
	if errRes != nil {
		return errRes
	}

	res := slackMessage{}
	if strings.HasPrefix(args.Query, "img:") {
		res.DeleteOriginal = true
		res.AsUser = true
		res.ResponseType = "in_channel"
		res.AuthorName = args.UserName
		res.Attachments = []interface{}{
			slackImageAttachment{Title: args.Query, ImageURL: result.S3ReadURL},
		}
	} else {
		res.ReplaceOriginal = true
		res.ResponseType = "ephemeral"
		var title string
		if len(result.Tags) > 0 {
			title = result.Tags[0].TagValue
		} else {
			title = fmt.Sprintf("search: `%s`", args.Query)
		}

		res.Attachments = []interface{}{
			slackImageAttachment{Title: title, ImageURL: result.S3ReadURL},
			i.buttonActions(args.Query, result.UUID),
		}
	}

	return i.renderResult(res, rc)
}

func (i Integrations) slackAction(rc *web.Ctx) web.Result {
	var payload slackActionPayload
	body, err := rc.PostBodyAsString()
	if err != nil {
		logger.MaybeError(i.Log, err)
		return web.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorBadPayload))
	}
	bodyUnescaped, err := url.QueryUnescape(strings.TrimPrefix(body, "payload="))
	if err != nil {
		logger.MaybeError(i.Log, err)
		return web.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorBadPayload))
	}
	err = fromJSON([]byte(bodyUnescaped), &payload)
	if err != nil {
		logger.MaybeError(i.Log, err)
		return web.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorBadPayload))
	}

	switch payload.Action() {
	case slackActionShuffle:
		return i.slackShuffle(payload, rc)
	case slackActionPost:
		return i.slackPost(payload, rc)
	case slackActionCancel:
		return i.renderResult(slackMessage{DeleteOriginal: true}, rc)
	}
	return web.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInvalidAction))
}

func (i Integrations) slackErrorNoResults() string {
	return fmt.Sprintf("Giffy couldn't find what you were looking for; maybe add it here? %s/add_image", i.Config.Web.BaseURL)
}

func (i Integrations) slackShuffle(payload slackActionPayload, rc *web.Ctx) web.Result {
	query, uuid := i.parseCallbackID(payload.CallbackID)
	if query == "" || uuid == "" {
		return web.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInvalidCallbackState))
	}
	contentRatingFilter, err := i.getContentRatingForTeamID(rc.Context(), payload.Team.ID)
	if err != nil {
		logger.MaybeFatal(i.Log, err)
		return web.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInternal))
	}

	logger.MaybeInfof(i.Log, "search query: %s, excludes: %s", query, uuid)
	result, err := i.Model.SearchImagesBestResult(rc.Context(), query, []string{uuid}, contentRatingFilter)
	if err != nil {
		logger.MaybeFatal(i.Log, err)
		return web.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInternal))
	}
	if result == nil || result.IsZero() {
		return web.RawWithContentType(slackContentTypeTextPlain, []byte(i.slackErrorNoResults()))
	}
	output := viewmodel.NewImage(*result, i.Config)

	res := slackMessage{
		ReplaceOriginal: true,
		ResponseType:    "ephemeral",
	}
	var title string
	if len(result.Tags) > 0 {
		title = output.Tags[0].TagValue
	} else {
		title = output.DisplayName
	}

	res.Attachments = []interface{}{
		slackImageAttachment{Title: title, ImageURL: output.S3ReadURL},
		i.buttonActions(query, result.UUID),
	}
	return i.renderResult(res, rc)
}

func (i Integrations) slackPost(payload slackActionPayload, rc *web.Ctx) web.Result {
	var resultID *int64
	var tagID *int64

	if payload.Team.Name != "" {
		defer func() {
			logger.MaybeTrigger(rc.Context(),
				i.Log,
				model.NewSearchHistoryDetailed("slack", payload.Team.ID, payload.Team.Name, payload.Channel.ID, payload.Channel.Name, payload.User.ID, payload.User.Name, payload.OriginalMessage.Text, true, resultID, tagID),
			)
		}()
	}

	_, uuid := i.parseCallbackID(payload.CallbackID)
	if uuid == "" {
		return web.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInvalidCallbackState))
	}

	img, err := i.Model.GetImageByUUID(rc.Context(), uuid)
	if err != nil {
		logger.MaybeFatal(i.Log, err)
		return web.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInternal))
	}
	if img == nil || img.IsZero() {
		return web.RawWithContentType(slackContentTypeTextPlain, []byte(i.slackErrorNoResults()))
	}

	result := viewmodel.NewImage(*img, i.Config)

	res := slackMessage{
		DeleteOriginal: true,
		AsUser:         true,
		ResponseType:   "in_channel",
		AuthorName:     payload.User.Name,
	}
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
		return API(rc).BadRequest(err)
	}

	switch e.Type {
	case "url_verification":
		return web.RawWithContentType("application/json", []byte(toJSON(slackEventChallegeRepsonse{Challenge: e.Challenge})))
	}

	return web.NoContent
}

// --------------------------------------------------------------------------------
// Slack Helpers
// --------------------------------------------------------------------------------

func (i Integrations) getContentRatingForTeamID(ctx context.Context, teamID string) (contentRatingFilter int, err error) {
	contentRatingFilter = model.ContentRatingNR
	var slackTeam *model.SlackTeam
	slackTeam, err = i.Model.GetSlackTeamByTeamID(ctx, teamID)
	if err != nil {
		return
	}
	if slackTeam != nil && !slackTeam.IsZero() {
		contentRatingFilter = slackTeam.ContentRatingFilter
	}
	return contentRatingFilter, nil
}

func (i Integrations) arguments(rc *web.Ctx) slackArguments {
	return slackArguments{
		TeamID:      web.StringValue(rc.Param("team_id")),
		ChannelID:   web.StringValue(rc.Param("channel_id")),
		UserID:      web.StringValue(rc.Param("user_id")),
		TeamName:    web.StringValue(rc.Param("team_domain")),
		ChannelName: web.StringValue(rc.Param("channel_name")),
		UserName:    web.StringValue(rc.Param("user_name")),
		Query:       web.StringValue(rc.Param("text")),
	}
}

func (i Integrations) getResult(args slackArguments, rc *web.Ctx) (*viewmodel.Image, web.Result) {
	var result *model.Image
	var resultID *int64
	var tagID *int64
	var foundResult bool
	var err error

	if args.TeamName != "" {
		defer func() {
			logger.MaybeTrigger(rc.Context(), i.Log, model.NewSearchHistoryDetailed("slack", args.TeamID, args.TeamName, args.ChannelID, args.ChannelName, args.UserID, args.UserName, args.Query, foundResult, resultID, tagID))
		}()
	}

	contentRatingFilter, err := i.getContentRatingForTeamID(rc.Context(), args.TeamID)
	if err != nil {
		logger.MaybeFatal(i.Log, err)
		return nil, web.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInternal))
	}

	if strings.HasPrefix(args.Query, "img:") {
		uuid := strings.TrimPrefix(args.Query, "img:")
		result, err = i.Model.GetImageByUUID(rc.Context(), uuid)
	} else {
		result, err = i.Model.SearchImagesBestResult(rc.Context(), args.Query, nil, contentRatingFilter)
	}

	if err != nil {
		logger.MaybeFatal(i.Log, err)
		return nil, web.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInternal))
	}

	if result == nil || result.IsZero() {
		return nil, web.RawWithContentType(slackContentTypeTextPlain, []byte(i.slackErrorNoResults()))
	}

	foundResult = true
	resultID = &result.ID
	output := viewmodel.NewImage(*result, i.Config)
	return &output, nil
}

func (i Integrations) renderResult(res slackMessage, rc *web.Ctx) web.Result {
	responseBytes, err := json.Marshal(res)
	if err != nil {
		logger.MaybeFatal(i.Log, err)
		return web.RawWithContentType(slackContentTypeTextPlain, []byte(slackErrorInternal))
	}

	return web.RawWithContentType(slackContentTypeJSON, responseBytes)
}

func (i Integrations) buttonActions(query, imageUUID string) slackActionAttachment {
	return slackActionAttachment{
		Text:           "Hit either `Post` or `Shuffle` (for a new image).",
		Fallback:       "Unable to do image things.",
		CallbackID:     i.createCallbackID(query, imageUUID),
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
				Style: "primary",
				Type:  "button",
				Value: slackActionPost,
			},
			{
				Name:  "action",
				Text:  "Cancel",
				Type:  "button",
				Value: slackActionCancel,
			},
		},
	}
}

func (i Integrations) createCallbackID(query, uuid string) string {
	return fmt.Sprintf("%s||%s", base64.StdEncoding.EncodeToString([]byte(query)), uuid)
}

func (i Integrations) parseCallbackID(callbackState string) (query, uuid string) {
	if len(callbackState) == 0 {
		return
	}
	parts := strings.SplitN(callbackState, "||", 2)
	if len(parts) < 2 {
		return
	}

	decoded, _ := base64.StdEncoding.DecodeString(parts[0])
	query = string(decoded)
	uuid = parts[1]
	return
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
	AuthorName      string        `json:"author_name"`
	AuthorLink      string        `json:"author_link"`
	ResponseType    string        `json:"response_type"`
	ReplaceOriginal bool          `json:"replace_original"`
	DeleteOriginal  bool          `json:"delete_original"`
	Text            string        `json:"text,omitempty"`
	AsUser          bool          `json:"as_user"`
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
	Style string `json:"style"`
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
