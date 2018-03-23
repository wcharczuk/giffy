package spiffy

import (
	"fmt"
	"testing"

	assert "github.com/blendlabs/go-assert"
)

func TestInvocationErr(t *testing.T) {
	assert := assert.New(t)

	inv := &Invocation{err: fmt.Errorf("this is only a test")}
	assert.NotNil(inv.Err())
}

func TestInvocationLabels(t *testing.T) {
	assert := assert.New(t)

	inv := &Invocation{}
	inv = inv.WithLabel("test")
	assert.NotEmpty(inv.Label())
}

func TestInvocationPrepare(t *testing.T) {
	assert := assert.New(t)

	inv := &Invocation{err: fmt.Errorf("test")}
	_, err := inv.Prepare("select 'ok!'")
	assert.NotNil(err)
}
