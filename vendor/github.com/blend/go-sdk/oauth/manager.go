package oauth

import (
	"crypto/hmac"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/request"
	"github.com/blend/go-sdk/util"
	"github.com/blend/go-sdk/uuid"
)

// New returns a new manager.
// By default it will error if you try and validate a profile.
// You must either enable `SkipDomainvalidation` or provide valid domains.
func New() *Manager {
	return &Manager{
		secret:       util.Crypto.MustCreateKey(32),
		nonceTimeout: DefaultNonceTimeout,
	}
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
	if len(secret) == 0 {
		secret = util.Crypto.MustCreateKey(32)
	}
	return &Manager{
		secret:               secret,
		skipDomainValidation: cfg.GetSkipDomainValidation(),
		redirectURI:          cfg.GetRedirectURI(),
		validDomains:         cfg.GetValidDomains(),
		clientID:             cfg.GetClientID(),
		clientSecret:         cfg.GetClientSecret(),
		hostedDomain:         cfg.GetHostedDomain(),
		nonceTimeout:         cfg.GetNonceTimeout(),
	}, nil
}

// Manager is the oauth manager.
type Manager struct {
	secret               []byte
	audience             string
	redirectURI          string
	skipDomainValidation bool
	hostedDomain         string
	validDomains         []string
	clientID             string
	clientSecret         string
	nonceTimeout         time.Duration
}

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

// WithSkipDomainValidation sets if we should skip domain validation.
// It defaults to false, meaning we must supply valid domains.
func (m *Manager) WithSkipDomainValidation(value bool) *Manager {
	m.skipDomainValidation = value
	return m
}

// SkipDomainValidation returns a property.
func (m *Manager) SkipDomainValidation() bool {
	return m.skipDomainValidation
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

// WithValidDomains sets the valid domains.
// If values are not prefixed with `@`, they will be prefixed when testing email domains.
func (m *Manager) WithValidDomains(validDomains ...string) *Manager {
	m.validDomains = validDomains
	return m
}

// ValidDomains returns all valid domains. This includes the hosted domain in configured.
func (m *Manager) ValidDomains() []string {
	all := map[string]bool{}
	for _, domain := range m.validDomains {
		all[domain] = true
	}
	if len(m.hostedDomain) > 0 {
		all[m.hostedDomain] = true
	}
	var final []string
	for domain := range all {
		final = append(final, domain)
	}
	return final
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

// WithNonceTimeout sets the nonce timeout.
func (m *Manager) WithNonceTimeout(timeout time.Duration) *Manager {
	m.nonceTimeout = timeout
	return m
}

// NonceTimeout returns the nonce timeout.
func (m *Manager) NonceTimeout() time.Duration {
	return m.nonceTimeout
}

// OAuthURL is the auth url for google with a given clientID.
// This is typically the link that a user will click on to start the auth process.
func (m *Manager) OAuthURL(redirect ...string) (string, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   "accounts.google.com",
		Path:   "/o/oauth2/auth",
	}

	query := &url.Values{}
	query.Add("response_type", "code")
	query.Add("client_id", m.clientID)
	query.Add("scope", "openid email profile")
	query.Add("redirect_uri", m.redirectURI)

	state, err := m.State(redirect...)
	if err != nil {
		return "", err
	}
	if len(state) > 0 {
		query.Add("state", state)
	}

	nonce, err := m.CreateNonce(time.Now().UTC())
	if err != nil {
		return "", err
	}
	query.Add("nonce", nonce)

	if len(m.hostedDomain) > 0 {
		query.Add("hd", m.hostedDomain)
	}

	u.ForceQuery = true
	u.RawQuery = query.Encode()
	return u.String(), nil
}

// Finish processes the returned code, exchanging for an access token, and fetches the user profile.
func (m *Manager) Finish(r *http.Request) (*Result, error) {
	result := Result{}

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

	// if we need to check the anti-forgery token
	if result.State == nil || len(result.State.Secure) == 0 {
		return nil, ErrStateMissing
	}

	err := m.ValidateOAuthState(result.State)
	if err != nil {
		return nil, err
	}

	// exchange the code for an access token.
	res, err := m.TokenExchange(code)
	if err != nil {
		return &result, err
	}

	result.Response = res
	jwt, err := DeserializeJWT(res.IDToken)
	if err != nil {
		return &result, err
	}

	if err := m.ValidateJWT(jwt); err != nil {
		return &result, err
	}

	// fetch the user profile
	profile, err := m.FetchProfile(res.AccessToken)
	if err != nil {
		return &result, err
	}

	// validate the profile
	if err := m.ValidateProfile(profile); err != nil {
		return nil, err
	}

	// this is (supposed to be) the actual unique id for the user.
	result.UniqueID = jwt.Payload.Sub
	result.IDToken = &jwt.Payload
	result.Profile = profile

	return &result, nil
}

// State returns the serialized auth state.
func (m *Manager) State(redirect ...string) (string, error) {
	if len(m.secret) == 0 {
		return "", ErrSecretRequired
	}
	return SerializeOAuthState(m.CreateState(redirect...))
}

// CreateState creates auth state.
func (m *Manager) CreateState(redirect ...string) *State {
	token, secure := m.CreateAntiForgeryTokenPair()
	state := &State{
		Token:  token,
		Secure: secure,
	}

	if len(redirect) > 0 && len(redirect[0]) > 0 {
		if state == nil {
			state = &State{}
		}
		state.RedirectURL = redirect[0]
	}

	return state
}

// TokenExchange performs the second phase of the oauth 2.0 flow with google.
func (m *Manager) TokenExchange(code string) (*Response, error) {
	var oar Response
	meta, err := request.New().
		AsPost().
		WithScheme("https").
		WithHost("accounts.google.com").
		WithPath("o/oauth2/token").
		WithPostData("client_id", m.clientID).
		WithPostData("client_secret", m.clientSecret).
		WithPostData("grant_type", "authorization_code").
		WithPostData("redirect_uri", m.redirectURI).
		WithPostData("code", code).
		WithMockProvider(request.MockedResponseInjector).
		JSONWithMeta(&oar)

	if err != nil {
		return nil, err
	}
	if meta.StatusCode > 299 {
		return nil, exception.NewFromErr(ErrGoogleResponseStatus).WithMessagef("status code; %d", meta.StatusCode)
	}
	return &oar, err
}

// FetchProfile gets a google proflile for an access token.
func (m *Manager) FetchProfile(accessToken string) (*Profile, error) {
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
		return nil, exception.NewFromErr(ErrGoogleResponseStatus).WithMessagef("status code; %d", meta.StatusCode)
	}
	return &profile, err
}

// ValidateConfig validates the manager configuration.
// This should be used on start to ensure that the manager has everything it needs.
func (m *Manager) ValidateConfig() error {
	if len(m.secret) == 0 {
		return ErrSecretRequired
	}
	if len(m.clientID) == 0 {
		return ErrClientIDRequired
	}
	if len(m.clientSecret) == 0 {
		return ErrClientSecretRequired
	}
	if len(m.redirectURI) == 0 {
		return ErrRedirectURIRequired
	}

	u, err := url.Parse(m.redirectURI)
	if err != nil {
		return ErrInvalidRedirectURI
	}

	if len(u.Scheme) == 0 {
		return ErrInvalidRedirectURI
	}

	if len(u.Host) == 0 {
		return ErrInvalidRedirectURI
	}

	return nil
}

// ValidateJWT validates a jwt.
func (m *Manager) ValidateJWT(jwt *JWT) error {
	if len(m.clientID) > 0 && !hmac.Equal([]byte(jwt.Payload.AUD), []byte(m.clientID)) {
		return ErrInvalidAUD
	}

	if m.skipDomainValidation {
		return nil
	}

	if len(jwt.Payload.HostedDomain) == 0 {
		return nil
	}

	validDomains := m.ValidDomains()
	if len(validDomains) == 0 {
		return ErrNoValidDomains
	}

	valid := false
	for _, domain := range m.ValidDomains() {
		valid = valid || util.String.CaseInsensitiveEquals(domain, jwt.Payload.HostedDomain)
		if valid {
			break
		}
	}

	if !valid {
		return ErrInvalidHostedDomain
	}

	// check the nonce ...
	return m.ValidateNonce(jwt.Payload.Nonce)
}

// ValidateProfile validates a profile.
func (m *Manager) ValidateProfile(p *Profile) error {
	if m.skipDomainValidation {
		return nil
	}

	if len(m.ValidDomains()) == 0 {
		return ErrNoValidDomains
	}

	valid := false
	for _, domain := range m.ValidDomains() {
		workingDomain := domain
		if !strings.HasPrefix(workingDomain, "@") {
			workingDomain = fmt.Sprintf("@%s", workingDomain)
		}
		valid = valid || util.String.HasSuffixCaseInsensitive(p.Email, workingDomain)
		if valid {
			break
		}
	}

	if !valid {
		return ErrInvalidEmailDomain
	}
	return nil
}

// ValidateOAuthState validates oauth state.
func (m *Manager) ValidateOAuthState(s *State) error {
	expected := m.hash(s.Token)
	actual := s.Secure
	if !hmac.Equal([]byte(expected), []byte(actual)) {
		return ErrInvalidAntiforgeryToken
	}
	return nil
}

// CreateAntiForgeryTokenPair generates an anti-forgery token.
func (m *Manager) CreateAntiForgeryTokenPair() (plaintext, ciphertext string) {
	plaintext = uuid.V4().String()
	ciphertext = m.hash(plaintext)
	return
}

// CreateNonce creates a nonce.
func (m *Manager) CreateNonce(t time.Time) (string, error) {
	tv := t.Format(time.RFC3339)
	cipherText, err := util.Crypto.Encrypt(m.secret, []byte(tv))
	if err != nil {
		return "", err
	}
	return url.QueryEscape(Base64Encode([]byte(cipherText))), nil
}

// ValidateNonce validates a nonce.
func (m *Manager) ValidateNonce(nonce string) error {
	unescaped, err := url.QueryUnescape(nonce)
	if err != nil {
		return exception.Wrap(err)
	}
	ciphertext, err := Base64Decode(unescaped)
	if err != nil {
		return exception.Wrap(err)
	}
	plaintext, err := util.Crypto.Decrypt(m.secret, ciphertext)
	if err != nil {
		return exception.Wrap(err)
	}
	nonceTimestamp, err := time.Parse(time.RFC3339, string(plaintext))
	if err != nil {
		return exception.Wrap(err)
	}

	if time.Now().UTC().Sub(nonceTimestamp) > m.NonceTimeout() {
		return ErrInvalidNonce
	}
	return nil
}

func (m *Manager) hash(plaintext string) string {
	return Base64Encode(util.Crypto.Hash(m.secret, []byte(plaintext)))
}
