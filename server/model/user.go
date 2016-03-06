package model

import (
	"database/sql"
	"time"

	"github.com/blendlabs/spiffy"
	"github.com/wcharczuk/go-slack"
)

type User struct {
	ID         int64     `json:"-" db:"id"`
	UUID       string    `json:"uuid" db:"uuid"`
	CreatedUTC time.Time `json:"created_utc" db:"created_utc"`
	Username   string    `json:"username" db:"username"`
	FirstName  string    `json:"first_name" db:"first_name"`
	LastName   string    `json:"last_name" db:"last_name"`
}

func (u User) TableName() string {
	return "users"
}

func NewUser() *User {
	return &User{
		UUID:       slack.UUIDv4().ToShortString(),
		CreatedUTC: time.Now().UTC(),
	}
}

func GetAllUsers(tx *sql.Tx) ([]User, error) {
	var all []User
	err := spiffy.DefaultDb().GetAllInTransaction(&all, tx)
	return all, err
}

func GetUserByID(id int, tx *sql.Tx) (*User, error) {
	var user User
	err := spiffy.DefaultDb().GetByIdInTransaction(&user, tx, id)
	return &user, err
}

func GetUserByUUID(uuid string, tx *sql.Tx) (*User, error) {
	var user User
	err := spiffy.DefaultDb().
		QueryInTransaction(`select * from users where uuid = $1`, tx, uuid).Out(&user)
	return &user, err
}
