package model

import (
	"database/sql"
	"time"

	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/giffy/server/core"
)

// User is a user in the app.
type User struct {
	ID         int64     `json:"-" db:"id,pk,serial"`
	UUID       string    `json:"uuid" db:"uuid"`
	CreatedUTC time.Time `json:"created_utc" db:"created_utc"`
	Username   string    `json:"username" db:"username"`
	FirstName  string    `json:"first_name" db:"first_name"`
	LastName   string    `json:"last_name" db:"last_name"`

	EmailAddress    string `json:"email_address" db:"email_address"`
	IsEmailVerified bool   `json:"is_email_verified" db:"is_email_verified"`
}

// TableName is the table name.
func (u User) TableName() string {
	return "users"
}

// IsZero returns if the user is set or not.
func (u User) IsZero() bool {
	return u.ID == 0
}

// NewUser returns a new user.
func NewUser(username string) *User {
	return &User{
		UUID:            core.UUIDv4().ToShortString(),
		CreatedUTC:      time.Now().UTC(),
		Username:        username,
		IsEmailVerified: false,
	}
}

// GetAllUsers returns all the users.
func GetAllUsers(tx *sql.Tx) ([]User, error) {
	var all []User
	err := spiffy.DefaultDb().GetAllInTransaction(&all, tx)
	return all, err
}

// GetUserByID returns a user by id.
func GetUserByID(id int, tx *sql.Tx) (*User, error) {
	var user User
	err := spiffy.DefaultDb().GetByIDInTransaction(&user, tx, id)
	return &user, err
}

// GetUserByUUID returns a user for a uuid.
func GetUserByUUID(uuid string, tx *sql.Tx) (*User, error) {
	var user User
	err := spiffy.DefaultDb().
		QueryInTransaction(`select * from users where uuid = $1`, tx, uuid).Out(&user)
	return &user, err
}

// GetUserByUsername returns a user for a uuid.
func GetUserByUsername(username string, tx *sql.Tx) (*User, error) {
	var user User
	err := spiffy.DefaultDb().
		QueryInTransaction(`select * from users where username = $1`, tx, username).Out(&user)
	return &user, err
}
