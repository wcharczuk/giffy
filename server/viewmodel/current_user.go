package viewmodel

import (
	"fmt"

	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/core/external"
	"github.com/wcharczuk/giffy/server/model"
	"github.com/wcharczuk/go-web"
)

// CurrentUser is the response for the current user api service.
type CurrentUser struct {
	IsLoggedIn     bool   `json:"is_logged_in"`
	UUID           string `json:"uuid"`
	Username       string `json:"username"`
	IsAdmin        bool   `json:"is_admin"`
	IsModerator    bool   `json:"is_moderator"`
	IsBanned       bool   `json:"is_banned"`
	GoogleLoginURL string `json:"google_login_url,ommitempty"`
}

// SetFromUser does things.
func (cu *CurrentUser) SetFromUser(u *model.User) {
	cu.IsLoggedIn = true
	cu.UUID = u.UUID
	cu.Username = u.Username
	cu.IsAdmin = u.IsAdmin
	cu.IsModerator = u.IsModerator
	cu.IsBanned = u.IsBanned
}

// SetLoggedOut does things.
func (cu *CurrentUser) SetLoggedOut(ctx *web.HTTPContext) {
	cu.IsLoggedIn = false
	cu.GoogleLoginURL = fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/auth?response_type=code&client_id=%s&redirect_uri=%s&scope=https://www.googleapis.com/auth/userinfo.email%%20https://www.googleapis.com/auth/userinfo.profile",
		core.ConfigGoogleClientID(),
		external.GoogleAuthReturnURL(),
	)
}
