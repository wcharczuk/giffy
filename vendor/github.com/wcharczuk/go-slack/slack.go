// Copyright 2016 Will Charczuk. Use of this source code is governed by a MIT-style license that can be found
// in the LICENSE file.

// Package slack is a event driven client for the popular Slack chat application.

// A trivial example is:
//  package main
//  import (
//      "fmt"
//      "os"
//      "github.com/wcharczuk/go-slack"
//  )
//
//  func main() {
//      client := slack.NewClient(os.Getenv("SLACK_TOKEN"))
//      client.Listen(slack.EventHello, func(m *slack.Message, c *slack.Client) {
//          fmt.Println("connected")
//      })
//      client.Listen(slack.EventMessage, func(m *slack.Message, c *slack.Client) {
//          fmt.Prinln("message received!")
//      })
//      session, err := client.Connect() //session has the current users list and channel list
//      if err != nil {
//          fmt.Printf("%v\n", err)
//          os.Exit(1)
//      }
//  }
// The client has two phases of initialization, NewClient and Start.

package slack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/blendlabs/go-exception"
	"github.com/gorilla/websocket"
)

// Event is a type alias for string to differentiate Slack event types.
type Event string

const (
	// APIScheme is the protocol used to communicate with slack.
	APIScheme = "https"
	// APIEndpoint is the host used to communicate with slack.
	APIEndpoint = "slack.com"

	// ErrorNotAuthed : No authentication token provided.
	ErrorNotAuthed = "not_authed"
	// ErrorInvalidAuth : Invalid authentication token
	ErrorInvalidAuth = "invalid_auth"
	// ErrorAccountInactive : Authentication token is for a deleted user or team.
	ErrorAccountInactive = "account_inactive"
	// ErrorInvalidArrayArg : The method was passed a PHP-style array argument (e.g. with a name like foo[7]). These are never valid with the Slack API.
	ErrorInvalidArrayArg = "invalid_array_arg"
	// ErrorInvalidCharset : The method was called via a POST request, but the charset specified in the Content-Type header was invalid. Valid charset names are: utf-8 or iso-8859-1.
	ErrorInvalidCharset = "invalid_charset"
	// ErrorInvalidFormData : The method was called via a POST request with Content-Type application/x-www-form-urlencoded or multipart/form-data, but the form data was either missing or syntactically invalid.
	ErrorInvalidFormData = "invalid_form_data"
	// ErrorInvalidPostType : The method was called via a POST request, but the specified Content-Type was invalid. Valid types are: application/json application/x-www-form-urlencoded multipart/form-data text/plain.
	ErrorInvalidPostType = "invalid_post_type"
	// ErrorMissingPostType : The method was called via a POST request and included a data payload, but the request did not include a Content-Type header.
	ErrorMissingPostType = "missing_post_type"
	// ErrorRequestTimeout : The method was called via a POST request, but the POST data was either missing or truncated.
	ErrorRequestTimeout = "request_timeout"
	// ErrorMessageNotFound No message exists with the requested timestamp.
	ErrorMessageNotFound = "message_not_found"
	// ErrorChannelNotFound : Value passed for channel was invalid.
	ErrorChannelNotFound = "channel_not_found"
	// ErrorCantDeleteMessage : Authenticated user does not have permission to delete this message.
	ErrorCantDeleteMessage = "cant_delete_message"
	// ErrorUserIsBot : Authenticated user is a bot and is restricted in it's use of certain api endpoints.
	ErrorUserIsBot = "user_is_bot"
	// ErrorBadTimestamp : Value passed for timestamp was invalid.
	ErrorBadTimestamp = "bad_timestamp"
	// ErrorFileNotFound : File specified by file does not exist.
	ErrorFileNotFound = "file_not_found"
	// ErrorFileCommentNotFound : File comment specified by file_comment does not exist.
	ErrorFileCommentNotFound = "file_comment_not_found"
	// ErrorNoItemSpecified : file, file_comment, or combination of channel and timestamp was not specified.
	ErrorNoItemSpecified = "no_item_specified"
	// ErrorInvalidName : Value passed for name was invalid.
	ErrorInvalidName = "invalid_name"
	// ErrorAlreadyReacted : The specified item already has the user/reaction combination.
	ErrorAlreadyReacted = "already_reacted"
	// ErrorTooManyEmoji : The limit for distinct reactions (i.e emoji) on the item has been reached.
	ErrorTooManyEmoji = "too_many_emoji"
	// ErrorTooManyReactions : 	The limit for reactions a person may add to the item has been reached.
	ErrorTooManyReactions = "too_many_reactions"

	// EventHello is an enumerated event.
	EventHello Event = "hello"
	// EventMessage is an enumerated event.
	EventMessage Event = "message"
	// EventUserTyping is an enumerated event.
	EventUserTyping Event = "user_typing"
	// EventChannelMarked is an enumerated event.
	EventChannelMarked Event = "channel_marked"
	// EventChannelJoined is an enumerated event.
	EventChannelJoined Event = "channel_joined"
	// EventChannelLeft is an enumerated event.
	EventChannelLeft Event = "channel_left"
	// EventChannelDeleted is an enumerated event.
	EventChannelDeleted Event = "channel_deleted"
	// EventChannelRename is an enumerated event.
	EventChannelRename Event = "channel_rename"
	// EventChannelArchive is an enumerated event.
	EventChannelArchive Event = "channel_archive"
	// EventChannelUnArchive is an enumerated event.
	EventChannelUnArchive Event = "channel_unarchive"
	// EventChannelHistoryChanged is an enumerated event.
	EventChannelHistoryChanged Event = "channel_history_changed"
	// EventDNDUpdated is an enumerated event.
	EventDNDUpdated Event = "dnd_updated"
	// EventDNDUpdatedUser is an enumerated event.
	EventDNDUpdatedUser Event = "dnd_updated_user"
	// EventIMCreated is an enumerated event.
	EventIMCreated Event = "im_created"
	// EventImOpen is an enumerated event.
	EventImOpen Event = "im_open"
	// EventImClose is an enumerated event.
	EventImClose Event = "im_close"
	// EventImMarked is an enumerated event.
	EventImMarked Event = "im_marked"
	// EventImHistoryChanged is an enumerated event.
	EventImHistoryChanged Event = "im_history_changed"
	// EventGroupJoined is an enumerated event.
	EventGroupJoined Event = "group_joined"
	// EventGroupLeft is an enumerated event.
	EventGroupLeft Event = "group_left"
	// EventGroupOpen is an enumerated event.
	EventGroupOpen Event = "group_open"
	// EventGroupClose is an enumerated event.
	EventGroupClose Event = "group_close"
	// EventGroupArchive is an enumerated event.
	EventGroupArchive Event = "group_archive"
	// EventGroupUnarchive is an enumerated event.
	EventGroupUnarchive Event = "group_unarchive"
	// EventGroupRename is an enumerated event.
	EventGroupRename Event = "group_rename"
	// EventGroupMarked is an enumerated event.
	EventGroupMarked Event = "group_marked"
	// EventGroupHistoryChanged is an enumerated event.
	EventGroupHistoryChanged Event = "group_history_changed"
	// EventFileCreated is an enumerated event.
	EventFileCreated Event = "file_created"
	// EventFileShared is an enumerated event.
	EventFileShared Event = "file_shared"
	// EventFileUnshared is an enumerated event.
	EventFileUnshared Event = "file_unshared"
	// EventFilePublic is an enumerated event.
	EventFilePublic Event = "file_public"
	// EventFilePrivate is an enumerated event.
	EventFilePrivate Event = "file_private"
	// EventFileChange is an enumerated event.
	EventFileChange Event = "file_change"
	// EventFileDeleted is an enumerated event.
	EventFileDeleted Event = "file_deleted"
	// EventFileCommentAdded is an enumerated event.
	EventFileCommentAdded Event = "file_comment_added"
	// EventFileCommentEdited is an enumerated event.
	EventFileCommentEdited Event = "file_comment_edited"
	// EventFileCommentDeleted is an enumerated event.
	EventFileCommentDeleted Event = "file_comment_deleted"
	// EventPinAdded is an enumerated event.
	EventPinAdded Event = "pin_added"
	// EventPinRemoved is an enumerated event.
	EventPinRemoved Event = "pin_removed"
	// EventPresenceChange is an enumerated event.
	EventPresenceChange Event = "presence_change"
	// EventManualPresenceChange is an enumerated event.
	EventManualPresenceChange Event = "manual_presence_change"
	// EventPrefChange is an enumerated event.
	EventPrefChange Event = "pref_change"
	// EventUserChange is an enumerated event.
	EventUserChange Event = "user_change"
	// EventTeamJoin is an enumerated event.
	EventTeamJoin Event = "team_join"
	// EventStarAdded is an enumerated event.
	EventStarAdded Event = "star_added"
	// EventStarRemoved is an enumerated event.
	EventStarRemoved Event = "star_removed"
	// EventReactionAdded is an enumerated event.
	EventReactionAdded Event = "reaction_added"
	// EventReactionRemoved is an enumerated event.
	EventReactionRemoved Event = "reaction_removed"
	// EventEmojiChanged is an enumerated event.
	EventEmojiChanged Event = "emoji_changed"
	// EventCommandsChanged is an enumerated event.
	EventCommandsChanged Event = "commands_changed"
	// EventTeamPlanChanged is an enumerated event.
	EventTeamPlanChanged Event = "team_plan_changed"
	// EventTeamPrefChanged is an enumerated event.
	EventTeamPrefChanged Event = "team_pref_changed"
	// EventEmailDomainChanged is an enumerated event.
	EventEmailDomainChanged Event = "email_domain_changed"
	// EventTeamProfileChange is an enumerated event.
	EventTeamProfileChange Event = "team_profile_change"
	// EventTeamProfileDelete is an enumerated event.
	EventTeamProfileDelete Event = "team_profile_delete"
	// EventTeamProfileReorder is an enumerated event.
	EventTeamProfileReorder Event = "team_profile_reorder"
	// EventBotAdded is an enumerated event.
	EventBotAdded Event = "bot_added"
	// EventBotChanged is an enumerated event.
	EventBotChanged Event = "bot_changed"
	// EventAccountsChanged is an enumerated event.
	EventAccountsChanged Event = "accounts_changed"
	// EventTeamMigrationStarted is an enumerated event.
	EventTeamMigrationStarted Event = "team_migration_started"

	// EventSubtypeBotMessage is an enumerated sub event.
	EventSubtypeBotMessage Event = "bot_message"
	// EventSubtypeMeMessage is an enumerated sub event.
	EventSubtypeMeMessage Event = "me_message"
	// EventSubtypeMessageChanged is an enumerated sub event.
	EventSubtypeMessageChanged Event = "message_changed"
	// EventSubtypeMessageDeleted is an enumerated sub event.
	EventSubtypeMessageDeleted Event = "message_deleted"
	// EventSubtypeChannelJoin is an enumerated sub event.
	EventSubtypeChannelJoin Event = "channel_join"
	// EventSubtypeChannelLeave is an enumerated sub event.
	EventSubtypeChannelLeave Event = "channel_leave"
	// EventSubtypeChannelTopic is an enumerated sub event.
	EventSubtypeChannelTopic Event = "channel_topic"
	// EventSubtypeChannelPurpose is an enumerated sub event.
	EventSubtypeChannelPurpose Event = "channel_purpose"
	// EventSubtypeChannelName is an enumerated sub event.
	EventSubtypeChannelName Event = "channel_name"
	// EventSubtypeChannelArchive is an enumerated sub event.
	EventSubtypeChannelArchive Event = "channel_archive"
	// EventSubtypeChannelUnarchive is an enumerated sub event.
	EventSubtypeChannelUnarchive Event = "channel_unarchive"
)

// Listener is a function that recieves messages from a client.
type Listener func(message *Message, client *Client)

// --------------------------------------------------------------------------------
// Connect(token string) is the main slack entry point.
// It returns a `*Client`, which is the root struct for interacting with slack.
// Register "Listeners" to hook into incoming messages and filter out ones you
// dont' care about.
// --------------------------------------------------------------------------------

// NewClient creates a Client with a given token.
func NewClient(token string) *Client {
	c := &Client{Token: token, EventListeners: map[Event][]Listener{}, ActiveChannels: []string{}, activeLock: &sync.Mutex{}}
	c.Listen(EventChannelJoined, c.handleChannelJoined)
	c.Listen(EventChannelDeleted, c.handleChannelDeleted)
	c.Listen(EventChannelUnArchive, c.handleChannelUnarchive)
	c.Listen(EventChannelLeft, c.handleChannelLeft)
	return c
}

// Client is the mechanism with which the package consumer interacts with Slack.
type Client struct {
	Token          string
	EventListeners map[Event][]Listener
	ActiveChannels []string

	activeLock       *sync.Mutex
	socketConnection *websocket.Conn
}

// Listen attaches a new Listener to the given event.
// There can be multiple listeners to an event.
// If an event is already being listened for, calling Listen will add a new listener to that event.
func (rtm *Client) Listen(event Event, handler Listener) {
	if listeners, handlesEvent := rtm.EventListeners[event]; handlesEvent {
		rtm.EventListeners[event] = append(listeners, handler)
	} else {
		rtm.EventListeners[event] = []Listener{handler}
	}
}

// StopListening removes all listeners for an event.
func (rtm *Client) StopListening(event Event) {
	delete(rtm.EventListeners, event)
}

// Connect begins a session with Slack.
func (rtm *Client) Connect() (*Session, error) {
	res := Session{}
	resErr := NewExternalRequest().
		AsPost().
		WithScheme(APIScheme).
		WithHost(APIEndpoint).
		WithPath("api/rtm.start").
		WithPostData("token", rtm.Token).
		WithPostData("no_unreads", "true").
		WithPostData("mpim_aware", "true").
		FetchJSONToObject(&res)

	if resErr != nil {
		return nil, resErr
	}

	if !IsEmpty(res.Error) {
		return nil, exception.New(res.Error)
	}

	//start socket connection
	u, urlErr := url.Parse(res.URL)
	if urlErr != nil {
		return nil, urlErr
	}

	var socketErr error
	rtm.socketConnection, _, socketErr = websocket.DefaultDialer.Dial(u.String(), nil)

	if socketErr != nil {
		return nil, socketErr
	}

	go func() { //fetch the (initial) channel list asyncronously
		rtm.activeLock.Lock()
		defer rtm.activeLock.Unlock()

		channels, chanelsErr := rtm.ChannelsList(true) //excludeArchived == true
		if chanelsErr != nil {
			return
		}

		for x := 0; x < len(channels); x++ {
			channel := channels[x]
			if channel.IsMember && !channel.IsArchived {
				rtm.ActiveChannels = append(rtm.ActiveChannels, channel.ID)
			}
		}
	}()

	go func() {
		rtm.listenLoop()
	}()

	return &res, nil
}

// Stop closes the connection with Slack.
func (rtm *Client) Stop() error {
	if rtm.socketConnection == nil {
		return nil
	}

	closeErr := rtm.socketConnection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if closeErr != nil {
		return closeErr
	}
	rtm.socketConnection.Close()
	return nil
}

// SendMessage sends a basic message over the open web socket connection to slack.
func (rtm Client) SendMessage(m *Message) error {
	if rtm.socketConnection == nil {
		return exception.New("Connection is closed.")
	}

	return rtm.socketConnection.WriteJSON(m)
}

// Say sends a basic message to a given channelID.
func (rtm Client) Say(channelID string, messageComponents ...interface{}) error {
	if rtm.socketConnection == nil {
		return exception.New("Connection is closed.")
	}

	m := &Message{Type: "message", Text: fmt.Sprint(messageComponents...), Channel: channelID}
	return rtm.SendMessage(m)
}

// Sayf is an overload that uses Printf style replacements for a basic message to a given channelID.
func (rtm Client) Sayf(channelID, format string, messageComponents ...interface{}) error {
	if rtm.socketConnection == nil {
		return exception.New("Connection is closed.")
	}

	m := &Message{Type: "message", Text: fmt.Sprintf(format, messageComponents...), Channel: channelID}
	return rtm.SendMessage(m)
}

// Ping sends a special type of "ping" message to Slack to remind it to keep the connection open.
// Currently unused internally by Slack.
func (rtm *Client) Ping() error {
	if rtm.socketConnection == nil {
		return exception.New("Connection is closed.")
	}

	m := &Message{ID: UUIDv4().ToShortString(), Type: "ping"}
	return rtm.SendMessage(m)
}

//--------------------------------------------------------------------------------
// INTERNAL METHODS
//--------------------------------------------------------------------------------

func (rtm *Client) listenLoop() error {
	for {
		if rtm.socketConnection == nil {
			return nil
		}
		_, messageBytes, err := rtm.socketConnection.ReadMessage()
		if err != nil {
			return err
		}

		var bm BareMessage
		jsonErr := json.Unmarshal(messageBytes, &bm)
		if bm.Type == EventChannelJoined {
			var cm ChannelJoinedMessage
			jsonErr = json.Unmarshal(messageBytes, &cm)
			if jsonErr == nil {
				rtm.dispatch(&Message{Type: EventChannelJoined, Channel: cm.Channel.ID})
			}
		} else {
			var m Message
			jsonErr = json.Unmarshal(messageBytes, &m)
			if jsonErr == nil {
				rtm.dispatch(&m)
			}
		}
	}
}

func (rtm *Client) dispatch(m *Message) {
	if listeners, hasListeners := rtm.EventListeners[m.Type]; hasListeners {
		for x := 0; x < len(listeners); x++ {
			listener := listeners[x]
			go func() {
				listener(m, rtm)
			}()
		}
	}
}

func (rtm *Client) handleChannelJoined(message *Message, client *Client) {
	rtm.activeLock.Lock()
	defer rtm.activeLock.Unlock()
	rtm.ActiveChannels = append(rtm.ActiveChannels, message.Channel)
}

func (rtm *Client) handleChannelUnarchive(message *Message, client *Client) {
	rtm.activeLock.Lock()
	defer rtm.activeLock.Unlock()

	channel, channelErr := rtm.ChannelsInfo(message.Channel)
	if channelErr != nil {
		return
	}
	if channel.IsMember {
		rtm.ActiveChannels = append(rtm.ActiveChannels, message.Channel)
	}
}

func (rtm *Client) handleChannelLeft(message *Message, client *Client) {
	rtm.removeActiveChannel(message.Channel)
}

func (rtm *Client) handleChannelDeleted(message *Message, client *Client) {
	rtm.removeActiveChannel(message.Channel)
}

func (rtm *Client) removeActiveChannel(channelID string) {
	rtm.activeLock.Lock()
	defer rtm.activeLock.Unlock()

	currentChannels := []string{}
	for x := 0; x < len(rtm.ActiveChannels); x++ {
		currentChannelID := rtm.ActiveChannels[x]
		if channelID != currentChannelID {
			currentChannels = append(currentChannels, currentChannelID)
		}
	}
	rtm.ActiveChannels = currentChannels
}

//--------------------------------------------------------------------------------
// API METHODS
//--------------------------------------------------------------------------------

// AuthTest tests if the token works for a client.
func (rtm *Client) AuthTest() (*AuthTestResponse, error) {
	res := AuthTestResponse{}
	resErr := NewExternalRequest().
		AsPost().
		WithScheme(APIScheme).
		WithHost(APIEndpoint).
		WithPath("api/auth.test").
		WithPostData("token", rtm.Token).
		FetchJSONToObject(&res)

	if resErr != nil {
		return nil, resErr
	}

	if len(res.Error) != 0 {
		return nil, exception.New(res.Error)
	}

	if !res.OK {
		return nil, exception.New("slack response `ok` is false.")
	}

	return &res, nil
}

// ChannelsHistory returns the messages in a channel.
func (rtm *Client) ChannelsHistory(channelID string, latest, oldest *time.Time, count int, unreads bool) (*ChannelsHistoryResponse, error) {
	unreadsValue := "0"
	if unreads {
		unreadsValue = "1"
	}

	if count == -1 {
		count = 1000
	} else if count < 1 {
		count = 1
	} else if count > 1000 {
		count = 1000
	}

	res := ChannelsHistoryResponse{}
	req := NewExternalRequest().
		AsPost().
		WithScheme(APIScheme).
		WithHost(APIEndpoint).
		WithPath("api/channels.history").
		WithPostData("token", rtm.Token).
		WithPostData("channel", channelID).
		WithPostData("count", string(count)).
		WithPostData("unreads", unreadsValue)

	if latest != nil {
		req = req.WithPostData("latest", Timestamp{time: *latest}.String())
	}

	if oldest != nil {
		req = req.WithPostData("oldest", Timestamp{time: *latest}.String())
	}

	resErr := req.FetchJSONToObject(&res)
	if resErr != nil {
		return nil, resErr
	}

	if len(res.Error) != 0 {
		return nil, exception.New(res.Error)
	}

	if !res.OK {
		return nil, exception.New("slack response `ok` is false.")
	}

	return &res, nil
}

// ChannelsInfo returns information about a given channelID.
func (rtm *Client) ChannelsInfo(channelID string) (*Channel, error) {
	res := channelsInfoResponse{}
	resErr := NewExternalRequest().
		AsPost().
		WithScheme(APIScheme).
		WithHost(APIEndpoint).
		WithPath("api/channels.info").
		WithPostData("token", rtm.Token).
		WithPostData("channel", channelID).
		FetchJSONToObject(&res)

	if resErr != nil {
		return nil, resErr
	}

	if !IsEmpty(res.Error) {
		return nil, exception.New(res.Error)
	}

	if !res.OK {
		return nil, exception.New("slack response `ok` is false.")
	}

	return res.Channel, nil
}

// ChannelsList returns the list of channels available to the bot.
func (rtm *Client) ChannelsList(excludeArchived bool) ([]Channel, error) {
	res := channelsListResponse{}
	req := NewExternalRequest().
		AsPost().
		WithScheme(APIScheme).
		WithHost(APIEndpoint).
		WithPath("api/channels.list").
		WithPostData("token", rtm.Token)

	if excludeArchived {
		req = req.WithPostData("exclude_archived", "1")
	}

	resErr := req.FetchJSONToObject(&res)

	if resErr != nil {
		return nil, resErr
	}

	if !IsEmpty(res.Error) {
		return nil, exception.New(res.Error)
	}

	if !res.OK {
		return nil, exception.New("slack response `ok` is false.")
	}

	return res.Channels, nil
}

// ChannelsMark marks a message.
func (rtm *Client) ChannelsMark(channelID string, ts Timestamp) error {
	res := basicResponse{}
	resErr := NewExternalRequest().
		AsPost().
		WithScheme(APIScheme).
		WithHost(APIEndpoint).
		WithPath("api/chat.mark").
		WithPostData("token", rtm.Token).
		WithPostData("channel", channelID).
		WithPostData("ts", ts.String()).
		FetchJSONToObject(&res)

	if resErr != nil {
		return resErr
	}

	if !IsEmpty(res.Error) {
		return exception.New(res.Error)
	}
	if !res.OK {
		return exception.New("slack response `ok` is false.")
	}

	return nil
}

// ChannelsSetPurpose sets the purpose for a given Slack channel.
func (rtm *Client) ChannelsSetPurpose(channelID, purpose string) error {
	res := basicResponse{}
	resErr := NewExternalRequest().
		AsPost().
		WithScheme(APIScheme).
		WithHost(APIEndpoint).
		WithPath("api/channels.setPurpose").
		WithPostData("token", rtm.Token).
		WithPostData("channel", channelID).
		WithPostData("purpose", purpose).
		FetchJSONToObject(&res)

	if resErr != nil {
		return resErr
	}

	if !IsEmpty(res.Error) {
		return exception.New(res.Error)
	}

	if !res.OK {
		return exception.New("slack response `ok` is false.")
	}

	return nil
}

// ChannelsSetTopic sets the topic for a given Slack channel.
func (rtm *Client) ChannelsSetTopic(channelID, topic string) error {
	res := basicResponse{}
	resErr := NewExternalRequest().
		AsPost().
		WithScheme(APIScheme).
		WithHost(APIEndpoint).
		WithPath("api/channels.setTopic").
		WithPostData("token", rtm.Token).
		WithPostData("channel", channelID).
		WithPostData("topic", topic).
		FetchJSONToObject(&res)

	if resErr != nil {
		return resErr
	}

	if !IsEmpty(res.Error) {
		return exception.New(res.Error)
	}

	if !res.OK {
		return exception.New("slack response `ok` is false.")
	}

	return nil
}

// ChatDelete deletes a message.
func (rtm *Client) ChatDelete(channelID string, ts Timestamp) error {
	res := basicResponse{}
	resErr := NewExternalRequest().
		AsPost().
		WithScheme(APIScheme).
		WithHost(APIEndpoint).
		WithPath("api/chat.delete").
		WithPostData("token", rtm.Token).
		WithPostData("channel", channelID).
		WithPostData("ts", ts.String()).
		FetchJSONToObject(&res)

	if resErr != nil {
		return resErr
	}

	if !IsEmpty(res.Error) {
		return exception.New(res.Error)
	}
	if !res.OK {
		return exception.New("slack response `ok` is false.")
	}

	return nil
}

// ChatPostMessage posts a message to Slack using the chat api.
func (rtm *Client) ChatPostMessage(m *ChatMessage) (*ChatMessageResponse, error) { //the response version of the message is returned for verification
	res := ChatMessageResponse{}
	resErr := NewExternalRequest().
		AsPost().
		WithScheme(APIScheme).
		WithHost(APIEndpoint).
		WithPath("api/chat.postMessage").
		WithPostData("token", rtm.Token).
		WithPostDataFromObject(m).
		FetchJSONToObject(&res)

	if resErr != nil {
		return nil, resErr
	}

	if !IsEmpty(res.Error) {
		return nil, exception.New(res.Error)
	}

	if !res.OK {
		return nil, exception.New("slack response `ok` is false.")
	}

	return &res, nil
}

// ChatUpdate updates a chat message.
func (rtm *Client) ChatUpdate(ts Timestamp, m *ChatMessage) (*ChatMessageResponse, error) { //the response version of the message is returned for verification
	res := ChatMessageResponse{}
	resErr := NewExternalRequest().
		AsPost().
		WithScheme(APIScheme).
		WithHost(APIEndpoint).
		WithPath("api/chat.update").
		WithPostData("token", rtm.Token).
		WithPostData("ts", ts.String()).
		WithPostDataFromObject(m).
		FetchJSONToObject(&res)

	if resErr != nil {
		return nil, resErr
	}

	if !IsEmpty(res.Error) {
		return nil, exception.New(res.Error)
	}

	if !res.OK {
		return nil, exception.New("slack response `ok` is false.")
	}

	return &res, nil
}

// EmojiList returns a list of current emoji's for a slack.
func (rtm *Client) EmojiList() (map[string]string, error) {
	res := emojiResponse{}
	resErr := NewExternalRequest().
		AsPost().
		WithScheme(APIScheme).
		WithHost(APIEndpoint).
		WithPath("api/emoji.list").
		WithPostData("token", rtm.Token).
		FetchJSONToObject(&res)

	if resErr != nil {
		return nil, resErr
	}

	if !IsEmpty(res.Error) {
		return nil, exception.New(res.Error)
	}

	if !res.OK {
		return nil, exception.New("slack response `ok` is false.")
	}
	return res.Emoji, nil
}

// ReactionsAdd adds a reaction.
func (rtm *Client) ReactionsAdd(name string, fileID, fileCommentID, channelID *string, ts *Timestamp) error {
	res := basicResponse{}
	req := NewExternalRequest().
		AsPost().
		WithScheme(APIScheme).
		WithHost(APIEndpoint).
		WithPath("api/reactions.add").
		WithPostData("token", rtm.Token).
		WithPostData("name", name)

	if fileID != nil {
		req = req.WithPostData("file", *fileID)
	} else if fileCommentID != nil {
		req = req.WithPostData("file_comment", *fileCommentID)
	} else if channelID != nil && ts != nil {
		req = req.WithPostData("channel", *channelID)
		req = req.WithPostData("timestamp", ts.String())
	} else {
		return exception.New("`fileId` or `fileCommentID` or (`channelID` and `ts`) must be not be nil.")
	}

	resErr := req.FetchJSONToObject(&res)

	if resErr != nil {
		return resErr
	}

	if !IsEmpty(res.Error) {
		return exception.New(res.Error)
	}

	if !res.OK {
		return exception.New("slack response `ok` is false.")
	}
	return nil
}

// ReactionsGet gets reactions.
func (rtm *Client) ReactionsGet(fileID, fileCommentID, channelID *string, ts *Timestamp) (*ChatMessageResponse, error) {
	res := ChatMessageResponse{}
	req := NewExternalRequest().
		AsPost().
		WithScheme(APIScheme).
		WithHost(APIEndpoint).
		WithPath("api/reactions.get").
		WithPostData("token", rtm.Token)

	if fileID != nil {
		req = req.WithPostData("file", *fileID)
	} else if fileCommentID != nil {
		req = req.WithPostData("file_comment", *fileCommentID)
	} else if channelID != nil && ts != nil {
		req = req.WithPostData("channel", *channelID)
		req = req.WithPostData("timestamp", ts.String())
	} else {
		return nil, exception.New("`fileId` or `fileCommentID` or (`channelID` and `ts`) must be not be nil.")
	}

	resErr := req.FetchJSONToObject(&res)

	if resErr != nil {
		return nil, resErr
	}

	if !IsEmpty(res.Error) {
		return nil, exception.New(res.Error)
	}

	if !res.OK {
		return nil, exception.New("slack response `ok` is false.")
	}
	return &res, nil
}

// ReactionsRemove removes a reaction.
func (rtm *Client) ReactionsRemove(name string, fileID, fileCommentID, channelID *string, ts *Timestamp) error {
	res := basicResponse{}
	req := NewExternalRequest().
		AsPost().
		WithScheme(APIScheme).
		WithHost(APIEndpoint).
		WithPath("api/reactions.remove").
		WithPostData("token", rtm.Token).
		WithPostData("name", name)

	if fileID != nil {
		req = req.WithPostData("file", *fileID)
	} else if fileCommentID != nil {
		req = req.WithPostData("file_comment", *fileCommentID)
	} else if channelID != nil && ts != nil {
		req = req.WithPostData("channel", *channelID)
		req = req.WithPostData("timestamp", ts.String())
	} else {
		return exception.New("`fileId` or `fileCommentID` or (`channelID` and `ts`) must be not be nil.")
	}

	resErr := req.FetchJSONToObject(&res)

	if resErr != nil {
		return resErr
	}

	if !IsEmpty(res.Error) {
		return exception.New(res.Error)
	}

	if !res.OK {
		return exception.New("slack response `ok` is false.")
	}
	return nil
}

// UsersList returns all users for a given Slack organization.
func (rtm *Client) UsersList() ([]User, error) {
	res := usersListResponse{}
	resErr := NewExternalRequest().
		AsPost().
		WithScheme(APIScheme).
		WithHost(APIEndpoint).
		WithPath("api/users.list").
		WithPostData("token", rtm.Token).
		FetchJSONToObject(&res)

	if resErr != nil {
		return nil, resErr
	}

	if !IsEmpty(res.Error) {
		return nil, exception.New(res.Error)
	}

	return res.Users, nil
}

// UsersInfo returns an User object for a given userID.
func (rtm *Client) UsersInfo(userID string) (*User, error) {
	res := usersInfoResponse{}
	resErr := NewExternalRequest().
		AsPost().
		WithScheme(APIScheme).
		WithHost(APIEndpoint).
		WithPath("api/users.info").
		WithPostData("token", rtm.Token).
		WithPostData("user", userID).
		FetchJSONToObject(&res)

	if resErr != nil {
		return nil, resErr
	}

	if !IsEmpty(res.Error) {
		return nil, exception.New(res.Error)
	}

	return res.User, nil
}
