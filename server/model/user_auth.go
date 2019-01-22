package model

import (
	"time"

	"github.com/blend/go-sdk/crypto"
)

// NewUserAuth returns a new user auth entry, encrypting the authToken and authSecret.
func NewUserAuth(userID int64, authToken, authSecret string, key []byte) (*UserAuth, error) {
	auth := &UserAuth{
		UserID:       userID,
		TimestampUTC: time.Now().UTC(),
	}

	token, err := crypto.Encrypt(key, []byte(authToken))
	if err != nil {
		return auth, err
	}
	auth.AuthToken = token
	auth.AuthTokenHash = crypto.HMAC512(key, []byte(authToken))

	if len(authSecret) != 0 {
		secret, err := crypto.Encrypt(key, []byte(authSecret))
		if err != nil {
			return auth, err
		}
		auth.AuthSecret = secret
	}

	return auth, nil
}

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
