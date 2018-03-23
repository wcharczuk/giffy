package spiffy

import (
	"testing"

	assert "github.com/blendlabs/go-assert"
)

func TestUUID_v4(t *testing.T) {
	m := make(map[string]bool)
	for x := 1; x < 32; x++ {
		uuid := UUIDv4()
		s := uuid.ToFullString()
		if m[s] {
			t.Errorf("NewRandom returned duplicated UUID %s\n", s)
		}
		m[s] = true
		if v := uuid.Version(); v != 4 {
			t.Errorf("Random UUID of version %v\n", v)
		}
	}
}

func TestParamTokens(t *testing.T) {
	assert := assert.New(t)

	assert.Empty("", ParamTokens(0, 0))
	assert.Empty("", ParamTokens(1, 0))
	assert.Equal("$0", ParamTokens(0, 1))
	assert.Equal("$1", ParamTokens(1, 1))
	assert.Equal("$1,$2,$3", ParamTokens(1, 3))
	assert.Equal("$3,$4,$5,$6", ParamTokens(3, 4))
}
