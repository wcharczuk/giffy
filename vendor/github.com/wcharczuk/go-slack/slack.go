package slack

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-request"
	"github.com/blendlabs/go-util"
	"github.com/gorilla/websocket"
)

type Event string

const (
	API_SCHEME   = "https"
	API_ENDPOINT = "slack.com"

	EVENT_HELLO                   Event = "hello"
	EVENT_MESSAGE                 Event = "message"
	EVENT_USER_TYPING             Event = "user_typing"
	EVENT_CHANNEL_MARKED          Event = "channel_marked"
	EVENT_CHANNEL_JOINED          Event = "channel_joined"
	EVENT_CHANNEL_LEFT            Event = "channel_left"
	EVENT_CHANNEL_DELETED         Event = "channel_deleted"
	EVENT_CHANNEL_RENAME          Event = "channel_rename"
	EVENT_CHANNEL_ARCHIVE         Event = "channel_archive"
	EVENT_CHANNEL_UNARCHIVE       Event = "channel_unarchive"
	EVENT_CHANNEL_HISTORY_CHANGED Event = "channel_history_changed"
	EVENT_DND_UPDATED             Event = "dnd_updated"
	EVENT_DND_UPDATED_USER        Event = "dnd_updated_user"
	EVENT_IM_CREATED              Event = "im_created"
	EVENT_IM_OPEN                 Event = "im_open"
	EVENT_IM_CLOSE                Event = "im_close"
	EVENT_IM_MARKED               Event = "im_marked"
	EVENT_IM_HISTORY_CHANGED      Event = "im_history_changed"
	EVENT_GROUP_JOINED            Event = "group_joined"
	EVENT_GROUP_LEFT              Event = "group_left"
	EVENT_GROUP_OPEN              Event = "group_open"
	EVENT_GROUP_CLOSE             Event = "group_close"
	EVENT_GROUP_ARCHIVE           Event = "group_archive"
	EVENT_GROUP_UNARCHIVE         Event = "group_unarchive"
	EVENT_GROUP_RENAME            Event = "group_rename"
	EVENT_GROUP_MARKED            Event = "group_marked"
	EVENT_GROUP_HISTORY_CHANGED   Event = "group_history_changed"
	EVENT_FILE_CREATED            Event = "file_created"
	EVENT_FILE_SHARED             Event = "file_shared"
	EVENT_FILE_UNSHARED           Event = "file_unshared"
	EVENT_FILE_PUBLIC             Event = "file_public"
	EVENT_FILE_PRIVATE            Event = "file_private"
	EVENT_FILE_CHANGE             Event = "file_change"
	EVENT_FILE_DELETED            Event = "file_deleted"
	EVENT_FILE_COMMENT_ADDED      Event = "file_comment_added"
	EVENT_FILE_COMMENT_EDITED     Event = "file_comment_edited"
	EVENT_FILE_COMMENT_DELETED    Event = "file_comment_deleted"
	EVENT_PIN_ADDED               Event = "pin_added"
	EVENT_PIN_REMOVED             Event = "pin_removed"
	EVENT_PRESENCE_CHANGE         Event = "presence_change"
	EVENT_MANUAL_PRESENCE_CHANGE  Event = "manual_presence_change"
	EVENT_PREF_CHANGE             Event = "pref_change"
	EVENT_USER_CHANGE             Event = "user_change"
	EVENT_TEAM_JOIN               Event = "team_join"
	EVENT_STAR_ADDED              Event = "star_added"
	EVENT_STAR_REMOVED            Event = "star_removed"
	EVENT_REACTION_ADDED          Event = "reaction_added"
	EVENT_REACTION_REMOVED        Event = "reaction_removed"
	EVENT_EMOJI_CHANGED           Event = "emoji_changed"
	EVENT_COMMANDS_CHANGED        Event = "commands_changed"
	EVENT_TEAM_PLAN_CHANGED       Event = "team_plan_changed"
	EVENT_TEAM_PREF_CHANGED       Event = "team_pref_changed"
	EVENT_EMAIL_DOMAIN_CHANGED    Event = "email_domain_changed"
	EVENT_TEAM_PROFILE_CHANGE     Event = "team_profile_change"
	EVENT_TEAM_PROFILE_DELETE     Event = "team_profile_delete"
	EVENT_TEAM_PROFILE_REORDER    Event = "team_profile_reorder"
	EVENT_BOT_ADDED               Event = "bot_added"
	EVENT_BOT_CHANGED             Event = "bot_changed"
	EVENT_ACCOUNTS_CHANGED        Event = "accounts_changed"
	EVENT_TEAM_MIGRATION_STARTED  Event = "team_migration_started"

	EVENT_SUBTYPE_BOT_MESSAGE       Event = "bot_message"
	EVENT_SUBTYPE_ME_MESSAGE        Event = "me_message"
	EVENT_SUBTYPE_MESSAGE_CHANGED   Event = "message_changed"
	EVENT_SUBTYPE_MESSAGE_DELETED   Event = "message_deleted"
	EVENT_SUBTYPE_CHANNEL_JOIN      Event = "channel_join"
	EVENT_SUBTYPE_CHANNEL_LEAVE     Event = "channel_leave"
	EVENT_SUBTYPE_CHANNEL_TOPIC     Event = "channel_topic"
	EVENT_SUBTYPE_CHANNEL_PURPOSE   Event = "channel_purpose"
	EVENT_SUBTYPE_CHANNEL_NAME      Event = "channel_name"
	EVENT_SUBTYPE_CHANNEL_ARCHIVE   Event = "channel_archive"
	EVENT_SUBTYPE_CHANNEL_UNARCHIVE Event = "channel_unarchive"
)

type User struct {
	Id                string       `json:"id"`
	Name              string       `json:"name"`
	Deleted           bool         `json:"deletd"`
	Color             string       `json:"color"`
	Profile           *UserProfile `json:"profile"`
	IsBot             bool         `json:"is_bot"`
	IsAdmin           bool         `json:"is_admin"`
	IsOwner           bool         `json:"is_owner"`
	IsPrimaryOwner    bool         `json:"is_primary_owner"`
	IsRestricted      bool         `json:"is_restricted"`
	IsUltraRestricted bool         `json:"is_ultra_restricted"`
	Has2FA            bool         `json:"has_2fa"`
	TwoFactorType     string       `json:"two_factor_type"`
	HasFiles          bool         `json:"has_files"`
}

type UserProfile struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	RealName  string `json:"real_name"`
	Email     string `json:"email"`
	Skype     string `json:"skype"`
	Phone     string `json:"phone"`
	Image24   string `json:"image_24"`
	Image32   string `json:"image_32"`
	Image48   string `json:"image_48"`
	Image72   string `json:"image_72"`
	Image192  string `json:"image_192"`
}

type Channel struct {
	Id                 string    `json:"id"`
	Name               string    `json:"name"`
	IsChannel          bool      `json:"is_channel"`
	Created            Timestamp `json:"created"`
	Creator            string    `json:"creator"`
	IsArchived         bool      `json:"is_archived"`
	IsGeneral          bool      `json:"is_general"`
	Members            []string  `json:"members"`
	Topic              *Topic    `json:"topic"`
	Purpose            *Topic    `json:"purpose"`
	IsMember           bool      `json:"is_member"`
	LastRead           Timestamp `json:"last_read"`
	UnreadCount        int       `json:"unread_count"`
	UnreadCountDisplay int       `json:"unread_count_display"`
	Latest             Message   `json:"latest"`
}

type Topic struct {
	Value   string    `json:"value"`
	Creator string    `json:"creator"`
	LastSet Timestamp `json:"last_set"`
}

type Group struct {
	Id                 string    `json:"id"`
	Name               string    `json:"name"`
	IsGroup            bool      `json:"is_group"`
	Created            Timestamp `json:"created"`
	Creator            string    `json:"creator"`
	IsArchived         bool      `json:"is_archived"`
	IsMPIM             bool      `json:"is_mpim"`
	Members            []string  `json:"members"`
	Topic              *Topic    `json:"topic"`
	Purpose            *Topic    `json:"purpose"`
	LastRead           Timestamp `json:"last_read"`
	UnreadCount        int       `json:"unread_count"`
	UnreadCountDisplay int       `json:"unread_count_display"`
	Latest             Message   `json:"latest"`
}

type InstantMessage struct {
	Id            string    `json:"id"`
	IsIM          bool      `json:"is_im"`
	User          string    `json:"user"`
	Created       Timestamp `json:"created"`
	IsUserDeleted bool      `json:"is_user_deleted"`
	Latest        Message   `json:"latest"`
}

type Icon struct {
	Image24  string `json:"image_24"`
	Image32  string `json:"image_32"`
	Image48  string `json:"image_48"`
	Image72  string `json:"image_72"`
	Image192 string `json:"image_192"`
}

type Bot struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Icons Icon   `json:"icons"`
}

type BareMessage struct {
	Type Event `json:"type"`
}

type Message struct {
	Id        string    `json:"id"`
	Type      Event     `json:"type"`
	SubType   string    `json:"subtype,omitempty"`
	Hidden    bool      `json:"hidden,omitempty"`
	Timestamp Timestamp `json:"ts,omitempty"`
	Channel   string    `json:"channel,omitempty"`
	User      string    `json:"user"`
	Text      string    `json:"text"`
}

type ChannelJoinedMessage struct {
	Type    Event   `json:"type"`
	Channel Channel `json:"channel,omitempty"`
}

type Self struct {
	Id             string    `json:"id"`
	Name           string    `json:"name"`
	Created        Timestamp `json:"created"`
	ManualPresense string    `json:"manual_presence"`
}

type Team struct {
	Id                string `json:"id"`
	Name              string `json:"name"`
	EmailDomain       string `json:"email_domain"`
	MsgEditWindowMins int    `json:"msg_edit_window_mins"`
	OverStorageLimit  bool   `json:"over_storage_limit"`
}

type Session struct {
	OK       bool             `json:"ok"`
	URL      string           `json:"url"`
	Self     *Self            `json:"self"`
	Team     *Team            `json:"team"`
	Users    []User           `json:"users"`
	Channels []Channel        `json:"channels"`
	Groups   []Group          `json:"groups"`
	IMs      []InstantMessage `json:"ims"`
	Error    string           `json:"error,omitempty"`
}

type basicResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

type ChatMessage struct {
	Token       string                  `json:"token"`
	Channel     string                  `json:"channel"`
	Text        string                  `json:"text"`
	Username    *string                 `json:"username,omitempty"`
	AsUser      *bool                   `json:"as_user,omitempty"`
	Parse       *string                 `json:"parse,omitempty"`
	LinkNames   *bool                   `json:"link_names,omitempty"`
	UnfurlLinks *bool                   `json:"unfurl_links,omitempty"`
	UnfurlMedia *bool                   `json:"unfurl_media,omitempty"`
	IconUrl     *string                 `json:"icon_url,omitempty"`
	IconEmoji   *string                 `json:"icon_emoji,omitempty"`
	Attachments []ChatMessageAttachment `json:"attachments,omitempty"`
}

type ChatMessageAttachment struct {
	Fallback      string  `json:"fallback"`
	Color         *string `json:"color"`
	Pretext       *string `json:"pretext,omitempty"`
	AuthorName    *string `json:"author_name,omitempty"`
	AuthorLink    *string `json:"author_link,omitempty"`
	AuthorIcon    *string `json:"author_icon,omitempty"`
	Title         *string `json:"title,omitempty"`
	TitleLink     *string `json:"title_link,omitempty"`
	Text          *string `json:"text,omitempty"`
	Fields        []Field `json:"fields,omitempty"`
	ImageUrl      *string `json:"image_url,omitempty"`
	ImageThumbUrl *string `json:"thumb_url,omitempty"`
}

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type Listener func(message *Message, client *Client)

// --------------------------------------------------------------------------------
// Connect(token string) is the main slack entry point.
// It returns a `*Client`, which is the root struct for interacting with slack.
// Register "Listeners" to hook into incoming messages and filter out ones you
// dont' care about.
// --------------------------------------------------------------------------------

func Connect(token string) *Client {
	c := &Client{Token: token, EventListeners: map[Event][]Listener{}, ActiveChannels: []string{}, activeLock: sync.Mutex{}}
	c.Listen(EVENT_CHANNEL_JOINED, c.handleChannelJoined)
	c.Listen(EVENT_CHANNEL_DELETED, c.handleChannelDeleted)
	c.Listen(EVENT_CHANNEL_UNARCHIVE, c.handleChannelUnarchive)
	c.Listen(EVENT_CHANNEL_LEFT, c.handleChannelLeft)
	return c
}

type Client struct {
	Token          string
	EventListeners map[Event][]Listener
	ActiveChannels []string

	activeLock       sync.Mutex
	socketConnection *websocket.Conn
}

func (rtm *Client) Listen(event Event, handler Listener) {
	if listeners, handlesEvent := rtm.EventListeners[event]; handlesEvent {
		rtm.EventListeners[event] = append(listeners, handler)
	} else {
		rtm.EventListeners[event] = []Listener{handler}
	}
}

func (rtm *Client) StopListening(event Event) {
	delete(rtm.EventListeners, event)
}

func (rtm *Client) Start() (*Session, error) {
	res := Session{}
	resErr := request.NewRequest().
		AsPost().
		WithScheme(API_SCHEME).
		WithHost(API_ENDPOINT).
		WithPath("api/rtm.start").
		WithPostData("token", rtm.Token).
		WithPostData("no_unreads", "true").
		WithPostData("mpim_aware", "true").
		FetchJsonToObject(&res)

	if resErr != nil {
		return nil, resErr
	}

	if !util.IsEmpty(res.Error) {
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
				rtm.ActiveChannels = append(rtm.ActiveChannels, channel.Id)
			}
		}
	}()

	go func() {
		rtm.listenLoop()
	}()

	return &res, nil
}

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

func (rtm Client) SendMessage(m *Message) error {
	if rtm.socketConnection == nil {
		return exception.New("Connection is closed.")
	}

	return rtm.socketConnection.WriteJSON(m)
}

func (rtm Client) Say(channelId string, messageComponents ...interface{}) error {
	if rtm.socketConnection == nil {
		return exception.New("Connection is closed.")
	}

	m := &Message{Type: "message", Text: fmt.Sprint(messageComponents...), Channel: channelId}
	return rtm.SendMessage(m)
}

func (rtm Client) Sayf(channelId, format string, messageComponents ...interface{}) error {
	if rtm.socketConnection == nil {
		return exception.New("Connection is closed.")
	}

	m := &Message{Type: "message", Text: fmt.Sprintf(format, messageComponents...), Channel: channelId}
	return rtm.SendMessage(m)
}

func (rtm *Client) Ping() error {
	if rtm.socketConnection == nil {
		return exception.New("Connection is closed.")
	}

	m := &Message{Id: util.UUID_v4().ToShortString(), Type: "ping"}
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

		body := string(messageBytes)

		var bm BareMessage
		jsonErr := util.DeserializeJson(&bm, body)
		if bm.Type == EVENT_CHANNEL_JOINED {
			var cm ChannelJoinedMessage
			jsonErr = util.DeserializeJson(&cm, body)
			if jsonErr == nil {
				rtm.dispatch(&Message{Type: EVENT_CHANNEL_JOINED, Channel: cm.Channel.Id})
			}
		} else {
			var m Message
			jsonErr = util.DeserializeJson(&m, body)
			if jsonErr == nil {
				rtm.dispatch(&m)
			}
		}
	}
	return nil
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

func (rtm *Client) removeActiveChannel(channelId string) {
	rtm.activeLock.Lock()
	defer rtm.activeLock.Unlock()

	currentChannels := []string{}
	for x := 0; x < len(rtm.ActiveChannels); x++ {
		currentChannelId := rtm.ActiveChannels[x]
		if channelId != currentChannelId {
			currentChannels = append(currentChannels, currentChannelId)
		}
	}
	rtm.ActiveChannels = currentChannels
}

//--------------------------------------------------------------------------------
// API METHODS
//--------------------------------------------------------------------------------

type channelsListResponse struct {
	Ok       bool      `json:"ok"`
	Error    string    `json:"error"`
	Channels []Channel `json:"channels"`
}

func (rtm *Client) ChannelsList(excludeArchived bool) ([]Channel, error) {
	res := channelsListResponse{}
	req := request.NewRequest().
		AsPost().
		WithScheme(API_SCHEME).
		WithHost(API_ENDPOINT).
		WithPath("api/channels.list").
		WithPostData("token", rtm.Token)

	if excludeArchived {
		req = req.WithPostData("exclude_archived", "1")
	}

	resErr := req.FetchJsonToObject(&res)

	if resErr != nil {
		return nil, resErr
	}

	if !util.IsEmpty(res.Error) {
		return nil, exception.New(res.Error)
	}

	return res.Channels, nil
}

type channelsInfoResponse struct {
	Ok      bool     `json:"ok"`
	Error   string   `json:"error"`
	Channel *Channel `json:"channel"`
}

func (rtm *Client) ChannelsInfo(channelId string) (*Channel, error) {
	res := channelsInfoResponse{}
	resErr := request.NewRequest().
		AsPost().
		WithScheme(API_SCHEME).
		WithHost(API_ENDPOINT).
		WithPath("api/channels.info").
		WithPostData("token", rtm.Token).
		WithPostData("channel", channelId).
		FetchJsonToObject(&res)

	if resErr != nil {
		return nil, resErr
	}

	if !util.IsEmpty(res.Error) {
		return nil, exception.New(res.Error)
	}

	return res.Channel, nil
}

func (rtm *Client) ChannelsSetTopic(channelId, topic string) error {
	res := basicResponse{}
	resErr := request.NewRequest().
		AsPost().
		WithScheme(API_SCHEME).
		WithHost(API_ENDPOINT).
		WithPath("api/channels.leave").
		WithPostData("token", rtm.Token).
		WithPostData("channel", channelId).
		WithPostData("topic", topic).
		FetchJsonToObject(&res)

	if resErr != nil {
		return resErr
	}

	if !util.IsEmpty(res.Error) {
		return exception.New(res.Error)
	}

	return nil
}

func (rtm *Client) ChannelsSetPurpose(channelId, purpose string) error {
	res := basicResponse{}
	resErr := request.NewRequest().
		AsPost().
		WithScheme(API_SCHEME).
		WithHost(API_ENDPOINT).
		WithPath("api/channels.leave").
		WithPostData("token", rtm.Token).
		WithPostData("channel", channelId).
		WithPostData("purpose", purpose).
		FetchJsonToObject(&res)

	if resErr != nil {
		return resErr
	}

	if !util.IsEmpty(res.Error) {
		return exception.New(res.Error)
	}

	return nil
}

type usersListResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
	Users []User `json:"members"`
}

func (rtm *Client) UsersList() ([]User, error) {
	res := usersListResponse{}
	resErr := request.NewRequest().
		AsPost().
		WithScheme(API_SCHEME).
		WithHost(API_ENDPOINT).
		WithPath("api/users.list").
		WithPostData("token", rtm.Token).
		FetchJsonToObject(&res)

	if resErr != nil {
		return nil, resErr
	}

	if !util.IsEmpty(res.Error) {
		return nil, exception.New(res.Error)
	}

	return res.Users, nil
}

type usersInfoResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
	User  *User  `json:"users"`
}

func (rtm *Client) UsersInfo(userId string) (*User, error) {
	res := usersInfoResponse{}
	resErr := request.NewRequest().
		AsPost().
		WithScheme(API_SCHEME).
		WithHost(API_ENDPOINT).
		WithPath("api/users.info").
		WithPostData("token", rtm.Token).
		WithPostData("user", userId).
		FetchJsonToObject(&res)

	if resErr != nil {
		return nil, resErr
	}

	if !util.IsEmpty(res.Error) {
		return nil, exception.New(res.Error)
	}

	return res.User, nil
}

func NewChatMessage(channelId, text string) *ChatMessage {
	return &ChatMessage{Channel: channelId, Text: text, Parse: OptionalString("full")}
}

type chatPostMessageResponse struct {
	Ok        bool         `json:"ok"`
	Timestamp Timestamp    `json:"timestamp"`
	Message   *ChatMessage `json:"message"`
	Error     string       `json:"error"`
}

func (rtm *Client) ChatPostMessage(m *ChatMessage) (*ChatMessage, error) { //the response version of the message is returned for verification
	m.Token = rtm.Token

	res := chatPostMessageResponse{}
	resErr := request.NewRequest().
		AsPost().
		WithScheme(API_SCHEME).
		WithHost(API_ENDPOINT).
		WithPath("api/chat.postMessage").
		WithPostDataFromObject(m).
		FetchJsonToObject(&res)

	if resErr != nil {
		return nil, resErr
	}

	if !util.IsEmpty(res.Error) {
		return nil, exception.New(res.Error)
	}

	return res.Message, nil
}

func OptionalBool(value bool) *bool {
	return &value
}

func OptionalString(value string) *string {
	return &value
}

func OptionalInt32(value int32) *int32 {
	return &value
}

func OptionalInt64(value int64) *int64 {
	return &value
}

func OptionalTimestamp(value time.Time) *time.Time {
	return &value
}

type Timestamp time.Time

func ParseTimestamp(strValue string) *Timestamp {
	if integerValue, integerErr := strconv.ParseInt(strValue, 10, 64); integerErr == nil {
		t := Timestamp(time.Unix(integerValue, 0))
		return &t
	}
	if _, floatErr := strconv.ParseFloat(strValue, 64); floatErr == nil {
		components := strings.Split(strValue, ".")
		if integerValue, integerErr := strconv.ParseInt(components[0], 10, 64); integerErr == nil {
			t := Timestamp(time.Unix(integerValue, 0))
			return &t
		}
	}
	return nil
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	strValue := string(data)
	t = ParseTimestamp(strValue)
	return nil
}

func (t Timestamp) DateTime() time.Time {
	return time.Time(t)
}
