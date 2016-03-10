package slack

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestUUIDv4(t *testing.T) {
	a := assert.New(t)

	sample := UUIDv4()
	a.Equal(byte(4), sample.Version())

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
