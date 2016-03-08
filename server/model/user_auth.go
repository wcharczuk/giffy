package model

import (
	"database/sql"
	"time"

	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
)

// UserAuth is what we use to store auth credentials.
type UserAuth struct {
	UserID       int64     `json:"user_id" db:"user_id,pk"`
	TimestampUTC time.Time `json:"timestamp_utc" db:"timestamp_utc"`
	Provider     string    `json:"provider" db:"provider"`
	AuthToken    []byte    `json:"auth_token"`
	AuthSecret   []byte    `json:"auth_secret"`
}

// TableName returns the table name.
func (ua UserAuth) TableName() string {
	return "user_auth"
}

// NewUserAuth returns a new user auth entry.
func NewUserAuth(userID int64, authToken, authSecret string) *UserAuth {
	auth := &UserAuth{
		UserID:       userID,
		TimestampUTC: time.Now().UTC(),
	}

	token, tokenErr := core.Encrypt(core.ConfigKey(), authToken)
	if tokenErr != nil {
		return auth
	}
	auth.AuthToken = token

	if len(authSecret) != 0 {
		secret, secretErr := core.Encrypt(core.ConfigKey(), authSecret)
		if secretErr != nil {
			return auth
		}
		auth.AuthSecret = secret
	}

	return auth
}

// GetUserAuthByCredentials returns an auth entry for the given credentials.
func GetUserAuthByCredentials(authToken, authSecret []byte, tx *sql.Tx) (*UserAuth, error) {
	var auth UserAuth
	err := spiffy.DefaultDb().Query("SELECT * FROM user_auth where authToken = $1 and authSecret = $2", authToken, authSecret).Out(&auth)
	return &auth, err
}
