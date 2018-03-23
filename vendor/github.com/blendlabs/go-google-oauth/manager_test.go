package google

import (
	"net/url"
	"testing"
	"time"

	assert "github.com/blendlabs/go-assert"
	util "github.com/blendlabs/go-util"
)

func TestNewFromConfig(t *testing.T) {
	assert := assert.New(t)

	m := NewFromConfig(&Config{
		SkipDomainValidation: false,
		RedirectURI:          "https://app.com/oauth/google",
		ValidDomains:         []string{"foo.com", "bar.com"},
		ClientID:             "foo_client",
		ClientSecret:         "bar_secret",
	})

	assert.Equal("https://app.com/oauth/google", m.RedirectURI())
	assert.Len(m.ValidDomains(), 2)
	assert.Equal("foo_client", m.ClientID())
	assert.Equal("bar_secret", m.ClientSecret())
}

func TestManagerValidDomains(t *testing.T) {
	assert := assert.New(t)

	assert.Empty(New().ValidDomains())

	domains := New().WithHostedDomain("foo.com").ValidDomains()
	assert.NotEmpty(domains)
	assert.Len(domains, 1)
	assert.Equal("foo.com", domains[0])

	domains = New().WithValidDomains("foo.com", "bar.com").ValidDomains()
	assert.NotEmpty(domains)
	assert.Len(domains, 2)
	assert.Any(domains, func(v interface{}) bool { return v.(string) == "foo.com" })
	assert.Any(domains, func(v interface{}) bool { return v.(string) == "bar.com" })

	domains = New().WithHostedDomain("buzz.com").WithValidDomains("foo.com", "bar.com").ValidDomains()
	assert.NotEmpty(domains)
	assert.Len(domains, 3)
	assert.Any(domains, func(v interface{}) bool { return v.(string) == "foo.com" })
	assert.Any(domains, func(v interface{}) bool { return v.(string) == "bar.com" })
	assert.Any(domains, func(v interface{}) bool { return v.(string) == "buzz.com" })

	domains = New().WithHostedDomain("bar.com").WithValidDomains("foo.com", "bar.com").ValidDomains()
	assert.NotEmpty(domains)
	assert.Len(domains, 2)
	assert.Any(domains, func(v interface{}) bool { return v.(string) == "foo.com" })
	assert.Any(domains, func(v interface{}) bool { return v.(string) == "bar.com" })
}

func TestManagerOAuthURL(t *testing.T) {
	assert := assert.New(t)

	m := New().
		WithClientID("test_client_id").
		WithHostedDomain("test.blend.com").
		WithRedirectURI("https://local.shortcut-service.centrio.com/oauth/google")

	_, err := m.OAuthURL()
	assert.NotNil(err, "we require a secret")
}

func TestManagerOAuthURLSecurity(t *testing.T) {
	assert := assert.New(t)

	m := New().
		WithSecret(util.Crypto.MustCreateKey(32)).
		WithClientID("test_client_id").
		WithRedirectURI("https://local.shortcut-service.centrio.com/oauth/google")

	urlFragment, err := m.OAuthURL()
	assert.Nil(err)

	u, err := url.Parse(urlFragment)
	assert.Nil(err)
	assert.NotEmpty(u.Query().Get("state"))
	assert.Equal("test_client_id", u.Query().Get("client_id"))

	state := u.Query().Get("state")
	deserialized, err := DeserializeOAuthState(state)
	assert.Nil(err)
	assert.Nil(m.ValidateOAuthState(deserialized))
}

func TestManagerOAuthURLRedirect(t *testing.T) {
	assert := assert.New(t)

	m := New().
		WithSecret(util.Crypto.MustCreateKey(32)).
		WithClientID("test_client_id").
		WithRedirectURI("https://local.shortcut-service.centrio.com/oauth/google")

	urlFragment, err := m.OAuthURL("bar_foo")
	assert.Nil(err)

	u, err := url.Parse(urlFragment)
	assert.Nil(err)
	assert.NotEmpty(u.Query().Get("state"))

	state := u.Query().Get("state")
	deserialized, err := DeserializeOAuthState(state)
	assert.Nil(err)
	assert.Nil(m.ValidateOAuthState(deserialized))
	assert.Equal("bar_foo", deserialized.RedirectURL)
}

func TestManagerValidateProfile(t *testing.T) {
	assert := assert.New(t)

	blender := &Profile{
		Email: "bailey@blend.com",
	}

	personal := &Profile{
		Email: "bailey@gmail.com",
	}

	suffixMatch := &Profile{
		Email: "bailey@sailblend.com",
	}

	prefixMatch := &Profile{
		Email: "bailey@blend.com.au",
	}

	empty := New()
	assert.NotNil(empty.ValidateProfile(blender), "we should error if no domains are configured with defaults")

	unvalidated := New().WithSkipDomainValidation(true)
	assert.Nil(unvalidated.ValidateProfile(blender), "we should not error if skip validation is true")
	assert.Nil(unvalidated.ValidateProfile(personal), "we should not error if skip validation is true")
	assert.Nil(unvalidated.ValidateProfile(suffixMatch), "we should not error if skip validation is true")
	assert.Nil(unvalidated.ValidateProfile(prefixMatch), "we should not error if skip validation is true")

	validated := New().WithValidDomains("blend.com")
	assert.Nil(validated.ValidateProfile(blender), "we should pass for @blend.com")
	assert.NotNil(validated.ValidateProfile(personal), "we fail for non-@blend.com emails")
	assert.NotNil(validated.ValidateProfile(suffixMatch), "we fail for non-@blend.com emails")
	assert.NotNil(validated.ValidateProfile(prefixMatch), "we fail for non-@blend.com emails")

	hostedOnly := New().WithHostedDomain("blend.com")
	assert.Nil(hostedOnly.ValidateProfile(blender), "we should pass for @blend.com")
	assert.NotNil(hostedOnly.ValidateProfile(personal), "we fail for non-@blend.com emails")
	assert.NotNil(hostedOnly.ValidateProfile(suffixMatch), "we fail for non-@blend.com emails")
	assert.NotNil(hostedOnly.ValidateProfile(prefixMatch), "we fail for non-@blend.com emails")

	validatedPrefixed := New().WithValidDomains("@blend.com")
	assert.Nil(validatedPrefixed.ValidateProfile(blender), "we should pass for @blend.com")
	assert.NotNil(validatedPrefixed.ValidateProfile(personal), "we fail for non-@blend.com emails")
	assert.NotNil(validatedPrefixed.ValidateProfile(suffixMatch), "we fail for non-@blend.com emails")
	assert.NotNil(validatedPrefixed.ValidateProfile(prefixMatch), "we fail for non-@blend.com emails")

	multi := New().WithValidDomains("@blend.com", "sailblend.com")
	assert.Nil(multi.ValidateProfile(blender), "we should pass for @blend.com")
	assert.NotNil(multi.ValidateProfile(personal), "we fail for non-@blend.com emails")
	assert.Nil(multi.ValidateProfile(suffixMatch), "we should pass for @sailblend.com as it was configured too")
	assert.NotNil(multi.ValidateProfile(prefixMatch), "we fail for non-@blend.com emails")
}

func TestManagerValidateOAuthState(t *testing.T) {
	assert := assert.New(t)

	m := New().WithSecret(util.Crypto.MustCreateKey(32))
	assert.Nil(m.ValidateOAuthState(m.CreateState()))
}

func TestManagerState(t *testing.T) {
	assert := assert.New(t)

	m := New().WithSecret(util.Crypto.MustCreateKey(32))
	serialized, err := m.State()
	assert.Nil(err)
	assert.NotEmpty(serialized, "at the baseline we should have the anti-forgery pair")

	state := m.CreateState("")
	assert.Empty(state.RedirectURL, "if we provide an empty string, it still shouldn't add a redirect url")

	serialized, err = m.State("foo")
	assert.Nil(err)
	assert.NotEmpty(serialized, "if we provide a valid string it should create a serialized state object")

	deserialized, err := DeserializeOAuthState(serialized)
	assert.Nil(err)
	assert.Nil(m.ValidateOAuthState(deserialized))
	assert.Equal("foo", deserialized.RedirectURL)

	m.WithSecret(util.Crypto.MustCreateKey(32))
	serialized, err = m.State("foo")
	assert.Nil(err)
	assert.NotEmpty(serialized, "if we provide a valid string it should create a serialized state object")

	deserialized, err = DeserializeOAuthState(serialized)
	assert.Nil(err)
	assert.Nil(m.ValidateOAuthState(deserialized))
	assert.Equal("foo", deserialized.RedirectURL)

	serialized, err = m.State()
	assert.Nil(err)
	assert.NotEmpty(serialized, "state should be issued if we have a secret, even if we don't provide a redirect")

	deserialized, err = DeserializeOAuthState(serialized)
	assert.Nil(err)
	assert.Nil(m.ValidateOAuthState(deserialized))
	assert.Empty(deserialized.RedirectURL)
}

func TestManagerValidateNonce(t *testing.T) {
	assert := assert.New(t)

	m := New().WithSecret(util.Crypto.MustCreateKey(32))
	nonce, err := m.CreateNonce(time.Now().UTC())
	assert.Nil(err)
	assert.Nil(m.ValidateNonce(nonce))

	nonce, err = m.CreateNonce(time.Now().UTC().Add(-time.Minute))
	assert.Nil(err)
	assert.NotNil(m.ValidateNonce(nonce))
}
