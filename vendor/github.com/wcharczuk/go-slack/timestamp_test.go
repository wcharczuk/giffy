package slack

import (
	"encoding/json"
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestTimestamp(t *testing.T) {
	a := assert.New(t)

	unixTS := &Timestamp{}
	unixTS.UnmarshalJSON([]byte("1356032811"))
	a.NotNil(unixTS)
	a.Equal(2012, unixTS.Time().Year())

	combo := &Timestamp{}
	combo.UnmarshalJSON([]byte("1456540321.000014"))
	a.NotNil(combo)
	a.Equal(2016, combo.Time().Year())
	a.Equal("000014", combo.UUID())

	comboAsString := combo.String()
	a.Equal("1456540321.000014", comboAsString)
}

func TestTimestampUnmarshal(t *testing.T) {
	a := assert.New(t)

	messageBody := `{"type":"message","user":"U0KMCE0MC","text":"this is a test.","ts":"1456540738.000017"}`

	var m Message
	err := json.Unmarshal([]byte(messageBody), &m)
	a.Nil(err)

	a.NotNil(m.Timestamp)
	a.Equal("1456540738.000017", m.Timestamp.String())
}
