package slack

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestParseTimestamp(t *testing.T) {
	a := assert.New(t)

	unix := ParseTimestamp("1356032811")
	a.NotNil(unix)
	a.Equal(2012, unix.DateTime().Year())
	combo := ParseTimestamp("1355517523.000005")
	a.NotNil(combo)
	a.Equal(2012, combo.DateTime().Year())
}
