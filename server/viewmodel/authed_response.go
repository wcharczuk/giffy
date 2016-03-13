package viewmodel

// AuthedResponse is an api response for login.
type AuthedResponse struct {
	SessionID string `json:"giffy_auth"`
}
