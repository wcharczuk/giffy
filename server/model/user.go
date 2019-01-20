package model

import (
	"time"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/uuid"
)

// User is a user in the app.
type User struct {
	ID         int64     `json:"id" db:"id,pk,serial"`
	UUID       string    `json:"uuid" db:"uuid"`
	CreatedUTC time.Time `json:"created_utc" db:"created_utc"`
	Username   string    `json:"username" db:"username"`
	FirstName  string    `json:"first_name" db:"first_name"`
	LastName   string    `json:"last_name" db:"last_name"`

	EmailAddress    string `json:"email_address" db:"email_address"`
	IsEmailVerified bool   `json:"is_email_verified" db:"is_email_verified"`

	IsAdmin     bool `json:"is_admin" db:"is_admin"`
	IsModerator bool `json:"is_moderator" db:"is_moderator"`
	IsBanned    bool `json:"is_banned" db:"is_banned"`
}

// TableName is the table name.
func (u User) TableName() string {
	return "users"
}

// IsZero returns if the user is set or not.
func (u User) IsZero() bool {
	return u.ID == 0
}

// Populate scans the rows into the struct.
func (u *User) Populate(r db.Rows) error {
	return r.Scan(
		&u.ID,
		&u.UUID,
		&u.CreatedUTC,
		&u.Username,
		&u.FirstName,
		&u.LastName,
		&u.EmailAddress,
		&u.IsEmailVerified,
		&u.IsAdmin,
		&u.IsModerator,
		&u.IsBanned,
	)
}

// NewUser returns a new user.
func NewUser(username string) *User {
	return &User{
		UUID:            uuid.V4().String(),
		CreatedUTC:      time.Now().UTC(),
		Username:        username,
		IsEmailVerified: false,
	}
}
