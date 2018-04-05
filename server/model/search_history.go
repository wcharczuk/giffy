package model

import (
	"database/sql"
	"fmt"
	"time"

	logger "github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/db"
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

func createSearchHistoryQuery(whereClause ...string) string {
	searchColumns := db.Columns(SearchHistory{}).NotReadOnly().ColumnNamesCSVFromAlias("sh")
	imageColumns := db.Columns(Image{}).NotReadOnly().CopyWithColumnPrefix("image_").ColumnNamesCSVFromAlias("i")
	tagColumns := db.Columns(Tag{}).NotReadOnly().CopyWithColumnPrefix("tag_").ColumnNamesCSVFromAlias("t")

	query := `
	select
		%s,
		%s,
		%s
	from
		search_history sh
		left join image i on i.id = sh.image_id
		left join tag t on t.id = sh.tag_id
	%s
	order by timestamp_utc desc
	`
	if len(whereClause) > 0 {
		return fmt.Sprintf(query, searchColumns, imageColumns, tagColumns, whereClause[0])
	}
	return fmt.Sprintf(query, searchColumns, imageColumns, tagColumns, "")
}

func searchHistoryConsumer(searchHistory *[]SearchHistory) db.RowsConsumer {
	searchColumns := db.Columns(SearchHistory{}).NotReadOnly()
	imageColumns := db.Columns(Image{}).NotReadOnly().CopyWithColumnPrefix("image_")
	tagColumns := db.Columns(Tag{}).NotReadOnly().CopyWithColumnPrefix("tag_")

	return func(r *sql.Rows) error {
		var sh SearchHistory
		var i Image
		var t Tag

		err := db.PopulateByName(&sh, r, searchColumns)
		if err != nil {
			return err
		}

		err = db.PopulateByName(&i, r, imageColumns)
		if err == nil && !i.IsZero() {
			sh.Image = &i
		}

		err = db.PopulateByName(&t, r, tagColumns)
		if err == nil && !t.IsZero() {
			sh.Tag = &t
		}

		*searchHistory = append(*searchHistory, sh)
		return nil
	}
}

// GetSearchHistory returns the entire search history in chrono order.
func GetSearchHistory(txs ...*sql.Tx) ([]SearchHistory, error) {
	var searchHistory []SearchHistory
	query := createSearchHistoryQuery()
	err := DB().QueryInTx(query, db.OptionalTx(txs...)).Each(searchHistoryConsumer(&searchHistory))
	return searchHistory, err
}

// GetSearchHistoryByCountAndOffset returns the search history in chrono order by count and offset.
func GetSearchHistoryByCountAndOffset(count, offset int, tx *sql.Tx) ([]SearchHistory, error) {
	var searchHistory []SearchHistory
	query := createSearchHistoryQuery()
	query = query + `limit $1 offset $2`
	err := DB().QueryInTx(query, tx, count, offset).Each(searchHistoryConsumer(&searchHistory))
	return searchHistory, err
}
