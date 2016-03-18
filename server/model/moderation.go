package model

import (
	"database/sql"
	"time"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
)

const (
	// ModerationVerbCreate = "create"
	ModerationVerbCreate = "create"

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
func NewModeration(userID int64, verb, object, name string) *Moderation {
	return &Moderation{
		UserID:        userID,
		UUID:          core.UUIDv4().ToShortString(),
		TimestampUTC:  time.Now().UTC(),
		Verb:          verb,
		Noun:          object,
		SecondaryNoun: name,
	}
}

// Moderation is the moderation log.
type Moderation struct {
	UserID        int64     `json:"user_id" db:"user_id"`
	UUID          string    `json:"uuid" db:"uuid,pk"`
	TimestampUTC  time.Time `json:"timestamp_utc" db:"timestamp_utc"`
	Verb          string    `json:"verb" db:"verb"`
	Noun          string    `json:"noun" db:"noun"`
	SecondaryNoun string    `json:"name" db:"secondary_noun"`
}

// TableName returns the table
func (m Moderation) TableName() string {
	return "moderation"
}

// IsZero returns if the object is set.
func (m Moderation) IsZero() bool {
	return m.UserID == 0
}

func writeModerationLogEntry(state interface{}) error {
	if typed, isTyped := state.(*Moderation); isTyped {
		return spiffy.DefaultDb().Create(typed)
	}
	return exception.New("`state` was not of the correct type.")
}

// QueueModerationEntry queues logging a new moderation log entry.
func QueueModerationEntry(userID int64, verb, object, name string) {
	m := NewModeration(userID, verb, object, name)
	util.QueueWorkItem(writeModerationLogEntry, m)
}

// GetModerationsForUser gets all the moderation entries for a user.
func GetModerationsForUser(userID int64, tx *sql.Tx) ([]Moderation, error) {
	var moderation []Moderation
	query := `select * from moderation where user_id = $1 order by timestamp_utc desc`
	err := spiffy.DefaultDb().QueryInTransaction(query, tx, userID).OutMany(&moderation)
	return moderation, err
}

// GetModerationsByTime returns all moderation entries after a specific time.
func GetModerationsByTime(after time.Time, tx *sql.Tx) ([]Moderation, error) {
	var moderation []Moderation
	query := `select * from moderation where timestamp_utc > $1 order by timestamp_utc desc`
	err := spiffy.DefaultDb().QueryInTransaction(query, tx, after).OutMany(&moderation)
	return moderation, err
}

// GetModerationsByCountAndOffset returns all moderation entries after a specific time.
func GetModerationsByCountAndOffset(count, offset int, tx *sql.Tx) ([]Moderation, error) {
	var moderation []Moderation
	query := `select * from moderation where order by timestamp_utc desc limit $1 offset $2`
	err := spiffy.DefaultDb().QueryInTransaction(query, tx, count, offset).OutMany(&moderation)
	return moderation, err
}
