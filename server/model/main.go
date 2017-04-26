package model

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
)

// DB is a helper for returning the default database connection.
func DB() *spiffy.Connection {
	return spiffy.Default()
}

// CreateObject creates an object (for use with the work queue)
func CreateObject(state ...interface{}) error {
	return DB().Create(state[0].(spiffy.DatabaseMapped))
}

// CreateTestUser creates a test user.
func CreateTestUser(tx *sql.Tx) (*User, error) {
	u := NewUser(fmt.Sprintf("__test_user_%s__", core.UUIDv4().ToShortString()))
	u.FirstName = "Test"
	u.LastName = "User"
	err := DB().CreateInTx(u, tx)
	return u, err
}

// CreateTestTag creates a test tag.
func CreateTestTag(userID int64, tagValue string, tx *sql.Tx) (*Tag, error) {
	tag := NewTag(userID, tagValue)
	err := DB().CreateInTx(tag, tx)
	return tag, err
}

// CreateTestTagForImageWithVote creates a test tag for an image and seeds the vote summary for it.
func CreateTestTagForImageWithVote(userID, imageID int64, tagValue string, tx *sql.Tx) (*Tag, error) {
	tag, err := CreateTestTag(userID, tagValue, tx)
	if err != nil {
		return nil, err
	}

	existing, err := GetVoteSummary(imageID, tag.ID, tx)
	if err != nil {
		return nil, err
	}

	if existing.IsZero() {
		v := NewVoteSummary(imageID, tag.ID, userID, time.Now().UTC())
		v.VotesFor = 1
		v.VotesAgainst = 0
		v.VotesTotal = 1
		err = DB().CreateInTx(v, tx)
		if err != nil {
			return nil, err
		}
	}

	err = DeleteVote(userID, imageID, tag.ID, tx)
	if err != nil {
		return nil, err
	}

	v := NewVote(userID, imageID, tag.ID, true)
	err = DB().CreateInTx(v, tx)
	if err != nil {
		return nil, err
	}
	return tag, err
}

// CreateTestImage creates a test image.
func CreateTestImage(userID int64, tx *sql.Tx) (*Image, error) {
	i := NewImage()
	i.CreatedBy = userID
	i.Extension = "gif"
	i.Width = 720
	i.Height = 480
	i.S3Bucket = core.UUIDv4().ToShortString()
	i.S3Key = core.UUIDv4().ToShortString()
	i.S3ReadURL = fmt.Sprintf("https://s3.amazonaws.com/%s/%s", i.S3Bucket, i.S3Key)
	i.MD5 = core.UUIDv4()
	i.DisplayName = "Test Image"
	err := DB().CreateInTx(i, tx)
	return i, err
}

// CreatTestVoteSummaryWithVote creates a test vote summary with an accopanying vote.
func CreatTestVoteSummaryWithVote(imageID, tagID, userID int64, votesFor, votesAgainst int, tx *sql.Tx) (*VoteSummary, error) {
	newLink := NewVoteSummary(imageID, tagID, userID, time.Now().UTC())
	newLink.VotesFor = votesFor
	newLink.VotesTotal = votesAgainst
	newLink.VotesTotal = votesFor - votesAgainst
	err := DB().CreateInTx(newLink, tx)

	if err != nil {
		return nil, err
	}

	err = DeleteVote(userID, imageID, tagID, tx)
	if err != nil {
		return nil, err
	}

	v := NewVote(userID, imageID, tagID, true)
	err = DB().CreateInTx(v, tx)
	if err != nil {
		return nil, err
	}

	return newLink, nil
}

// CreateTestUserAuth creates a test user auth.
func CreateTestUserAuth(userID int64, token, secret string, tx *sql.Tx) (*UserAuth, error) {
	ua, err := NewUserAuth(userID, token, secret)
	if err != nil {
		return ua, err
	}
	ua.Provider = "test"
	err = DB().CreateInTx(ua, tx)
	return ua, err
}

// CreateTestUserSession creates a test user session.
func CreateTestUserSession(userID int64, tx *sql.Tx) (*UserSession, error) {
	us := NewUserSession(userID)
	err := DB().CreateInTx(us, tx)
	return us, err
}

// CreateTestSearchHistory creates a test search history entry.
func CreateTestSearchHistory(source, searchQuery string, imageID, tagID *int64, tx *sql.Tx) (*SearchHistory, error) {
	sh := NewSearchHistory(source, searchQuery, true, imageID, tagID)
	err := DB().CreateInTx(sh, tx)
	return sh, err
}
