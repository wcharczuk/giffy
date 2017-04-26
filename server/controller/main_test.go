package controller

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/blendlabs/go-assert"
	web "github.com/blendlabs/go-web"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/webutil"
)

func TestMain(m *testing.M) {
	// we do this because a lot of static results depend on relative paths.
	core.Setwd("../../")
	core.DBInit()
	os.Exit(m.Run())
}

func MockAuth(a *assert.Assertions, tx *sql.Tx, mockUserProvider func(*sql.Tx) (*web.AuthManager, *web.Session, error)) (*web.AuthManager, *web.Session) {
	auth, session, err := mockUserProvider(tx)
	a.Nil(err)
	a.NotNil(auth)
	a.NotNil(session)
	return auth, session
}

func MockAdminLogin(tx *sql.Tx) (*web.AuthManager, *web.Session, error) {
	u, err := CreateTestAdminUser(tx)
	if err != nil {
		return nil, nil, err
	}
	return AuthTestUser(u, tx)
}

func MockModeratorLogin(tx *sql.Tx) (*web.AuthManager, *web.Session, error) {
	u, err := CreateTestModeratorUser(tx)
	if err != nil {
		return nil, nil, err
	}
	return AuthTestUser(u, tx)
}

func MockBannedLogin(tx *sql.Tx) (*web.AuthManager, *web.Session, error) {
	u, err := CreateTestModeratorUser(tx)
	if err != nil {
		return nil, nil, err
	}
	return AuthTestUser(u, tx)
}

func CreateTestAdminUser(tx *sql.Tx) (*model.User, error) {
	u := model.NewUser(fmt.Sprintf("__test_user_%s__", core.UUIDv4().ToShortString()))
	u.FirstName = "Test"
	u.LastName = "User"
	u.IsAdmin = true
	u.IsModerator = true
	err := model.DB().CreateInTx(u, tx)
	return u, err
}

func CreateTestModeratorUser(tx *sql.Tx) (*model.User, error) {
	u := model.NewUser(fmt.Sprintf("__test_user_%s__", core.UUIDv4().ToShortString()))
	u.FirstName = "Test"
	u.LastName = "User"
	u.IsAdmin = false
	u.IsModerator = true
	err := model.DB().CreateInTx(u, tx)
	return u, err
}

func CreateTestBannedUser(tx *sql.Tx) (*model.User, error) {
	u := model.NewUser(fmt.Sprintf("__test_user_%s__", core.UUIDv4().ToShortString()))
	u.FirstName = "Test"
	u.LastName = "User"
	u.IsAdmin = false
	u.IsModerator = false
	u.IsBanned = true
	err := model.DB().CreateInTx(u, tx)
	return u, err
}

func AuthTestUser(user *model.User, tx *sql.Tx) (*web.AuthManager, *web.Session, error) {
	auth := web.NewAuthManager()
	session, err := auth.Login(user.ID, nil)
	if err != nil {
		return nil, nil, err
	}
	webutil.SetUser(session, user)
	cachedSession, _ := auth.SessionCache().Get(session.SessionID)
	return auth, cachedSession, nil
}
