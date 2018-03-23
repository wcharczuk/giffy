package google

import (
	"testing"

	assert "github.com/blendlabs/go-assert"
	util "github.com/blendlabs/go-util"
)

func TestSerializeOAuthState(t *testing.T) {
	assert := assert.New(t)

	state := OAuthState{
		RedirectURL: "https://foo.com/bar",
		Token:       util.String.RandomLetters(32),
		Secure:      util.String.RandomLetters(64),
	}

	contents, err := SerializeOAuthState(&state)
	assert.Nil(err)
	assert.NotEmpty(contents)

	deserialized, err := DeserializeOAuthState(contents)
	assert.Nil(err)
	assert.NotNil(deserialized)
	assert.Equal(state.RedirectURL, deserialized.RedirectURL)
	assert.Equal(state.Secure, deserialized.Secure)
}
