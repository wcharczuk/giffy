package model

import (
	"fmt"
	"io"
	"time"

	logger "github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/uuid"
	"github.com/wcharczuk/giffy/server/core"
)

const (
	// ModerationVerbCreate = "create"
	ModerationVerbCreate = "create"
	// ModerationVerbUpdate = "update"
	ModerationVerbUpdate = "update"
	// ModerationVerbDelete = "delete"
	ModerationVerbDelete = "delete"
	// ModerationVerbConsolidate = "consolidate"
	ModerationVerbConsolidate = "consolidate"
	// ModerationVerbPromoteAsModerator = "promote_moderator"
	ModerationVerbPromoteAsModerator = "promote_moderator"
	// ModerationVerbDemoteAsModerator = "demote_moderator"
	ModerationVerbDemoteAsModerator = "demote_moderator"
	// ModerationVerbBan = "ban"
	ModerationVerbBan = "ban"
	// ModerationVerbUnban = "unban"
	ModerationVerbUnban = "unban"
	// ModerationObjectImage = "image"
	ModerationObjectImage = "image"
	// ModerationObjectTag = "tag"
	ModerationObjectTag = "tag"
	// ModerationObjectLink = "link"
	ModerationObjectLink = "link"
	// ModerationObjectUser = "user"
	ModerationObjectUser = "user"
)

// NewModeration returns a new moderation object
func NewModeration(userID int64, verb, object string, nouns ...string) *Moderation {
	m := &Moderation{
		UserID:       userID,
		UUID:         uuid.V4().String(),
		TimestampUTC: time.Now().UTC(),
		Verb:         verb,
		Object:       object,
	}
	if len(nouns) > 0 {
		m.Noun = nouns[0]
	}
	if len(nouns) > 1 {
		m.SecondaryNoun = nouns[1]
	}
	return m
}

var (
	_ logger.Event        = (*Moderation)(nil)
	_ logger.TextWritable = (*Moderation)(nil)
)

// Moderation is the moderation log.
type Moderation struct {
	UserID        int64     `json:"user_id" db:"user_id"`
	UUID          string    `json:"uuid" db:"uuid,pk"`
	TimestampUTC  time.Time `json:"timestamp_utc" db:"timestamp_utc"`
	Verb          string    `json:"verb" db:"verb"`
	Object        string    `json:"object" db:"object"`
	Noun          string    `json:"noun" db:"noun"`
	SecondaryNoun string    `json:"secondary_noun" db:"secondary_noun"`

	Moderator *User `json:"moderator" db:"-"`

	User  *User  `json:"user,omitempty" db:"-"`
	Image *Image `json:"image,omitempty" db:"-"`
	Tag   *Tag   `json:"tag,omitempty" db:"-"`
}

// TableName returns the table
func (m Moderation) TableName() string {
	return "moderation"
}

// IsZero returns if the object is set.
func (m Moderation) IsZero() bool {
	return m.UserID == 0
}

// Flag implements logger.event.
func (m Moderation) GetFlag() string {
	return core.FlagModeration
}

// WriteText implements logger.TextWritable.
func (m Moderation) WriteText(tf logger.TextFormatter, output io.Writer) {
	if m.Moderator != nil {
		fmt.Fprintf(output, "Moderator: %s", m.Moderator.Username)
	}
	if m.Verb != "" {
		fmt.Fprint(output, " Verb: ")
		fmt.Fprint(output, m.Verb)
	}
	if m.Object != "" {
		fmt.Fprint(output, " Object: ")
		fmt.Fprint(output, m.Object)
	}
	if m.Noun != "" {
		fmt.Fprint(output, " Noun: ")
		fmt.Fprint(output, m.Noun)
	}
	if m.SecondaryNoun != "" {
		fmt.Fprint(output, " Secondary Noun: ")
		fmt.Fprint(output, m.SecondaryNoun)
	}
}
