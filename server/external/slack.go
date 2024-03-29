package external

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/blend/go-sdk/r2"
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

var defaultScopes = []string{
	"commands",
	"chat:write.public",
	"chat:write.customize",
	"chat:write",
}

// SlackAuthURL is the url to start the OAuth 2.0 process with slack.
func SlackAuthURL(cfg *config.Giffy) string {
	return fmt.Sprintf("https://slack.com/oauth/authorize?client_id=%s&scope=%s&redirect_uri=%s", cfg.SlackClientID, strings.Join(defaultScopes, ","), SlackAuthReturnURL(cfg))
}

//SlackAuthReturnURL formats an oauth return uri.
func SlackAuthReturnURL(cfg *config.Giffy) string {
	if cfg.SlackAuthReturnURL != "" {
		return cfg.SlackAuthReturnURL
	}
	if cfg.Web.BaseURL != "" {
		baseParsed, _ := url.Parse(cfg.Web.BaseURL)
		baseParsed.Path = "/oauth/slack"
		return baseParsed.String()
	}
	return "https://www.gifffy.com/oauth/slack"
}

// FetchSlackProfile gets the slack user details for an access token.
func FetchSlackProfile(accessToken string, cfg *config.Giffy) (*SlackProfile, error) {
	var auth SlackProfile
	_, err := r2.New("https://slack.com/api/auth.test",
		r2.OptPost(),
		r2.OptPostFormValue("token", accessToken)).JSON(&auth)
	return &auth, err
}

// SlackOAuth completes the oauth 2.0 process with slack.
func SlackOAuth(code string, cfg *config.Giffy) (*SlackOAuthResponse, error) {
	var oar SlackOAuthResponse
	_, err := r2.New("https://slack.com/api/oauth.access",
		r2.OptPost(),
		r2.OptPostFormValue("client_id", cfg.SlackClientID),
		r2.OptPostFormValue("client_secret", cfg.SlackClientSecret),
		r2.OptPostFormValue("redirect_uri", SlackAuthReturnURL(cfg)),
		r2.OptPostFormValue("code", code),
	).JSON(&oar)
	return &oar, err
}
