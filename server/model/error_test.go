package model

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/ex"
	"github.com/blend/go-sdk/testutil"
)

func TestGetAllErrorsWithLimitAndOffset(t *testing.T) {
	assert := assert.New(t)
	todo := context.TODO()
	tx, err := testutil.DefaultDB().Begin()
	assert.Nil(err)
	defer tx.Rollback()
	m := NewTestManager(tx)

	err = m.Invoke(todo).Create(NewError(
		ex.New("This is only a test"),
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
		ex.New("This is only a test"),
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
