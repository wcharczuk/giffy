package logger

import (
	"net/http"
	"testing"

	assert "github.com/blendlabs/go-assert"
)

func TestGetIP(t *testing.T) {
	assert := assert.New(t)

	hdr := http.Header{}
	hdr.Set("X-Forwarded-For", "123")
	r := http.Request{
		Header: hdr,
	}
	assert.Equal("123", GetIP(&r))

	hdr = http.Header{}
	hdr.Set("X-FORWARDED-FOR", "1")
	r = http.Request{
		Header: hdr,
	}
	assert.Equal("1", GetIP(&r))

	hdr = http.Header{}
	hdr.Set("X-FORWARDED-FOR", "1,2,3")
	r = http.Request{
		Header: hdr,
	}
	assert.Equal("1", GetIP(&r))

	hdr = http.Header{}
	hdr.Set("X-Real-Ip", "1")
	r = http.Request{
		Header: hdr,
	}
	assert.Equal("1", GetIP(&r))

	hdr = http.Header{}
	hdr.Set("X-REAL-IP", "1")
	r = http.Request{
		Header: hdr,
	}
	assert.Equal("1", GetIP(&r))

	hdr = http.Header{}
	hdr.Set("X-REAL-IP", "1,2,3")
	r = http.Request{
		Header: hdr,
	}
	assert.Equal("1", GetIP(&r))

	r = http.Request{
		RemoteAddr: "1:1",
	}
	assert.Equal("1", GetIP(&r))

	r = http.Request{
		RemoteAddr: "1",
	}
	assert.Equal("", GetIP(&r))
}
