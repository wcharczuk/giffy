package external

import (
	"fmt"

	"github.com/wcharczuk/giffy/server/core"
)

// SlackOAuthResponse is the response from the second phase of slack oauth 2.0
type SlackOAuthResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scopes"`
}

// SlackAuthTestResponse is the response from the auth.test service.
type SlackAuthTestResponse struct {
	OK     bool   `json:"ok"`
	URL    string `json:"url"`
	User   string `json:"user"`
	Team   string `json:"team"`
	UserID string `json:"user_id"`
	TeamID string `json:"team_id"`
}

// SlackLoginURL is the url to start the OAuth 2.0 process with slack.
func SlackLoginURL(scope string) string {
	return fmt.Sprintf("https://slack.com/oauth/authorize?client_id=%s&scope=%s&redirect_uri=%s", core.ConfigSlackClientID(), scope, SlackAuthReturnURL())
}

//SlackAuthReturnURL formats an oauth return uri.
func SlackAuthReturnURL() string {
	return fmt.Sprintf("http://%s/oauth/slack", core.ConfigHostname())
}

// SlackOAuth completes the oauth 2.0 process with slack.
func SlackOAuth(code string) (*SlackOAuthResponse, error) {
	var oar SlackOAuthResponse

	err := NewRequest().
		AsPost().
		WithURL("https://slack.com/api/oauth.access").
		WithPostData("client_id", core.ConfigSlackClientID()).
		WithPostData("client_secret", core.ConfigSlackClientSecret()).
		WithPostData("code", code).
		FetchJSONToObject(&oar)

	return &oar, err
}
