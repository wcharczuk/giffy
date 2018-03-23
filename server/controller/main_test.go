package controller

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/blendlabs/go-assert"
	util "github.com/blendlabs/go-util"
	web "github.com/blendlabs/go-web"
	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/giffy/server/webutil"
)

func TestMain(m *testing.M) {
	// we do this because a lot of static results depend on relative paths.
	core.Setwd("../../")
	err := spiffy.OpenDefault(spiffy.NewFromConfig(spiffy.NewConfigFromEnv()))
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}

func MockAuth(a *assert.Assertions, tx *sql.Tx, mockUserProvider func(*sql.Tx) (*web.AuthManager, *web.Session, error)) (*web.AuthManager, *web.Session) {
	auth, r, err := mockUserProvider(tx)
	a.Nil(err)
	a.NotNil(auth)
	a.NotNil(r)
	return auth, r
}

func MockLogout(a *assert.Assertions, am *web.AuthManager, session *web.Session, tx *sql.Tx) {
	ctx := web.NewCtx(nil, &http.Request{}, nil, web.State{})
	web.WithTx(ctx, tx)
	ctx.SetSession(session)
	a.Nil(am.Logout(ctx))
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
	session, err := auth.Login(util.String.FromInt64(user.ID), nil)
	if err != nil {
		return nil, nil, err
	}
	webutil.SetUser(session, user)
	cachedSession := auth.SessionCache().Get(session.SessionID)
	return auth, cachedSession, nil
}
