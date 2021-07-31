package controller

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/testutil"
	"github.com/blend/go-sdk/uuid"
	"github.com/blend/go-sdk/web"
	"github.com/blend/go-sdk/webutil"

	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/model"
)

func TestMain(m *testing.M) {
	cfg := config.MustNewFromEnv()
	testutil.New(m,
		testutil.OptLog(logger.All()),
		testutil.OptWithDefaultDB(),
		testutil.OptBefore(
			func(ctx context.Context) error {
				// we do this because a lot of static results depend on relative paths.
				return core.Setwd("../..")
			},
			func(ctx context.Context) error {
				return model.Schema(cfg).Apply(ctx, testutil.DefaultDB())
			},
			func(ctx context.Context) error {
				return model.Migrations(cfg).Apply(ctx, testutil.DefaultDB())
			},
		),
	).Run()
}

func MockAuth(a *assert.Assertions, mgr *model.Manager, mockUserProvider func(*model.Manager) (*web.AuthManager, *web.Session, error)) (*web.AuthManager, *web.Session) {
	auth, r, err := mockUserProvider(mgr)
	a.Nil(err)
	a.NotNil(auth)
	a.NotNil(r)
	return auth, r
}

func MockLogout(a *assert.Assertions, mgr *model.Manager, am *web.AuthManager, session *web.Session) {
	ctx := web.NewCtx(nil, &http.Request{})
	ctx.Session = session
	a.Nil(am.Logout(ctx))
}

func MockAdminLogin(mgr *model.Manager) (*web.AuthManager, *web.Session, error) {
	u, err := CreateTestAdminUser(mgr)
	if err != nil {
		return nil, nil, err
	}
	return AuthTestUser(u, mgr)
}

func MockModeratorLogin(mgr *model.Manager) (*web.AuthManager, *web.Session, error) {
	u, err := CreateTestModeratorUser(mgr)
	if err != nil {
		return nil, nil, err
	}
	return AuthTestUser(u, mgr)
}

func MockBannedLogin(mgr *model.Manager) (*web.AuthManager, *web.Session, error) {
	u, err := CreateTestModeratorUser(mgr)
	if err != nil {
		return nil, nil, err
	}
	return AuthTestUser(u, mgr)
}

func CreateTestAdminUser(mgr *model.Manager) (*model.User, error) {
	u := model.NewUser(fmt.Sprintf("__test_user_%s__", uuid.V4().String()))
	u.FirstName = "Test"
	u.LastName = "User"
	u.IsAdmin = true
	u.IsModerator = true
	err := mgr.Invoke(context.Background()).Create(u)
	return u, err
}

func CreateTestModeratorUser(mgr *model.Manager) (*model.User, error) {
	u := model.NewUser(fmt.Sprintf("__test_user_%s__", uuid.V4().String()))
	u.FirstName = "Test"
	u.LastName = "User"
	u.IsAdmin = false
	u.IsModerator = true
	err := mgr.Invoke(context.Background()).Create(u)
	return u, err
}

func CreateTestBannedUser(mgr *model.Manager) (*model.User, error) {
	u := model.NewUser(fmt.Sprintf("__test_user_%s__", uuid.V4().String()))
	u.FirstName = "Test"
	u.LastName = "User"
	u.IsAdmin = false
	u.IsModerator = false
	u.IsBanned = true
	err := mgr.Invoke(context.Background()).Create(u)
	return u, err
}

func AuthTestUser(user *model.User, mgr *model.Manager) (*web.AuthManager, *web.Session, error) {
	cache := web.NewLocalSessionCache()
	auth, _ := web.NewLocalAuthManagerFromCache(cache)
	session, err := auth.Login(strconv.FormatInt(user.ID, 10), web.NewCtx(webutil.NewMockResponse(ioutil.Discard), &http.Request{Host: "localhost"}))
	if err != nil {
		return nil, nil, err
	}
	SetUser(session, user)
	cachedSession := cache.Get(session.SessionID)
	return &auth, cachedSession, nil
}
