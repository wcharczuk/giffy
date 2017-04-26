package external

import (
	"fmt"

	"github.com/wcharczuk/giffy/server/core"
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
func SlackAuthURL() string {
	return fmt.Sprintf("https://slack.com/oauth/authorize?client_id=%s&scope=%s&redirect_uri=%s", core.ConfigSlackClientID(), "identity.basic,commands,chat:write:user,chat:write:bot", SlackAuthReturnURL())
}

//SlackAuthReturnURL formats an oauth return uri.
func SlackAuthReturnURL() string {
	return fmt.Sprintf("http://%s/oauth/slack", core.ConfigHostname())
}

// FetchSlackProfile gets the slack user details for an access token.
func FetchSlackProfile(accessToken string) (*SlackProfile, error) {
	var auth SlackProfile
	err := NewRequest().
		AsPost().
		WithURL("https://slack.com/api/auth.test").
		WithPostData("token", accessToken).
		JSON(&auth)
	return &auth, err
}

// SlackOAuth completes the oauth 2.0 process with slack.
func SlackOAuth(code string) (*SlackOAuthResponse, error) {
	var oar SlackOAuthResponse

	err := NewRequest().
		AsPost().
		WithURL("https://slack.com/api/oauth.access").
		WithPostData("client_id", core.ConfigSlackClientID()).
		WithPostData("client_secret", core.ConfigSlackClientSecret()).
		WithPostData("redirect_uri", SlackAuthReturnURL()).
		WithPostData("code", code).
		JSON(&oar)

	return &oar, err
}
