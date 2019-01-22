package controller

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/web"
	"github.com/wcharczuk/giffy/server/model"
)

// GetUser returns the user state from a session.
func GetUser(session *web.Session) *model.User {
	if session == nil {
		return nil
	}
	if userData, hasUser := session.State[SessionStateUserKey]; hasUser {
		if user, isUser := userData.(*model.User); isUser {
			return user
		}
	}
	return nil
}

// SetUser stores a user on a session.
func SetUser(session *web.Session, user *model.User) {
	if session.State == nil {
		session.State = map[string]interface{}{}
	}
	session.State[SessionStateUserKey] = user
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func parseInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

func toJSON(v interface{}) string {
	buf := bytes.NewBuffer(nil)
	json.NewEncoder(buf).Encode(v)
	return buf.String()
}

func fromJSON(corpus []byte, output interface{}) error {
	return exception.New(json.Unmarshal(corpus, output))
}
