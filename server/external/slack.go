package external

import (
	"fmt"

	"github.com/wcharczuk/giffy/server/config"
)

// SlackOAuthResponse is the response from the second phase of slack oauth 2.0
type SlackOAuthResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`

	AccessToken string `json:"access_token"`
	Scope       string `json:"scopes"`

	UserID   string `json:"user_id"`
	TeamID   string `json:"team_id"`
	TeamName string `json:"team_name"`
}

// SlackProfile is the response from the auth.test service.
type SlackProfile struct {
	OK     bool   `json:"ok"`
	URL    string `json:"url"`
	User   string `json:"user"`
	Team   string `json:"team"`
	UserID string `json:"user_id"`
	TeamID string `json:"team_id"`
}

// SlackAuthURL is the url to start the OAuth 2.0 process with slack.
func SlackAuthURL(cfg *config.Giffy) string {
	return fmt.Sprintf("https://slack.com/oauth/authorize?client_id=%s&scope=%s&redirect_uri=%s", cfg.SlackClientID, "identity.basic,commands,chat:write:user,chat:write:bot", SlackAuthReturnURL(cfg))
}

//SlackAuthReturnURL formats an oauth return uri.
func SlackAuthReturnURL(cfg *config.Giffy) string {
	return fmt.Sprintf("http://%s/oauth/slack", cfg.Web.GetBaseURL())
}

// FetchSlackProfile gets the slack user details for an access token.
func FetchSlackProfile(accessToken string, cfg *config.Giffy) (*SlackProfile, error) {
	var auth SlackProfile
	err := NewRequest().
		AsPost().
		WithURL("https://slack.com/api/auth.test").
		WithPostData("token", accessToken).
		JSON(&auth)
	return &auth, err
}

// SlackOAuth completes the oauth 2.0 process with slack.
func SlackOAuth(code string, cfg *config.Giffy) (*SlackOAuthResponse, error) {
	var oar SlackOAuthResponse

	err := NewRequest().
		AsPost().
		WithURL("https://slack.com/api/oauth.access").
		WithPostData("client_id", cfg.SlackClientID).
		WithPostData("client_secret", cfg.SlackClientSecret).
		WithPostData("redirect_uri", SlackAuthReturnURL(cfg)).
		WithPostData("code", code).
		JSON(&oar)

	return &oar, err
}
