package model

import (
	"net/http"
	"net/url"
	"testing"

	assert "github.com/blend/go-sdk/assert"
	exception "github.com/blend/go-sdk/exception"
)

func TestGetAllErrorsWithLimitAndOffset(t *testing.T) {
	assert := assert.New(t)
	tx, err := DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	err = DB().CreateInTx(NewError(
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
	), tx)
	assert.Nil(err)

	err = DB().CreateInTx(NewError(
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
	), tx)
	assert.Nil(err)

	errors, err := GetAllErrorsWithLimitAndOffset(1, 0, tx)
	assert.Nil(err)
	assert.Len(1, errors)
}
