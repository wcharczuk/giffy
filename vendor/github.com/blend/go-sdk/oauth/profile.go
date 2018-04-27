package oauth

import (
	"strings"

	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/request"
)

// Profile is a profile with google.
type Profile struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Link          string `json:"link"`
	Gender        string `json:"gender"`
	Locale        string `json:"locale"`
	PictureURL    string `json:"picture"`
}

// Username returns the <username>@fqdn component
// of the email address.
func (p Profile) Username() string {
	if len(p.Email) == 0 {
		return ""
	}
	if !strings.Contains(p.Email, "@") {
		return p.Email
	}

	parts := strings.SplitN(p.Email, "@", 2)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

// FetchProfile gets a google profile for an access token.
func FetchProfile(accessToken string) (*Profile, error) {
	var profile Profile
	meta, err := request.New().AsGet().
		WithURL("https://www.googleapis.com/oauth2/v1/userinfo").
		WithQueryString("alt", "json").
		WithQueryString("access_token", accessToken).
		WithMockProvider(request.MockedResponseInjector).
		JSONWithMeta(&profile)

	if err != nil {
		return nil, err
	}
	if meta.StatusCode > 299 {
		return nil, exception.NewFromErr(ErrGoogleResponseStatus).WithMessagef("status code: %d", meta.StatusCode)
	}
	return &profile, err
}
