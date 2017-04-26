package external

import (
	"fmt"

	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/model"
)

// FacebookOAuthAuthResponse is the inner body of the OAuthResponse.
type FacebookOAuthAuthResponse struct {
	AccessToken   string `json:"accessToken"`
	ExpiresIn     string `json:"expiresIn"`
	SignedRequest string `json:"signedRequest"`
	UserID        string `json:"userID"`
}

// FacebookOAuthResponse is the response format for the OAuth token exchange step.
type FacebookOAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// FacebookProfile is a facebook profile.
type FacebookProfile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

// AsUser returns the profile as a model.User.
func (fp FacebookProfile) AsUser() *model.User {
	user := model.NewUser(fp.Email)
	user.EmailAddress = fp.Email
	user.IsEmailVerified = true
	user.FirstName = fp.FirstName
	user.LastName = fp.LastName
	return user
}

// FacebookAuthURL is the link url to open the facebook auth dialogue.
func FacebookAuthURL() string {
	grantedScopes := "public_profile,email"
	return fmt.Sprintf("https://www.facebook.com/dialog/oauth?response_type=code&client_id=%s&redirect_uri=%s&scope=%s", core.ConfigFacebookClientID(), FacebookAuthReturnURL(), grantedScopes)
}

// FacebookAuthReturnURL formats an oauth return uri.
func FacebookAuthReturnURL() string {
	return fmt.Sprintf("%s/oauth/facebook", core.ConfigURL())
}

// FacebookOAuth exchanges an auth code for a token.
func FacebookOAuth(code string) (*FacebookOAuthResponse, error) {
	var res FacebookOAuthResponse
	err := NewRequest().
		AsPost().
		WithScheme("https").
		WithHost("graph.facebook.com").
		WithPath("v2.3/oauth/access_token").
		WithPostData("client_id", core.ConfigFacebookClientID()).
		WithPostData("client_secret", core.ConfigFacebookClientSecret()).
		WithPostData("redirect_uri", FacebookAuthReturnURL()).
		WithPostData("code", code).
		JSON(&res)

	return &res, err
}

// FetchFacebookProfile fetches a facebook profile.
func FetchFacebookProfile(accessToken string) (*FacebookProfile, error) {
	var res FacebookProfile

	fields := "email,first_name,last_name"

	err := NewRequest().AsGet().
		WithScheme("https").
		WithHost("graph.facebook.com").
		WithPath("/v2.5/me").
		WithQueryString("access_token", accessToken).
		WithQueryString("fields", fields).
		JSON(&res)

	return &res, err
}
