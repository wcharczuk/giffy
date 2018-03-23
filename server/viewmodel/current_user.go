package viewmodel

import (
	"github.com/wcharczuk/giffy/server/model"
)

//NewCurrentUser creates a new current user view model.
func NewCurrentUser(u *model.User) *CurrentUser {
	return &CurrentUser{
		UUID:        u.UUID,
		Username:    u.Username,
		IsAdmin:     u.IsAdmin,
		IsModerator: u.IsModerator,
		IsBanned:    u.IsBanned,
	}
}

// CurrentUser is the response for the current user api service.
type CurrentUser struct {
	// Standard user fields
	UUID        string `json:"uuid"`
	Username    string `json:"username"`
	IsAdmin     bool   `json:"is_admin"`
	IsModerator bool   `json:"is_moderator"`
	IsBanned    bool   `json:"is_banned"`
	// Meta Fields used for the login header
	IsLoggedIn     bool   `json:"is_logged_in"`
	GoogleLoginURL string `json:"google_login_url,omitempty"`
	SlackLoginURL  string `json:"slack_login_url,omitempty"`
}

// SetFromUser sets a current user from a model user.
func (cu *CurrentUser) SetFromUser(u *model.User) {
	cu.UUID = u.UUID
	cu.Username = u.Username
	cu.IsAdmin = u.IsAdmin
	cu.IsModerator = u.IsModerator
	cu.IsBanned = u.IsBanned
}
