package oauth

// Result is the full result of an oauth exchange.
type Result struct {
	UniqueID string
	Profile  *Profile
	IDToken  *JWTPayload
	Response *Response
	State    *State
}
