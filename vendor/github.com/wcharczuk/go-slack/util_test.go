package slack

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestOptionals(t *testing.T) {
	a := assert.New(t)

	a.NotNil(OptionalBool(true))
	a.NotNil(OptionalByte(0))
	a.NotNil(OptionalFloat32(0))
	a.NotNil(OptionalFloat64(0))
	a.NotNil(OptionalInt(0))
	a.NotNil(OptionalInt16(0))
	a.NotNil(OptionalInt32(0))
	a.NotNil(OptionalInt64(0))
}
