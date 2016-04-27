package controller

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/auth"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/model"
)

func TestMain(m *testing.M) {
	// we do this because a lot of static results depend on relative paths.
	core.Setwd("../../")
	core.DBInit()
	os.Exit(m.Run())
}

func MockAdminLogin(tx *sql.Tx) (*auth.Session, error) {
	u, err := CreateTestAdminUser(tx)
	if err != nil {
		return nil, err
	}
	return AuthTestUser(u.ID, tx)
}

func MockModeratorLogin(tx *sql.Tx) (*auth.Session, error) {
	u, err := CreateTestModeratorUser(tx)
	if err != nil {
		return nil, err
	}
	return AuthTestUser(u.ID, tx)
}

func MockBannedLogin(tx *sql.Tx) (*auth.Session, error) {
	u, err := CreateTestModeratorUser(tx)
	if err != nil {
		return nil, err
	}
	return AuthTestUser(u.ID, tx)
}

func CreateTestAdminUser(tx *sql.Tx) (*model.User, error) {
	u := model.NewUser(fmt.Sprintf("__test_user_%s__", core.UUIDv4().ToShortString()))
	u.FirstName = "Test"
	u.LastName = "User"
	u.IsAdmin = true
	u.IsModerator = true
	err := spiffy.DefaultDb().CreateInTransaction(u, tx)
	return u, err
}

func CreateTestModeratorUser(tx *sql.Tx) (*model.User, error) {
	u := model.NewUser(fmt.Sprintf("__test_user_%s__", core.UUIDv4().ToShortString()))
	u.FirstName = "Test"
	u.LastName = "User"
	u.IsAdmin = false
	u.IsModerator = true
	err := spiffy.DefaultDb().CreateInTransaction(u, tx)
	return u, err
}

func CreateTestBannedUser(tx *sql.Tx) (*model.User, error) {
	u := model.NewUser(fmt.Sprintf("__test_user_%s__", core.UUIDv4().ToShortString()))
	u.FirstName = "Test"
	u.LastName = "User"
	u.IsAdmin = false
	u.IsModerator = false
	u.IsBanned = true
	err := spiffy.DefaultDb().CreateInTransaction(u, tx)
	return u, err
}

func AuthTestUser(userID int64, tx *sql.Tx) (*auth.Session, error) {
	sessionID, err := auth.Login(userID, nil, tx)
	if err != nil {
		return nil, err
	}
	return auth.SessionState().Get(sessionID), nil
}

// to undo the above defer auth.Logout(session.UserID, session.SessionID, tx)
