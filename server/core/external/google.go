package external

import (
	"github.com/wcharczuk/giffy/server/core"
	"github.com/wcharczuk/giffy/server/model"
)

// GoogleOAuthResponse is a response from google oauth.
type GoogleOAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	IDToken     string `json:"id_token"`
}

// GoogleProfile is a profile with google.
type GoogleProfile struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Link          string `json:"link"`
	Gender        string `json:"male"`
	Locale        string `json:"locale"`
	PictureURL    string `json:"picture"`
}

// User returns a profile as a user.
func (gp GoogleProfile) User() *model.User {
	user := model.NewUser(gp.Email)
	user.EmailAddress = gp.Email
	user.IsEmailVerified = gp.VerifiedEmail
	user.FirstName = gp.GivenName
	user.LastName = gp.FamilyName
	return user
}

// FetchGoogleProfile gets a google proflile for an access token.
func FetchGoogleProfile(accessToken string) (*GoogleProfile, error) {
	var profile GoogleProfile
	err := core.NewExternalRequest().AsGet().
		WithURL("https://www.googleapis.com/oauth2/v1/userinfo").
		WithQueryString("alt", "json").
		WithQueryString("access_token", accessToken).
		FetchJSONToObject(&profile)
	return &profile, err
}
