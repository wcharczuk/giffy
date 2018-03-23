package google

// OAuthResult is the full result of an oauth exchange.
type OAuthResult struct {
	UniqueID string
	Profile  *Profile
	IDToken  *JWTPayload
	Response *OAuthResponse
	State    *OAuthState
}
