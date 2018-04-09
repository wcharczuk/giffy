package oauth

const (
	// ErrCodeMissing is returned if the code was missing from an oauth return request.
	ErrCodeMissing Error = "state missing from request"
	// ErrStateMissing is returned if the state was missing from an oauth return request.
	ErrStateMissing Error = "state missing from request"
	// ErrNoValidDomains is an error that occurs during profile validation if the manager
	// doesn't have any valid domains configured.
	ErrNoValidDomains Error = "domain validation enabled and no valid domains provided"
	// ErrInvalidEmailDomain is an error that occurs during profile validation if the
	// profile email doesn't match the valid domain list.
	ErrInvalidEmailDomain Error = "domain validation failed; doesn't match valid domain list"
	// ErrInvalidHostedDomain is an error returned if the JWT hosted zone doesn't match any of the whitelisted domains.
	ErrInvalidHostedDomain Error = "hosted domain validation failed"
	// ErrInvalidAUD is an error returned during jwt validation.
	ErrInvalidAUD Error = "invalid JWT AUD"

	// ErrInvalidAntiforgeryToken is an error that occurs during oauth finish if the forgery token is required
	// and missing or invalid.
	ErrInvalidAntiforgeryToken Error = "invalid anti-forgery token"
	// ErrGoogleResponseStatus is an error that can occur when querying the google apis.
	ErrGoogleResponseStatus Error = "google returned a non 2xx response"

	// ErrSecretRequired is a configuration error indicating we did not provide a secret.
	ErrSecretRequired Error = "manager secret required"
	// ErrClientIDRequired is a self validation error.
	ErrClientIDRequired Error = "clientID is required"
	// ErrClientSecretRequired is a self validation error.
	ErrClientSecretRequired Error = "clientSecret is required"
	// ErrRedirectURIRequired is a self validation error.
	ErrRedirectURIRequired Error = "redirectURI is required"
	// ErrInvalidRedirectURI is an error in validating the redirect uri.
	ErrInvalidRedirectURI Error = "invalid redirectURI"
	// ErrInvalidJWT is an error that occurs when deserializing the jwt.
	ErrInvalidJWT Error = "invalid jwt"
	// ErrInvalidNonce is an error that occurs when checking the jwt nonce.
	ErrInvalidNonce Error = "invalid nonce"
)

// Error is an error string.
type Error string

// Error implements error.
func (e Error) Error() string {
	return string(e)
}
