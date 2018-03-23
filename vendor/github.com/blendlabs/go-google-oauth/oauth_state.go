package google

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"

	exception "github.com/blendlabs/go-exception"
)

// DeserializeOAuthState deserializes the oauth state.
func DeserializeOAuthState(raw string) (*OAuthState, error) {
	corpus, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, exception.Wrap(err)
	}
	buffer := bytes.NewBuffer(corpus)
	var state OAuthState
	if err := gob.NewDecoder(buffer).Decode(&state); err != nil {
		return nil, exception.Wrap(err)
	}

	return &state, nil
}

// SerializeOAuthState serializes the oauth state.
func SerializeOAuthState(state *OAuthState) (output string, err error) {
	buffer := bytes.NewBuffer(nil)
	err = gob.NewEncoder(buffer).Encode(state)
	if err != nil {
		return
	}
	output = base64.StdEncoding.EncodeToString(buffer.Bytes())
	return
}

// OAuthState is the oauth state.
type OAuthState struct {
	Token       string
	Secure      string
	RedirectURL string
}
