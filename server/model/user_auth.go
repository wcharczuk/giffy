package model

import (
	"database/sql"
	"time"

	"github.com/blendlabs/go-exception"
	"github.com/wcharczuk/giffy/server/core"
)

// UserAuth is what we use to store auth credentials.
type UserAuth struct {
	UserID        int64     `json:"user_id" db:"user_id,pk"`
	TimestampUTC  time.Time `json:"timestamp_utc" db:"timestamp_utc"`
	Provider      string    `json:"provider" db:"provider,pk"`
	AuthToken     []byte    `json:"auth_token" db:"auth_token"`
	AuthTokenHash []byte    `json:"auth_token_hash" db:"auth_token_hash"`
	AuthSecret    []byte    `json:"auth_secret" db:"auth_secret"`
}

// TableName returns the table name.
func (ua UserAuth) TableName() string {
	return "user_auth"
}

// IsZero returns if the object has been set or not.
func (ua UserAuth) IsZero() bool {
	return ua.UserID == 0
}

// NewUserAuth returns a new user auth entry, encrypting the authToken and authSecret.
func NewUserAuth(userID int64, authToken, authSecret string) (*UserAuth, error) {
	auth := &UserAuth{
		UserID:       userID,
		TimestampUTC: time.Now().UTC(),
	}

	key := core.ConfigKey()
	token, err := core.Encrypt(key, authToken)
	if err != nil {
		return auth, err
	}
	auth.AuthToken = token
	auth.AuthTokenHash = core.Hash(key, authToken)

	if len(authSecret) != 0 {
		secret, err := core.Encrypt(key, authSecret)
		if err != nil {
			return auth, err
		}
		auth.AuthSecret = secret
	}

	return auth, nil
}

// GetUserAuthByToken returns an auth entry for the given auth token.
func GetUserAuthByToken(token string, tx *sql.Tx) (*UserAuth, error) {
	if len(core.ConfigKey()) == 0 {
		return nil, exception.New("`ENCRYPTION_KEY` is not set, cannot continue.")
	}

	key := core.ConfigKey()
	authTokenHash := core.Hash(key, token)

	var auth UserAuth
	err := DB().QueryInTx("SELECT * FROM user_auth where auth_token_hash = $1", tx, authTokenHash).Out(&auth)
	return &auth, err
}

// GetUserAuthByProvider returns an auth entry for the given auth token.
func GetUserAuthByProvider(userID int64, provider string, tx *sql.Tx) (*UserAuth, error) {
	var auth UserAuth
	err := DB().QueryInTx("SELECT * FROM user_auth where user_id = $1 and provider = $2", tx, userID, provider).Out(&auth)
	return &auth, err
}

// DeleteUserAuthForProvider deletes auth entries for a provider for a user.
func DeleteUserAuthForProvider(userID int64, provider string, tx *sql.Tx) error {
	return DB().ExecInTx("DELETE FROM user_auth where user_id = $1 and provider = $2", tx, userID, provider)
}
