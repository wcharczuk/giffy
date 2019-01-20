package model

import (
	"time"

	logger "github.com/blend/go-sdk/logger"
	"github.com/wcharczuk/giffy/server/core"
)

// NewSearchHistory returns a new search history.
func NewSearchHistory(source, searchQuery string, didFindMatch bool, imageID, tagID *int64) *SearchHistory {
	return &SearchHistory{
		Source:       source,
		TimestampUTC: time.Now().UTC(),
		SearchQuery:  searchQuery,
		DidFindMatch: didFindMatch,
		ImageID:      imageID,
		TagID:        tagID,
	}
}

// NewSearchHistoryDetailed queues logging a new moderation log entry.
func NewSearchHistoryDetailed(source, sourceTeamID, sourceTeamName, sourceChannelID, sourceChannelName, sourceUserID, sourceUserName, searchQuery string, didFindMatch bool, imageID, tagID *int64) *SearchHistory {
	return &SearchHistory{
		Source:       source,
		TimestampUTC: time.Now().UTC(),
		SearchQuery:  searchQuery,
		DidFindMatch: didFindMatch,
		ImageID:      imageID,
		TagID:        tagID,

		//below are the "detailed" fields
		SourceTeamIdentifier:    sourceTeamID,
		SourceChannelIdentifier: sourceChannelID,
		SourceUserIdentifier:    sourceUserID,
		SourceTeamName:          sourceTeamName,
		SourceChannelName:       sourceChannelName,
		SourceUserName:          sourceUserName,
	}
}

// SearchHistory is a record of searches and the primary result.
type SearchHistory struct {
	Source string `json:"source" db:"source"`

	SourceTeamIdentifier    string `json:"source_team_identifier" db:"source_team_identifier"`
	SourceTeamName          string `json:"source_team_name" db:"source_team_name"`
	SourceChannelIdentifier string `json:"source_channel_identifier" db:"source_channel_identifier"`
	SourceChannelName       string `json:"source_channel_name" db:"source_channel_name"`
	SourceUserIdentifier    string `json:"source_user_identifier" db:"source_user_identifier"`
	SourceUserName          string `json:"source_user_name" db:"source_user_name"`

	TimestampUTC time.Time `json:"timestamp_utc" db:"timestamp_utc"`
	SearchQuery  string    `json:"search_query" db:"search_query"`

	DidFindMatch bool `json:"did_find_match" db:"did_find_match"`

	ImageID *int64 `json:"image_id" db:"image_id"`
	Image   *Image `json:"image" db:"-"`
	TagID   *int64 `json:"tag_id" db:"tag_id"`
	Tag     *Tag   `json:"tag" db:"-"`
}

// TableName returns the table name
func (sh SearchHistory) TableName() string {
	return "search_history"
}

// Flag implements logger.Event.
func (sh SearchHistory) Flag() logger.Flag {
	return core.FlagSearch
}

// Timestamp implements logger.Event.
func (sh SearchHistory) Timestamp() time.Time {
	return sh.TimestampUTC
}
