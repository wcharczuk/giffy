package model

import (
	"database/sql"
	"fmt"
	"time"

	logger "github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/db"
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
		UUID:         core.UUIDv4().ToShortString(),
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
func (m Moderation) Flag() logger.Flag {
	return core.FlagModeration
}

// Timestamp implements logger.event.
func (m Moderation) Timestamp() time.Time {
	return m.TimestampUTC
}

func getModerationQuery(whereClause string) string {
	moderatorColumns := db.Columns(User{}).NotReadOnly().CopyWithColumnPrefix("moderator_").ColumnNamesCSVFromAlias("mu")
	userColumns := db.Columns(User{}).NotReadOnly().CopyWithColumnPrefix("target_user_").ColumnNamesCSVFromAlias("u")
	imageColumns := db.Columns(Image{}).NotReadOnly().CopyWithColumnPrefix("image_").ColumnNamesCSVFromAlias("i")
	tagColumns := db.Columns(Tag{}).NotReadOnly().CopyWithColumnPrefix("tag_").ColumnNamesCSVFromAlias("t")

	return fmt.Sprintf(`
	select
		m.*,
		%s,
		%s,
		%s,
		%s
	from
		moderation m
		join users mu on m.user_id = mu.id
		left join users u on m.noun = u.uuid or m.secondary_noun = u.uuid
		left join image i on m.noun = i.uuid or m.secondary_noun = i.uuid
		left join tag t on m.noun = t.uuid or m.secondary_noun = t.uuid
	%s
	order by timestamp_utc desc
	`,
		moderatorColumns,
		userColumns,
		imageColumns,
		tagColumns,
		whereClause)
}

// GetModerationForUserID gets all the moderation entries for a user.
func GetModerationForUserID(userID int64, tx *sql.Tx) ([]Moderation, error) {
	var moderationLog []Moderation
	whereClause := `where user_id = $1`
	err := DB().QueryInTx(getModerationQuery(whereClause), tx, userID).Each(moderationConsumer(&moderationLog))
	return moderationLog, err
}

// GetModerationsByTime returns all moderation entries after a specific time.
func GetModerationsByTime(after time.Time, tx *sql.Tx) ([]Moderation, error) {
	var moderationLog []Moderation
	whereClause := `where timestamp_utc > $1`
	err := DB().QueryInTx(getModerationQuery(whereClause), tx, after).Each(moderationConsumer(&moderationLog))
	return moderationLog, err
}

// GetModerationLogByCountAndOffset returns all moderation entries after a specific time.
func GetModerationLogByCountAndOffset(count, offset int, tx *sql.Tx) ([]Moderation, error) {
	var moderationLog []Moderation
	query := getModerationQuery("")
	query = query + `limit $1 offset $2`
	err := DB().QueryInTx(query, tx, count, offset).Each(moderationConsumer(&moderationLog))
	return moderationLog, err
}

func moderationConsumer(moderationLog *[]Moderation) db.RowsConsumer {
	moderationColumns := db.Columns(Moderation{})
	moderatorColumns := db.Columns(User{}).NotReadOnly().CopyWithColumnPrefix("moderator_")
	userColumns := db.Columns(User{}).NotReadOnly().CopyWithColumnPrefix("target_user_")
	imageColumns := db.Columns(Image{}).NotReadOnly().CopyWithColumnPrefix("image_")
	tagColumns := db.Columns(Tag{}).NotReadOnly().CopyWithColumnPrefix("tag_")

	return func(r *sql.Rows) error {
		var m Moderation
		var mu User

		var u User
		var i Image
		var t Tag

		err := db.PopulateByName(&m, r, moderationColumns)
		if err != nil {
			return err
		}

		err = db.PopulateByName(&mu, r, moderatorColumns)
		if err != nil {
			return err
		}
		m.Moderator = &mu

		err = db.PopulateByName(&u, r, userColumns)
		if err == nil && !u.IsZero() {
			m.User = &u
		}

		err = db.PopulateByName(&i, r, imageColumns)
		if err == nil && !i.IsZero() {
			m.Image = &i
		}

		err = db.PopulateByName(&t, r, tagColumns)
		if err == nil && !t.IsZero() {
			m.Tag = &t
		}

		*moderationLog = append(*moderationLog, m)
		return nil
	}
}
