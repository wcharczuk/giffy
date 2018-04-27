package oauth

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"

	"github.com/blend/go-sdk/exception"
)

// DeserializeState deserializes the oauth state.
func DeserializeState(raw string) (*State, error) {
	corpus, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, exception.Wrap(err)
	}
	buffer := bytes.NewBuffer(corpus)
	var state State
	if err := gob.NewDecoder(buffer).Decode(&state); err != nil {
		return nil, exception.Wrap(err)
	}

	return &state, nil
}

// SerializeState serializes the oauth state.
func SerializeState(state *State) (output string, err error) {
	buffer := bytes.NewBuffer(nil)
	err = gob.NewEncoder(buffer).Encode(state)
	if err != nil {
		return
	}
	output = base64.StdEncoding.EncodeToString(buffer.Bytes())
	return
}

// State is the oauth state.
type State struct {
	// Token is a plaintext random token.
	Token string
	// SecureToken is the hashed version of the token.
	// If a key is set, it validates that our app created the oauth state.
	SecureToken string
	// RedirectURL is the redirect url.
	RedirectURL string
}
