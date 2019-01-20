package model

import (
	"net/http"
	"net/url"
	"testing"

	assert "github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/db"
	exception "github.com/blend/go-sdk/exception"
)

func TestGetAllErrorsWithLimitAndOffset(t *testing.T) {
	assert := assert.New(t)
	todo := testCtx()
	tx, err := db.Default().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := Manager{DB: db.Default(), Tx: tx}

	err = m.Invoke(todo).Create(NewError(
		exception.New("This is only a test"),
		&http.Request{
			Method: "GET",
			Proto:  "http",
			Host:   "localhost",
			URL: &url.URL{
				Path:     "/foo",
				RawQuery: "foo=bar",
			},
		},
	))
	assert.Nil(err)

	err = m.Invoke(todo).Create(NewError(
		exception.New("This is only a test"),
		&http.Request{
			Method: "GET",
			Proto:  "http",
			Host:   "localhost",
			URL: &url.URL{
				Path:     "/foo",
				RawQuery: "foo=bar",
			},
		},
	))
	assert.Nil(err)

	errors, err := m.GetAllErrorsWithLimitAndOffset(todo, 1, 0)
	assert.Nil(err)
	assert.Len(errors, 1)
}
