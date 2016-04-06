package external

import (
	"fmt"

	"github.com/wcharczuk/giffy/server/core"
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
	Status       string                    `json:"status"`
	AuthResponse FacebookOAuthAuthResponse `json:"authResponse"`
}

// FacebookProfile is a facebook profile.
type FacebookProfile struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// FacebookAuthURL is the link url to open the facebook auth dialogue.
func FacebookAuthURL() string {
	grantedScopes := "public_profile,email"
	return fmt.Sprintf("https://www.facebook.com/dialog/oauth?response_type=code&client_id=%s&redirect_uri=%s&granted_scopes=%s", core.ConfigFacebookClientID(), FacebookAuthReturnURL(), grantedScopes)
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
		FetchJSONToObject(&res)

	return &res, err
}

// FetchFacebookProfile fetches a facebook profile.
func FetchFacebookProfile(accessToken string) (*FacebookProfile, error) {
	var res FacebookProfile

	err := NewRequest().AsGet().
		WithScheme("https").
		WithHost("graph.facebook.com").
		WithPath("/v2.5/me").
		WithQueryString("access_token", accessToken).
		FetchJSONToObject(&res)

	return &res, err
}
