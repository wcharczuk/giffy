package external

import (
	"fmt"

	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/model"
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

// SlackUserProfile is a profile within a users.info response.
type SlackUserProfile struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	RealName  string `json:"real_name"`
	Email     string `json:"email"`
}

// SlackUserInfo is the response from the users.info service.
type SlackUserInfo struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Team    string            `json:"team"`
	Profile *SlackUserProfile `json:"profile"`
}

// AsUser returns the slack user as a model.User.
func (sui SlackUserInfo) AsUser() *model.User {
	return &model.User{
		Username:        fmt.Sprintf("%s @ %s", sui.Name, sui.Team),
		FirstName:       sui.Profile.FirstName,
		LastName:        sui.Profile.LastName,
		EmailAddress:    sui.Profile.Email,
		IsEmailVerified: true,
	}
}

// FetchSlackProfile returns the slack profile for an access token.
func FetchSlackProfile(accessToken string) (*SlackUserInfo, error) {
	var authTestResponse SlackAuthTestResponse
	err := core.NewExternalRequest().AsGet().WithURL("https://slack.com/api/auth.test").WithQueryString("token", accessToken).FetchJSONToObject(&authTestResponse)
	if err != nil {
		return nil, err
	}

	var userInfo SlackUserInfo

	err = core.NewExternalRequest().AsGet().
		WithURL("https://slack.com/api/users.info").
		WithQueryString("token", accessToken).
		WithQueryString("user", authTestResponse.UserID).
		FetchJSONToObject(&userInfo)

	if err != nil {
		return nil, err
	}

	userInfo.Team = authTestResponse.Team
	return &userInfo, nil
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

	err := core.NewExternalRequest().
		AsPost().
		WithURL("https://slack.com/oauth/authorize").
		WithPostData("client_id", core.ConfigSlackClientID()).
		WithPostData("client_secret", core.ConfigSlackClientSecret()).
		WithPostData("code", code).
		WithPostData("redirect_uri", SlackAuthReturnURL()).
		FetchJSONToObject(&oar)

	return &oar, err
}
