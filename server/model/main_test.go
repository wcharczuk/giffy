package model

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
)

func TestMain(m *testing.M) {
	core.DBInit()
	os.Exit(m.Run())
}

func createTestTag(userID int64, tagValue string, tx *sql.Tx) (*Tag, error) {
	tag := NewTag(userID, tagValue)
	err := spiffy.DefaultDb().CreateInTransaction(tag, tx)
	return tag, err
}

func createTestTagForImage(userID, imageID int64, tagValue string, tx *sql.Tx) (*Tag, error) {
	tag, err := createTestTag(userID, tagValue, tx)
	if err != nil {
		return nil, err
	}

	existing, existingErr := GetVoteSummary(imageID, tag.ID, tx)
	if existingErr != nil {
		return nil, existingErr
	}

	if existing.IsZero() {
		v := NewVoteSummary(imageID, tag.ID, userID, time.Now().UTC())
		v.VotesFor = 1
		v.VotesAgainst = 0
		v.VotesTotal = 1
		err = spiffy.DefaultDb().CreateInTransaction(v, tx)
	}
	return tag, err
}

func createTestImage(userID int64, tx *sql.Tx) (*Image, error) {
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
	err := spiffy.DefaultDb().CreateInTransaction(i, tx)
	return i, err
}

func createTestUser(tx *sql.Tx) (*User, error) {
	u := NewUser(fmt.Sprintf("__test_user_%s__", core.UUIDv4().ToShortString()))
	u.FirstName = "Test"
	u.LastName = "User"
	err := spiffy.DefaultDb().CreateInTransaction(u, tx)
	return u, err
}

func createTestUserAuth(userID int64, token, secret string, tx *sql.Tx) (*UserAuth, error) {
	ua := NewUserAuth(userID, token, secret)
	ua.Provider = "test"
	err := spiffy.DefaultDb().CreateInTransaction(ua, tx)
	return ua, err
}

func createTestUserSession(userID int64, tx *sql.Tx) (*UserSession, error) {
	us := NewUserSession(userID)
	err := spiffy.DefaultDb().CreateInTransaction(us, tx)
	return us, err
}
