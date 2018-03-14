package viewmodel

import (
	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/external"
	"github.com/wcharczuk/giffy/server/model"
)

// CurrentUser is the response for the current user api service.
type CurrentUser struct {
	IsLoggedIn       bool   `json:"is_logged_in"`
	UUID             string `json:"uuid"`
	Username         string `json:"username"`
	IsAdmin          bool   `json:"is_admin"`
	IsModerator      bool   `json:"is_moderator"`
	IsBanned         bool   `json:"is_banned"`
	FacebookLoginURL string `json:"facebook_login_url,omitempty"`
	GoogleLoginURL   string `json:"google_login_url,omitempty"`
	SlackLoginURL    string `json:"slack_login_url,omitempty"`
}

// SetFromUser does things.
func (cu *CurrentUser) SetFromUser(u *model.User, cfg *config.Giffy) {
	cu.IsLoggedIn = true
	cu.UUID = u.UUID
	cu.Username = u.Username
	cu.IsAdmin = u.IsAdmin
	cu.IsModerator = u.IsModerator
	cu.IsBanned = u.IsBanned
	cu.SlackLoginURL = external.SlackAuthURL(cfg)
}

// SetLoggedOut does things.
func (cu *CurrentUser) SetLoggedOut(cfg *config.Giffy) {
	cu.IsLoggedIn = false
	cu.FacebookLoginURL = external.FacebookAuthURL(cfg)
	cu.GoogleLoginURL = external.GoogleAuthURL(cfg)
	cu.SlackLoginURL = external.SlackAuthURL(cfg)
}
