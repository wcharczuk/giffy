package env

import (
	"os"
	"strings"
	"sync"

	util "github.com/blendlabs/go-util"
)

var (
	_env     Vars
	_envLock = sync.Mutex{}
)

const (
	// VarServiceEnv is a common env var name.
	VarServiceEnv = "SERVICE_ENV"
	// VarServiceName is a common env var name.
	VarServiceName = "SERVICE_NAME"
	// VarServiceSecret is a common env var name.
	VarServiceSecret = "SERVICE_SECRET"
	// VarPort is a common env var name.
	VarPort = "PORT"
	// VarSecurePort is a common env var name.
	VarSecurePort = "SECURE_PORT"
	// VarTLSCertPath is a common env var name.
	VarTLSCertPath = "TLS_CERT_PATH"
	// VarTLSKeyPath is a common env var name.
	VarTLSKeyPath = "TLS_KEY_PATH"
	// VarTLSCert is a common env var name.
	VarTLSCert = "TLS_CERT"
	// VarTLSKey is a common env var name.
	VarTLSKey = "TLS_KEY"

	// VarPGIdleConns is a common env var name.
	VarPGIdleConns = "PG_IDLE_CONNS"
	// VarPGMaxConns is a common env var name.
	VarPGMaxConns = "PG_MAX_CONNS"

	// ServiceEnvDev is a service environment.
	ServiceEnvDev = "dev"
	// ServiceEnvCI is a service environment.
	ServiceEnvCI = "ci"
	// ServiceEnvPreprod is a service environment.
	ServiceEnvPreprod = "preprod"
	// ServiceEnvBeta is a service environment.
	ServiceEnvBeta = "beta"
	// ServiceEnvProd is a service environment.
	ServiceEnvProd = "prod"
)

// Env returns the current env var set.
func Env() Vars {
	if _env == nil {
		_envLock.Lock()
		defer _envLock.Unlock()
		if _env == nil {
			_env = NewVarsFromEnvironment()
		}
	}
	return _env
}

// SetEnv sets the env vars.
func SetEnv(vars Vars) {
	_envLock.Lock()
	_env = vars
	_envLock.Unlock()
}

// NewVars returns a new env var set.
func NewVars() Vars {
	return Vars{}
}

// NewVarsFromEnvironment reads an EnvVar set from the environment.
func NewVarsFromEnvironment() Vars {
	vars := Vars{}
	envVars := os.Environ()
	for _, ev := range envVars {
		parts := strings.SplitN(ev, "=", 2)
		if len(parts) > 1 {
			vars[parts[0]] = parts[1]
		}
	}
	return vars
}

// Vars is a set of environment variables.
type Vars map[string]string

// Set sets a value for a key.
func (ev Vars) Set(key, value string) {
	ev[key] = value
}

// Restore resets an environment variable to it's environment value.
func (ev Vars) Restore(key string) {
	ev[key] = os.Getenv(key)
}

// String returns a string value for a given key.
func (ev Vars) String(key string, defaults ...string) string {
	if value, hasValue := ev[key]; hasValue {
		return value
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return ""
}

// Bool returns a boolean value for a key, defaulting to false.
// Valid "truthy" values are `true`, `yes`, and `1`.
// Everything else is false, including `REEEEEEEEEEEEEEE`.
func (ev Vars) Bool(key string, defaults ...bool) bool {
	if value, hasValue := ev[key]; hasValue {
		if len(value) > 0 {
			return util.String.CaseInsensitiveEquals(value, "true") || util.String.CaseInsensitiveEquals(value, "yes") || value == "1"
		}
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return false
}

// Int returns an integer value for a given key.
func (ev Vars) Int(key string, defaults ...int) int {
	if value, hasValue := ev[key]; hasValue {
		return util.String.ParseInt(value)
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return 0
}

// Int64 returns an int64 value for a given key.
func (ev Vars) Int64(key string, defaults ...int64) int64 {
	if value, hasValue := ev[key]; hasValue {
		return util.String.ParseInt64(value)
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return 0
}

// Bytes returns a []byte value for a given key.
func (ev Vars) Bytes(key string, defaults ...[]byte) []byte {
	if value, hasValue := ev[key]; hasValue && len(value) > 0 {
		return []byte(value)
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return nil
}

// Base64 returns a []byte value for a given key whose value is encoded in base64.
func (ev Vars) Base64(key string, defaults ...[]byte) []byte {
	if value, hasValue := ev[key]; hasValue && len(value) > 0 {
		result, _ := util.String.Base64Decode(value)
		return result
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return nil
}

// HasKey returns if a key is present in the set.
func (ev Vars) HasKey(key string) bool {
	_, hasKey := ev[key]
	return hasKey
}

// HasKeys returns if all of the given keys are present in the set.
func (ev Vars) HasKeys(key ...string) bool {
	for _, key := range key {
		if !ev.HasKey(key) {
			return false
		}
	}
	return true
}

// Union returns the union of the two sets, other replacing conflicts.
func (ev Vars) Union(other Vars) Vars {
	newSet := NewVars()
	for key, value := range ev {
		newSet[key] = value
	}
	for key, value := range other {
		newSet[key] = value
	}
	return newSet
}

// Keys returns all the keys stored in the env var set.
func (ev Vars) Keys() []string {
	var keys = make([]string, len(ev))
	var index int
	for key := range ev {
		keys[index] = key
		index++
	}
	return keys
}

// ServiceEnv is a common environment variable for the services environment.
// Common values include "dev", "ci", "sandbox", "preprod", "beta", and "prod".
func (ev Vars) ServiceEnv(defaults ...string) string {
	return ev.String(VarServiceEnv, defaults...)
}

// IsProduction returns if the ServiceEnv is a production environment.
func (ev Vars) IsProduction() bool {
	return ev.ServiceEnv() == ServiceEnvPreprod ||
		ev.ServiceEnv() == ServiceEnvProd
}

// IsProdLike returns if the ServiceEnv is "prodlike".
func (ev Vars) IsProdLike() bool {
	return ev.ServiceEnv() == ServiceEnvPreprod ||
		ev.ServiceEnv() == ServiceEnvBeta ||
		ev.ServiceEnv() == ServiceEnvProd
}

// ServiceName is a common environment variable for the service's name.
func (ev Vars) ServiceName(defaults ...string) string {
	return ev.String(VarServiceName, defaults...)
}

// ServiceSecret is the main secret for the app.
// It is typically a 32 byte / 256 bit key.
func (ev Vars) ServiceSecret(defaults ...[]byte) []byte {
	return ev.Base64(VarServiceSecret, defaults...)
}

// Port is a common environment variable.
// It is what TCP port to bind to for the HTTP server.
func (ev Vars) Port(defaults ...string) string {
	return ev.String(VarPort, defaults...)
}

// SecurePort is a common environment variable.
// It is what TCP port to bind to for the HTTPS server.
func (ev Vars) SecurePort(defaults ...string) string {
	return ev.String(VarSecurePort, defaults...)
}

// TLSCertFilepath is a common environment variable for the (whole) TLS cert to use with https.
func (ev Vars) TLSCertFilepath(defaults ...string) string {
	return ev.String(VarTLSCertPath, defaults...)
}

// TLSKeyFilepath is a common environment variable for the (whole) TLS key to use with https.
func (ev Vars) TLSKeyFilepath(defaults ...string) string {
	return ev.String(VarTLSKeyPath, defaults...)
}

// TLSCert is a common environment variable for the (whole) TLS cert to use with https.
func (ev Vars) TLSCert(defaults ...[]byte) []byte {
	return ev.Bytes(VarTLSCert, defaults...)
}

// TLSKey is a common environment variable for the (whole) TLS key to use with https.
func (ev Vars) TLSKey(defaults ...[]byte) []byte {
	return ev.Bytes(VarTLSKey, defaults...)
}
