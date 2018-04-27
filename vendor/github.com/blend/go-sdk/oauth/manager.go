package oauth

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/blend/go-sdk/util"
	"github.com/blend/go-sdk/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// New returns a new manager.
// By default it will error if you try and validate a profile.
// You must either enable `SkipDomainvalidation` or provide valid domains.
func New() *Manager {
	return &Manager{}
}

// Must is a helper for handling NewFromEnv() and NewFromConfig().
func Must(m *Manager, err error) *Manager {
	if err != nil {
		panic(err)
	}
	return m
}

// NewFromEnv returns a new manager from the environment.
func NewFromEnv() (*Manager, error) {
	return NewFromConfig(NewConfigFromEnv())
}

// NewFromConfig returns a new oauth manager from a config.
func NewFromConfig(cfg *Config) (*Manager, error) {
	secret, err := cfg.GetSecret()
	if err != nil {
		return nil, err
	}
	return &Manager{
		secret:       secret,
		redirectURI:  cfg.GetRedirectURI(),
		clientID:     cfg.GetClientID(),
		clientSecret: cfg.GetClientSecret(),
		hostedDomain: cfg.GetHostedDomain(),
	}, nil
}

// Manager is the oauth manager.
type Manager struct {
	secret       []byte
	redirectURI  string
	hostedDomain string
	clientID     string
	clientSecret string
}

func (m *Manager) conf() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     m.clientID,
		ClientSecret: m.clientSecret,
		RedirectURL:  m.redirectURI,
		Scopes: []string{
			"openid",
			"email",
			"profile",
		},
		Endpoint: google.Endpoint,
	}
}

// OAuthURL is the auth url for google with a given clientID.
// This is typically the link that a user will click on to start the auth process.
func (m *Manager) OAuthURL(redirect ...string) (string, error) {
	state, err := SerializeState(m.CreateState(redirect...))
	if err != nil {
		return "", err
	}

	var opts []oauth2.AuthCodeOption
	if len(m.hostedDomain) > 0 {
		opts = append(opts, oauth2.SetAuthURLParam("hd", m.hostedDomain))
	}
	return m.conf().AuthCodeURL(state, opts...), nil
}

// Finish processes the returned code, exchanging for an access token, and fetches the user profile.
func (m *Manager) Finish(r *http.Request) (*Result, error) {
	var result Result
	var err error

	// grab the code off the request.
	code := r.URL.Query().Get("code")
	if len(code) == 0 {
		return nil, ErrCodeMissing
	}

	// fetch the state
	state := r.URL.Query().Get("state")
	if len(state) > 0 {
		deserialized, err := DeserializeState(state)
		if err != nil {
			return nil, err
		}
		result.State = deserialized
	}

	err = m.ValidateState(result.State)
	if err != nil {
		return nil, err
	}

	// Handle the exchange code to initiate a transport.
	tok, err := m.conf().Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatal(err)
	}
	result.Response.AccessToken = tok.AccessToken
	result.Response.TokenType = tok.TokenType
	result.Response.RefreshToken = tok.RefreshToken
	result.Response.Expiry = tok.Expiry

	prof, err := FetchProfile(tok.AccessToken)
	if err != nil {
		return nil, err
	}
	result.Profile = prof
	return &result, nil
}

// CreateState creates auth state.
func (m *Manager) CreateState(redirect ...string) *State {
	var state State
	if len(m.secret) > 0 {
		state.Token = uuid.V4().String()
		state.SecureToken = m.hash(state.Token)
	}

	if len(redirect) > 0 && len(redirect[0]) > 0 {
		state.RedirectURL = redirect[0]
	}

	return &state
}

// Validation Helpers

// ValidateState validates oauth state.
func (m *Manager) ValidateState(state *State) error {
	if state == nil {
		return nil
	}
	if len(m.secret) > 0 {
		expected := m.hash(state.Token)
		actual := state.SecureToken
		if !hmac.Equal([]byte(expected), []byte(actual)) {
			return ErrInvalidAntiforgeryToken
		}
	}
	return nil
}

// ValidateProfile validates a profile.
func (m *Manager) ValidateProfile(p *Profile) error {
	if len(m.HostedDomain()) == 0 {
		return nil
	}

	workingDomain := m.hostedDomain
	if !strings.HasPrefix(workingDomain, "@") {
		workingDomain = fmt.Sprintf("@%s", workingDomain)
	}
	if !util.String.HasSuffixCaseInsensitive(p.Email, workingDomain) {
		return ErrInvalidHostedDomain
	}
	return nil
}

// --------------------------------------------------------------------------------
// Properties
// --------------------------------------------------------------------------------

// WithSecret sets the secret used to create state tokens.
func (m *Manager) WithSecret(secret []byte) *Manager {
	m.secret = secret
	return m
}

// Secret returns a property
func (m *Manager) Secret() []byte {
	return m.secret
}

// WithRedirectURI sets the return url.
func (m *Manager) WithRedirectURI(redirectURI string) *Manager {
	m.redirectURI = redirectURI
	return m
}

// RedirectURI returns a property.
func (m *Manager) RedirectURI() string {
	return m.redirectURI
}

// WithHostedDomain returns the hosted domain.
func (m *Manager) WithHostedDomain(hostedDomain string) *Manager {
	m.hostedDomain = hostedDomain
	return m
}

// HostedDomain returns the hosted domain.
func (m *Manager) HostedDomain() string {
	return m.hostedDomain
}

// WithClientID sets the client id.
func (m *Manager) WithClientID(clientID string) *Manager {
	m.clientID = clientID
	return m
}

// ClientID returns a property.
func (m *Manager) ClientID() string {
	return m.clientID
}

// WithClientSecret sets the client id.
func (m *Manager) WithClientSecret(clientSecret string) *Manager {
	m.clientSecret = clientSecret
	return m
}

// ClientSecret returns a client secret.
func (m *Manager) ClientSecret() string {
	return m.clientSecret
}

// --------------------------------------------------------------------------------
// internal helpers
// --------------------------------------------------------------------------------

func (m *Manager) hash(plaintext string) string {
	return base64.URLEncoding.EncodeToString(m.hmac([]byte(plaintext)))
}

// hmac hashes data with the given key.
func (m *Manager) hmac(plainText []byte) []byte {
	mac := hmac.New(sha512.New, m.secret)
	mac.Write([]byte(plainText))
	return mac.Sum(nil)
}
