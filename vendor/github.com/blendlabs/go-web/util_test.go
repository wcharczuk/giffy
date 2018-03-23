package web

import (
	"testing"

	assert "github.com/blendlabs/go-assert"
)

func TestNestMiddleware(t *testing.T) {
	assert := assert.New(t)

	var callIndex int

	var mw1Called int
	mw1 := func(action Action) Action {
		return func(ctx *Ctx) Result {
			mw1Called = callIndex
			callIndex = callIndex + 1
			return action(ctx)
		}
	}

	var mw2Called int
	mw2 := func(action Action) Action {
		return func(ctx *Ctx) Result {
			mw2Called = callIndex
			callIndex = callIndex + 1
			return action(ctx)
		}
	}

	var mw3Called int
	mw3 := func(action Action) Action {
		return func(ctx *Ctx) Result {
			mw3Called = callIndex
			callIndex = callIndex + 1
			return action(ctx)
		}
	}

	nested := NestMiddleware(func(ctx *Ctx) Result { return nil }, mw2, mw3, mw1)

	nested(nil)

	assert.Equal(2, mw2Called)
	assert.Equal(1, mw3Called)
	assert.Equal(0, mw1Called)
}

func TestPortFromBindAddr(t *testing.T) {
	assert := assert.New(t)

	testCases := map[string]int32{
		"":               0,
		"2":              2,
		":2":             2,
		"127.0.0.1:1234": 1234,
		":8080":          8080,
	}

	for input, expected := range testCases {
		assert.Equal(expected, PortFromBindAddr(input))
	}
}
