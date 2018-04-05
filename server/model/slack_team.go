package model

import (
	"database/sql"
	"time"

	"github.com/blend/go-sdk/spiffy"
)

// NewSlackTeam returns a new SlackTeam.
func NewSlackTeam(teamID, teamName, userID, userName string) *SlackTeam {
	return &SlackTeam{
		TeamID:              teamID,
		TeamName:            teamName,
		CreatedUTC:          time.Now().UTC(),
		IsEnabled:           true,
		CreatedByID:         userID,
		CreatedByName:       userName,
		ContentRatingFilter: ContentRatingPG13,
	}
}

// SlackTeam is a team that is mapped to giffy.
type SlackTeam struct {
	TeamID              string    `json:"team_id" db:"team_id,pk"`
	TeamName            string    `json:"team_name" db:"team_name"`
	CreatedUTC          time.Time `json:"created_utc" db:"created_utc"`
	IsEnabled           bool      `json:"is_enabled" db:"is_enabled"`
	CreatedByID         string    `json:"created_by_id" db:"created_by_id"`
	CreatedByName       string    `json:"created_by_name" db:"created_by_name"`
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

// GetAllSlackTeams gets all slack teams.
func GetAllSlackTeams(txs ...*sql.Tx) ([]SlackTeam, error) {
	var teams []SlackTeam
	err := DB().QueryInTx(`select * from slack_team order by team_name asc`, spiffy.OptionalTx(txs...)).OutMany(&teams)
	return teams, err
}

// GetSlackTeamByTeamID gets a slack team by the team id.
func GetSlackTeamByTeamID(teamID string, txs ...*sql.Tx) (*SlackTeam, error) {
	var team SlackTeam
	err := DB().InTx(txs...).Get(&team, teamID)
	return &team, err
}
