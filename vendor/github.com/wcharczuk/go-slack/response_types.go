package slack

// NewChatMessage instantiates a ChatMessage for use with ChatPostMessage.
func NewChatMessage(channelID, text string) *ChatMessage {
	return &ChatMessage{Channel: channelID, Text: text, Parse: OptionalString("full")}
}

// User is the struct that represents a Slack user.
type User struct {
	ID                string       `json:"id"`
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

// UserProfile represents additional information about a Slack user.
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

// Channel is the struct that represents a Slack channel.
type Channel struct {
	ID                 string    `json:"id"`
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

// Topic represents a Slack topic.
type Topic struct {
	Value   string    `json:"value"`
	Creator string    `json:"creator"`
	LastSet Timestamp `json:"last_set"`
}

// Group represents a Slack group.
type Group struct {
	ID                 string    `json:"id"`
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

// InstantMessage represents a Slack instant message.
type InstantMessage struct {
	ID            string    `json:"id"`
	IsIM          bool      `json:"is_im"`
	User          string    `json:"user"`
	Created       Timestamp `json:"created"`
	IsUserDeleted bool      `json:"is_user_deleted"`
	Latest        Message   `json:"latest"`
}

// Icon represents a Slack icon.
type Icon struct {
	Image24  string `json:"image_24"`
	Image32  string `json:"image_32"`
	Image48  string `json:"image_48"`
	Image72  string `json:"image_72"`
	Image192 string `json:"image_192"`
}

// Bot represents a Slack bot.
type Bot struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Icons Icon   `json:"icons"`
}

// BareMessage is an intermediate type used to figure out what final type to deserialize a message as.
type BareMessage struct {
	Type Event `json:"type"`
}

// Message is a basic final message type that encapsulates the most commonly used fields.
type Message struct {
	ID        string     `json:"id"`
	Type      Event      `json:"type"`
	SubType   string     `json:"subtype,omitempty"`
	Hidden    bool       `json:"hidden,omitempty"`
	Timestamp *Timestamp `json:"ts,omitempty"`
	Channel   string     `json:"channel,omitempty"`
	User      string     `json:"user"`
	Text      string     `json:"text"`
	Reactions []Reaction `json:"reactions,omitempty"`
}

type Reaction struct {
	Name  string   `json:"name"`
	Count int      `json:"count"`
	Users []string `json:"users"`
}

// ChannelJoinedMessage is a final message type for the EventChannelJoined event type.
type ChannelJoinedMessage struct {
	Type    Event   `json:"type"`
	Channel Channel `json:"channel,omitempty"`
}

// Self represents information about the bot itself.
type Self struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Created        Timestamp `json:"created"`
	ManualPresense string    `json:"manual_presence"`
}

// Team represents information about a Slack team.
type Team struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	EmailDomain       string `json:"email_domain"`
	MsgEditWindowMins int    `json:"msg_edit_window_mins"`
	OverStorageLimit  bool   `json:"over_storage_limit"`
}

// Session represents information about a Slack session and is returned by APIStart.
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

// basicResponse is a utility intermediate type.
type basicResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

// ChatMessage is a struct that represents an outgoing chat message for the Slack chat message api.
type ChatMessage struct {
	Channel     string                  `json:"channel"`
	Text        string                  `json:"text"`
	Username    *string                 `json:"username,omitempty"`
	AsUser      *bool                   `json:"as_user,omitempty"`
	Parse       *string                 `json:"parse,omitempty"`
	LinkNames   *bool                   `json:"link_names,omitempty"`
	UnfurlLinks *bool                   `json:"unfurl_links,omitempty"`
	UnfurlMedia *bool                   `json:"unfurl_media,omitempty"`
	IconURL     *string                 `json:"icon_url,omitempty"`
	IconEmoji   *string                 `json:"icon_emoji,omitempty"`
	Attachments []ChatMessageAttachment `json:"attachments,omitempty"`
}

// ChatMessageAttachment is a struct that represents an attachment to a chat message for the Slack chat message api.
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
	ImageURL      *string `json:"image_url,omitempty"`
	ImageThumbURL *string `json:"thumb_url,omitempty"`
}

type File struct {
	ID                 string     `json:"id"`
	Created            Timestamp  `json:"created"`
	Timestamp          Timestamp  `json:"timestamp"`
	Name               string     `json:"name"`
	Title              string     `json:"title"`
	MimeType           string     `json:"mime_type"`
	FileType           string     `json:"filetype"`
	PrettyType         string     `json:"pretty_type"`
	UserID             string     `json:"user"`
	Mode               string     `json:"mode"`
	Editable           bool       `json:"editable"`
	IsExternal         bool       `json:"is_external"`
	ExternalType       string     `json:"external_type"`
	Username           string     `json:"username"`
	Size               int        `json:"size"`
	URLPrivate         string     `json:"url_private"`
	URLPrivateDownload string     `json:"url_private_download"`
	Thumb64            string     `json:"thumb_64"`
	Thumb80            string     `json:"thumb_80"`
	Thumb160           string     `json:"thumb_160"`
	Thumb360           string     `json:"thumb_360"`
	Thumb360GIF        string     `json:"thumb_360_gif"`
	Thumb360W          int        `json:"thumb_360_w"`
	Thumb360H          int        `json:"thumb_360_h"`
	Thumb480           string     `json:"thumb_480"`
	Thumb480W          int        `json:"thumb_480_w"`
	Thumb480H          int        `json:"thumb_480_h"`
	Permalink          string     `json:"permalink"`
	PermalinkPublic    string     `json:"permalink_public"`
	EditLink           string     `json:"edit_link"`
	Preview            string     `json:"preview"`
	PreviewHighlight   string     `json:"preview_highlight"`
	Lines              int        `json:"lines"`
	LinesMore          int        `json:"lines_more"`
	IsPublic           bool       `json:"is_public"`
	PublicURLShared    bool       `json:"public_url_shared"`
	DisplayAsBot       bool       `json:"display_as_bot"`
	Channels           []string   `json:"channels"`
	Groups             []string   `json:"groups"`
	IMs                []string   `json:"ims"`
	NumStars           int        `json:"num_stars"`
	IsStarred          bool       `json:"is_starred"`
	PinnedTo           []string   `json:"pinned_to"`
	Reactions          []Reaction `json:"reactions"`
	CommentsCount      int        `json:"comments_count"`
}

// Field represents a field on a Slack ChatMessageAttachment.
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// AuthTestArgs is an alias for the general format json.Unmarshal uses for JSON objects.
type ApiTestArgs map[string]interface{}

type ApiTestResponse struct {
	OK    bool        `json:"ok"`
	Error string      `json:"error,omitempty"`
	Args  ApiTestArgs `json:"args"`
}

// AuthTestResponse is the response format from slack for auth.test endpoint.
type AuthTestResponse struct {
	OK     bool   `json:"ok"`
	URL    string `json:"url,omitempty"`
	Team   string `json:"team,omitempty"`
	User   string `json:"user,omitemtpy"`
	TeamID string `json:"team_id,omitempty"`
	UserID string `json:"user_id,omitempty"`
	Error  string `json:"error,omitempty"`
}

type ChannelsHistoryResponse struct {
	OK                 bool      `json:"ok"`
	Error              string    `json:"error"`
	Latest             Timestamp `json:"latest"`
	IsLimited          bool      `json:"is_limited"`
	HasMore            bool      `json:"has_more"`
	UnreadCountDisplay int       `json:"unread_count_display"`
	Messages           []Message `json:"messages"`
}

type channelsListResponse struct {
	OK       bool      `json:"ok"`
	Error    string    `json:"error"`
	Channels []Channel `json:"channels"`
}

type channelsInfoResponse struct {
	OK      bool     `json:"ok"`
	Error   string   `json:"error"`
	Channel *Channel `json:"channel"`
}

type emojiResponse struct {
	OK    bool              `json:"ok"`
	Error string            `json:"error"`
	Emoji map[string]string `json:"emoji"`
}

type usersListResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
	Users []User `json:"members"`
}

type usersInfoResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
	User  *User  `json:"users"`
}

type ChatMessageResponse struct {
	OK          bool      `json:"ok"`
	Timestamp   Timestamp `json:"timestamp"`
	Message     *Message  `json:"message,omitempty"`
	File        *File     `json:"file,omitempty"`
	FileComment *File     `json:"file_comment,omitempty"`
	Error       string    `json:"error"`
}
