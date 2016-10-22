package model

import "time"

// SlackTeam is a team that is mapped to giffy.
type SlackTeam struct {
	TeamID              string    `json:"team_id" db:"team_id,pk"`
	TimestampUTC        time.Time `json:"timestamp_utc" db:"timestamp_utc"`
	IsEnabled           bool      `json:"is_enabled" db:"is_enabled"`
	CreatedBy           string    `json:"created_by" db:"created_by"`
	ContentRatingFilter int       `json:"content_rating" db:"content_rating"`
}

// TableName returns the mapped table name.
func (st SlackTeam) TableName() string {
	return "slack_team"
}

// IsZero returns if the object has been set or not.
func (st SlackTeam) IsZero() bool {
	return len(st.TeamID) == 0
}
