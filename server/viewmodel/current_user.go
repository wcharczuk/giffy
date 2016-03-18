package viewmodel

import (
	"fmt"
	"net/http"

	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/core/web"
	"github.com/wcharczuk/giffy/server/model"
)

// CurrentUser is the response for the current user api service.
type CurrentUser struct {
	IsLoggedIn  bool   `json:"is_logged_in"`
	UserID      int64  `json:"-"`
	UserUUID    string `json:"user_uuid"`
	Username    string `json:"username"`
	IsAdmin     bool   `json:"is_admin"`
	IsModerator bool   `json:"is_moderator"`
	LoginURL    string `json:"login_url,ommitempty"`
}

// SetFromUser does things.
func (cu *CurrentUser) SetFromUser(u *model.User) {
	cu.IsLoggedIn = true
	cu.UserID = u.ID
	cu.UserUUID = u.UUID
	cu.Username = u.Username
	cu.IsAdmin = u.IsAdmin
	cu.IsModerator = u.IsModerator
}

// SetLoggedOut does things.
func (cu *CurrentUser) SetLoggedOut(ctx *web.HTTPContext) {
	cu.IsLoggedIn = false
	cu.LoginURL = fmt.Sprintf("https://accounts.google.com/o/oauth2/auth?response_type=code&client_id=%s&redirect_uri=%s&scope=https://www.googleapis.com/auth/userinfo.email%%20https://www.googleapis.com/auth/userinfo.profile", core.ConfigGoogleClientID(), OAuthRedirectURI(ctx.Request))
}

//OAuthRedirectURI formats a uri.
func OAuthRedirectURI(r *http.Request) string {
	return fmt.Sprintf("http://%s/oauth", core.ConfigHostname())
}
