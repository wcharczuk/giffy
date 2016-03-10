package viewmodel

import "fmt"

// Login is the viewmodel for the login page.
type Login struct {
	Title            string
	ClientID         string
	Secret           string
	OAUTHRedirectURI string
}

// OAuthURL returns the OAUTH url.
func (l Login) OAuthURL() string {
	return fmt.Sprintf("https://accounts.google.com/o/oauth2/auth?response_type=code&client_id=%s&redirect_uri=%s&scope=https://www.googleapis.com/auth/userinfo.email%%20https://www.googleapis.com/auth/userinfo.profile", l.ClientID, l.OAUTHRedirectURI)
}
