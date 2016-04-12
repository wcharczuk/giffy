package model

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
	"github.com/blendlabs/spiffy"
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

func searchHistoryQuery(whereClause string) string {
	searchColumns := spiffy.CSV(spiffy.NewColumnCollectionFromInstance(SearchHistory{}).NotReadOnly().ColumnNamesFromAlias("sh"))
	imageColumns := spiffy.CSV(spiffy.NewColumnCollectionFromInstance(Image{}).NotReadOnly().WithColumnPrefix("image_").ColumnNamesFromAlias("i"))
	tagColumns := spiffy.CSV(spiffy.NewColumnCollectionFromInstance(Tag{}).NotReadOnly().WithColumnPrefix("tag_").ColumnNamesFromAlias("t"))
	return fmt.Sprintf(`
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
`, searchColumns, imageColumns, tagColumns, whereClause)
}

func searchHistoryConsumer(searchHistory *[]SearchHistory) spiffy.RowsConsumer {
	searchColumns := spiffy.NewColumnCollectionFromInstance(SearchHistory{}).NotReadOnly()
	imageColumns := spiffy.NewColumnCollectionFromInstance(Image{}).NotReadOnly().WithColumnPrefix("image_")
	tagColumns := spiffy.NewColumnCollectionFromInstance(Tag{}).NotReadOnly().WithColumnPrefix("tag_")

	return func(r *sql.Rows) error {
		var sh SearchHistory
		var i Image
		var t Tag

		err := spiffy.PopulateByName(&sh, r, searchColumns)
		if err != nil {
			return err
		}

		err = spiffy.PopulateByName(&i, r, imageColumns)
		if err == nil && !i.IsZero() {
			sh.Image = &i
		}

		err = spiffy.PopulateByName(&t, r, tagColumns)
		if err == nil && !t.IsZero() {
			sh.Tag = &t
		}

		*searchHistory = append(*searchHistory, sh)
		return nil
	}
}

// GetSearchHistory returns the entire search history in chrono order.
func GetSearchHistory(tx *sql.Tx) ([]SearchHistory, error) {
	var searchHistory []SearchHistory
	query := searchHistoryQuery("")
	err := spiffy.DefaultDb().QueryInTransaction(query, tx).Each(searchHistoryConsumer(&searchHistory))
	return searchHistory, err
}

// GetSearchHistoryByCountAndOffset returns the search history in chrono order by count and offset.
func GetSearchHistoryByCountAndOffset(count, offset int, tx *sql.Tx) ([]SearchHistory, error) {
	var searchHistory []SearchHistory
	query := searchHistoryQuery("")
	query = query + `limit $1 offset $2`
	err := spiffy.DefaultDb().QueryInTransaction(query, tx, count, offset).Each(searchHistoryConsumer(&searchHistory))
	return searchHistory, err
}

func writeSearchHistoryEntry(state interface{}) error {
	if typed, isTyped := state.(*SearchHistory); isTyped {
		return spiffy.DefaultDb().Create(typed)
	}
	return exception.New("`state` was not of the correct type.")
}

// QueueSearchHistoryEntry queues logging a new moderation log entry.
func QueueSearchHistoryEntry(source, sourceTeamID, sourceTeamName, sourceChannelID, sourceChannelName, sourceUserID, sourceUserName, searchQuery string, didFindMatch bool, imageID, tagID *int64) {
	sh := NewSearchHistory(source, searchQuery, didFindMatch, imageID, tagID)
	sh.SourceTeamIdentifier = sourceTeamID
	sh.SourceChannelIdentifier = sourceChannelID
	sh.SourceUserIdentifier = sourceUserID

	sh.SourceTeamName = sourceTeamName
	sh.SourceChannelName = sourceChannelName
	sh.SourceUserName = sourceUserName

	util.QueueWorkItem(writeSearchHistoryEntry, sh)
}